package store

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	tarBlockSize           = 512
	tarNameSize            = 100
	tarRecordSize          = 10240 // blocking factor 20 * 512
	tarEndOfArchiveBlocks  = 2     // two zero blocks mark end-of-archive
	tarPAXLenFallbackWidth = 11    // fallback length field width for extremely large PAX records
	tarPAXHdrMode          = 0o644
	tarTypePAX             = 'x'
	tarTypeFile            = '0'
	tarTypeDir             = '5'
	tarTypeSymlink         = '2'
)

// tarEntryInfo holds the metadata needed to write a tar entry.
type tarEntryInfo struct {
	mode       int64
	size       int64
	mtime      int64
	typeflag   byte
	linkname   string
	dataReader io.Reader
}

// gnuTarWriter writes tar archives compatible with GNU tar's PAX format output.
// Matches the output of:
//
//	tar --sort=name --mtime="@315532800" --owner=0 --group=0 --numeric-owner \
//	  --pax-option=exthdr.name=%d/PaxHeaders/%f,delete=atime,delete=ctime -cf -
type gnuTarWriter struct {
	w       io.Writer
	written int64
}

func newGNUTarWriter(w io.Writer) *gnuTarWriter {
	return &gnuTarWriter{w: w}
}

// writeBytes writes raw bytes and tracks the total written count.
func (tw *gnuTarWriter) writeBytes(p []byte) error {
	n, err := tw.w.Write(p)
	tw.written += int64(n)

	return err
}

// writeEntry writes a complete tar entry with PAX extended header if needed.
func (tw *gnuTarWriter) writeEntry(tarPath string, info tarEntryInfo) error {
	if len(tarPath) <= tarNameSize && len(info.linkname) <= tarNameSize {
		return tw.writeUSTAREntry(tarPath, info)
	}

	return tw.writePAXEntry(tarPath, info)
}

// writeUSTAREntry writes a single USTAR entry without PAX headers.
func (tw *gnuTarWriter) writeUSTAREntry(name string, info tarEntryInfo) error {
	if err := tw.writeHeaderBlock(tarHeader{
		name:     name,
		mode:     info.mode,
		size:     info.size,
		mtime:    info.mtime,
		typeflag: info.typeflag,
		linkname: info.linkname,
	}); err != nil {
		return err
	}

	return tw.writeFileData(info)
}

// writePAXEntry writes a PAX extended header followed by the actual USTAR entry.
func (tw *gnuTarWriter) writePAXEntry(tarPath string, info tarEntryInfo) error {
	paxData := buildPAXRecords(tarPath, info.linkname)
	paxDataBytes := []byte(paxData)

	// Write PAX extended header block
	paxName := buildPAXHeaderName(tarPath)
	if err := tw.writeHeaderBlock(tarHeader{
		name:     paxName,
		mode:     tarPAXHdrMode,
		size:     int64(len(paxDataBytes)),
		mtime:    info.mtime,
		typeflag: tarTypePAX,
	}); err != nil {
		return err
	}

	// Write PAX data + padding
	if err := tw.writeBytes(paxDataBytes); err != nil {
		return err
	}

	if err := tw.writeZeroPadding(len(paxDataBytes)); err != nil {
		return err
	}

	// Write actual file entry with truncated name/linkname
	linkname := info.linkname
	if len(linkname) > tarNameSize {
		linkname = linkname[:tarNameSize]
	}

	if err := tw.writeHeaderBlock(tarHeader{
		name:     truncateToNameSize(tarPath),
		mode:     info.mode,
		size:     info.size,
		mtime:    info.mtime,
		typeflag: info.typeflag,
		linkname: linkname,
	}); err != nil {
		return err
	}

	return tw.writeFileData(info)
}

// writeFileData writes file content followed by zero padding to block boundary.
func (tw *gnuTarWriter) writeFileData(info tarEntryInfo) error {
	if info.typeflag != tarTypeFile || info.size <= 0 || info.dataReader == nil {
		return nil
	}

	n, err := io.Copy(tw.w, info.dataReader)
	tw.written += n

	if err != nil {
		return err
	}

	return tw.writeZeroPadding(int(n))
}

// close writes the end-of-archive marker and pads to blocking factor boundary.
func (tw *gnuTarWriter) close() error {
	// Write two zero blocks (end-of-archive marker)
	zeros := make([]byte, tarBlockSize*tarEndOfArchiveBlocks)
	if err := tw.writeBytes(zeros); err != nil {
		return err
	}

	// Pad to blocking factor boundary (10240 bytes)
	remainder := int(tw.written % int64(tarRecordSize))
	if remainder != 0 {
		padding := make([]byte, tarRecordSize-remainder)
		if err := tw.writeBytes(padding); err != nil {
			return err
		}
	}

	return nil
}

// tarHeader represents the fields of a USTAR header block.
type tarHeader struct {
	name     string
	prefix   string
	mode     int64
	size     int64
	mtime    int64
	typeflag byte
	linkname string
}

// writeHeaderBlock writes a 512-byte USTAR header block.
func (tw *gnuTarWriter) writeHeaderBlock(h tarHeader) error {
	var block [tarBlockSize]byte

	copy(block[0:100], h.name)           // name
	formatOctal(block[100:108], h.mode)  // mode
	formatOctal(block[108:116], 0)       // uid
	formatOctal(block[116:124], 0)       // gid
	formatOctal(block[124:136], h.size)  // size
	formatOctal(block[136:148], h.mtime) // mtime
	fillSpaces(block[148:156])           // chksum placeholder
	block[156] = h.typeflag              // typeflag
	copy(block[157:257], h.linkname)     // linkname
	copy(block[257:263], "ustar\x00")    // magic
	copy(block[263:265], "00")           // version
	// devmajor (329-336) and devminor (337-344): left as zero bytes (NUL-filled)
	// GNU tar writes all-NUL for non-device files, not "0000000\0"
	copy(block[345:500], h.prefix) // prefix

	// Compute and write checksum
	var checksum int64
	for _, b := range block {
		checksum += int64(b)
	}

	copy(block[148:156], fmt.Sprintf("%06o\x00 ", checksum))

	return tw.writeBytes(block[:])
}

// writeZeroPadding writes zero bytes to align to the next block boundary.
func (tw *gnuTarWriter) writeZeroPadding(dataSize int) error {
	remainder := dataSize % tarBlockSize
	if remainder == 0 {
		return nil
	}

	padding := make([]byte, tarBlockSize-remainder)

	return tw.writeBytes(padding)
}

// formatOctal writes a zero-padded, NUL-terminated octal string into dst.
func formatOctal(dst []byte, val int64) {
	width := len(dst) - 1
	s := fmt.Sprintf("%0*o", width, val)

	if len(s) > width {
		s = s[len(s)-width:]
	}

	copy(dst, s)
	dst[len(dst)-1] = 0
}

// fillSpaces fills a byte slice with space characters (0x20).
func fillSpaces(dst []byte) {
	for i := range dst {
		dst[i] = ' '
	}
}

// buildPAXRecords constructs PAX extended header data.
// Only path and linkpath records are included (atime/ctime are deleted per --pax-option).
func buildPAXRecords(tarPath string, linkname string) string {
	var records string
	records += formatPAXRecord("path", tarPath)

	if len(linkname) > tarNameSize {
		records += formatPAXRecord("linkpath", linkname)
	}

	return records
}

// formatPAXRecord formats a single PAX record as "length key=value\n".
// The length includes itself, allowing correct parsing.
func formatPAXRecord(key, value string) string {
	content := key + "=" + value + "\n"

	for width := 1; width <= 10; width++ {
		totalLen := width + 1 + len(content)
		s := strconv.Itoa(totalLen)

		if len(s) == width {
			return s + " " + content
		}
	}

	// Fallback for extremely large records
	totalLen := tarPAXLenFallbackWidth + 1 + len(content)

	return strconv.Itoa(totalLen) + " " + content
}

// buildPAXHeaderName generates the PAX header name matching GNU tar's
// --pax-option=exthdr.name=%d/PaxHeaders/%f format, truncated to 100 chars.
func buildPAXHeaderName(tarPath string) string {
	p := strings.TrimRight(tarPath, "/")

	idx := strings.LastIndex(p, "/")

	var dir, base string
	if idx < 0 {
		dir = "."
		base = p
	} else {
		dir = p[:idx]
		base = p[idx+1:]
	}

	name := dir + "/PaxHeaders/" + base

	return truncateToNameSize(name)
}

// truncateToNameSize truncates a string to fit in the USTAR name field.
func truncateToNameSize(name string) string {
	if len(name) > tarNameSize {
		return name[:tarNameSize]
	}

	return name
}
