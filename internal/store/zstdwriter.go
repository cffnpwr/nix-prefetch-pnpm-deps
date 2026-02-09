package store

/*
#cgo pkg-config: libzstd
#include <zstd.h>

// compressChunk wraps ZSTD_compressStream2 to avoid CGo enum type issues.
static size_t compressChunk(ZSTD_CCtx* cctx, ZSTD_outBuffer* output,
                            ZSTD_inBuffer* input, int endOp) {
	return ZSTD_compressStream2(cctx, output, input, (ZSTD_EndDirective)endOp);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"unsafe"
)

// zstdCLILevel is the default compression level used by the CLI zstd command
// and GNU tar's --zstd flag.
const zstdCLILevel = 3

const (
	zstdEContinue = 0 // ZSTD_e_continue
	zstdEEnd      = 2 // ZSTD_e_end
)

// zstdWriter compresses data using the C zstd library with parameters matching CLI zstd.
// It buffers input in ZSTD_CStreamInSize chunks and uses ZSTD_e_end on the final chunk
// in Close(), producing byte-identical output to the zstd CLI command.
type zstdWriter struct {
	w           io.Writer
	cctx        *C.ZSTD_CCtx
	inBuf       []byte // input buffer (ZSTD_CStreamInSize)
	inPos       int    // bytes buffered in inBuf
	outBuf      []byte // output buffer (ZSTD_CStreamOutSize)
	initialized bool   // whether initStream has been called
}

// newZstdWriter creates a zstd compressor matching CLI zstd defaults:
// compression level 3, content checksum enabled.
func newZstdWriter(w io.Writer) (*zstdWriter, error) {
	cctx := C.ZSTD_createCCtx()
	if cctx == nil {
		return nil, errors.New("failed to create zstd context")
	}

	C.ZSTD_CCtx_setParameter(cctx, C.ZSTD_c_compressionLevel, C.int(zstdCLILevel))
	C.ZSTD_CCtx_setParameter(cctx, C.ZSTD_c_checksumFlag, 1)
	// Disable content size in frame header to match zstd CLI streaming behavior.
	// When zstd reads from stdin/pipe, it doesn't know the total size upfront
	// and omits the content size field. We must do the same for byte-identical output.
	C.ZSTD_CCtx_setParameter(cctx, C.ZSTD_c_contentSizeFlag, 0)
	// Enable multi-threaded mode with one worker to match zstd CLI's streaming behavior.
	// The CLI uses multi-threaded mode (nbWorkers >= 1) for pipe/stdin input, which uses
	// a different compression code path than single-threaded (nbWorkers=0) and produces
	// different output. Setting nbWorkers=1 matches the CLI's output exactly.
	C.ZSTD_CCtx_setParameter(cctx, C.ZSTD_c_nbWorkers, 1)

	inBufSize := int(C.ZSTD_CStreamInSize())
	outBufSize := int(C.ZSTD_CStreamOutSize())

	return &zstdWriter{
		w:      w,
		cctx:   cctx,
		inBuf:  make([]byte, inBufSize),
		inPos:  0,
		outBuf: make([]byte, outBufSize),
	}, nil
}

func (z *zstdWriter) Write(p []byte) (int, error) {
	if err := z.initStream(); err != nil {
		return 0, err
	}

	total := 0

	for len(p) > 0 {
		// Fill input buffer
		n := copy(z.inBuf[z.inPos:], p)
		z.inPos += n
		p = p[n:]
		total += n

		// Flush when buffer is full
		if z.inPos == len(z.inBuf) {
			if err := z.flushInBuf(zstdEContinue); err != nil {
				return total, err
			}
		}
	}

	return total, nil
}

func (z *zstdWriter) Close() error {
	if z.cctx == nil {
		return nil
	}

	if err := z.initStream(); err != nil {
		C.ZSTD_freeCCtx(z.cctx)
		z.cctx = nil

		return err
	}

	// Flush remaining buffered data with ZSTD_e_end to finalize the frame.
	// This matches the zstd CLI behavior where the last chunk uses ZSTD_e_end
	// instead of a separate ZSTD_e_continue + empty ZSTD_e_end.
	if err := z.flushInBuf(zstdEEnd); err != nil {
		C.ZSTD_freeCCtx(z.cctx)
		z.cctx = nil

		return err
	}

	C.ZSTD_freeCCtx(z.cctx)
	z.cctx = nil

	return nil
}

// initStream sends an empty ZSTD_e_continue to initialize the compression stream.
// This locks the window size to the default (windowLog=21 at level 3) before any
// data is provided, preventing zstd from auto-adjusting the window to the input size
// when all data fits in a single ZSTD_e_end call. Without this, small inputs produce
// a different frame header than the zstd CLI in streaming (pipe/stdin) mode.
func (z *zstdWriter) initStream() error {
	if z.initialized {
		return nil
	}

	z.initialized = true

	var dummy [1]byte

	var pinner runtime.Pinner

	pinner.Pin(&z.outBuf[0])
	pinner.Pin(&dummy[0])

	defer pinner.Unpin()

	input := C.ZSTD_inBuffer{
		src:  unsafe.Pointer(&dummy[0]),
		size: 0,
		pos:  0,
	}

	_, err := z.compressAndWrite(&input, zstdEContinue)

	return err
}

// compressAndWrite runs one compression step and writes the output.
func (z *zstdWriter) compressAndWrite(
	input *C.ZSTD_inBuffer,
	endOp int,
) (C.size_t, error) {
	output := C.ZSTD_outBuffer{
		dst:  unsafe.Pointer(&z.outBuf[0]),
		size: C.size_t(len(z.outBuf)),
		pos:  0,
	}

	//nolint:gocritic // CGo-generated code triggers false positive dupSubExpr
	remaining := C.compressChunk(z.cctx, &output, input, C.int(endOp))
	if C.ZSTD_isError(remaining) != 0 {
		return 0, fmt.Errorf(
			"zstd compress: %s",
			C.GoString(C.ZSTD_getErrorName(remaining)),
		)
	}

	if output.pos > 0 {
		if _, err := z.w.Write(z.outBuf[:output.pos]); err != nil {
			return 0, err
		}
	}

	return remaining, nil
}

// flushInBuf compresses the current input buffer contents with the given end directive.
func (z *zstdWriter) flushInBuf(endOp int) error {
	if z.inPos == 0 && endOp == zstdEContinue {
		return nil
	}

	var pinner runtime.Pinner

	pinner.Pin(&z.outBuf[0])
	defer pinner.Unpin()

	var input C.ZSTD_inBuffer

	if z.inPos > 0 {
		pinner.Pin(&z.inBuf[0])

		input = C.ZSTD_inBuffer{
			src:  unsafe.Pointer(&z.inBuf[0]),
			size: C.size_t(z.inPos),
			pos:  0,
		}
	}

	for {
		remaining, err := z.compressAndWrite(&input, endOp)
		if err != nil {
			return err
		}

		if endOp == zstdEContinue {
			if input.pos >= input.size {
				break
			}
		} else {
			if remaining == 0 {
				break
			}
		}
	}

	z.inPos = 0

	return nil
}
