//go:build integration

package integration_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/cli"
)

const nixpkgsRev = "a1bab9e494f5f4939442a57a58d0449a109593fe"

// requireNix checks nix availability, skips test if not found.
func requireNix(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("nix"); err != nil {
		t.Skip("nix is not available, skipping integration test")
	}
}

// nixBuild runs "nix build <args> --no-link --print-out-paths" and returns the store path.
func nixBuild(t *testing.T, args ...string) string {
	t.Helper()

	cmdArgs := append([]string{"build"}, args...)
	cmdArgs = append(cmdArgs, "--no-link", "--print-out-paths")

	cmd := exec.Command("nix", cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("nix build failed: %v\nstderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String())
}

// getPnpmPath returns the pnpm binary path from pinned nixpkgs.
func getPnpmPath(t *testing.T, pnpmPkg string) string {
	t.Helper()

	storePath := nixBuild(t, fmt.Sprintf("github:NixOS/nixpkgs/%s#%s", nixpkgsRev, pnpmPkg))

	return storePath + "/bin/pnpm"
}

// fetchSource fetches a GitHub source via nix fetchFromGitHub.
func fetchSource(t *testing.T, owner, repo, rev, hash string) string {
	t.Helper()

	nixpkgsTarball := fmt.Sprintf(
		"https://github.com/NixOS/nixpkgs/archive/%s.tar.gz",
		nixpkgsRev,
	)
	expr := fmt.Sprintf(
		`(import (builtins.fetchTarball "%s") {}).fetchFromGitHub `+
			`{ owner = "%s"; repo = "%s"; rev = "%s"; hash = "%s"; }`,
		nixpkgsTarball, owner, repo, rev, hash,
	)

	return nixBuild(t, "--impure", "--expr", expr)
}

// copyToWritable copies the nix store source to a writable temp directory.
// Nix store paths are read-only, but pnpm install needs to create node_modules.
func copyToWritable(t *testing.T, storePath string) string {
	t.Helper()

	dst := t.TempDir()

	cmd := exec.Command("cp", "-a", storePath+"/.", dst)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to copy source: %v\n%s", err, out)
	}

	// Make writable (nix store files are read-only)
	cmd = exec.Command("chmod", "-R", "u+w", dst)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to chmod: %v\n%s", err, out)
	}

	return dst
}

// executeCLI sets os.Args, captures stdout, calls cli.Execute(), returns (output, error).
func executeCLI(t *testing.T, args []string) (string, error) {
	t.Helper()

	// Save and restore os.Args
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Save and restore os.Stdout
	origStdout := os.Stdout
	defer func() { os.Stdout = origStdout }()

	// Create pipe for stdout capture
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	os.Stdout = w

	// Start goroutine to read from pipe (prevents deadlock with large pnpm output)
	outputCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outputCh <- buf.String()
	}()

	// Set os.Args - use "cmd" (different from rootCmd.Name()) to trigger os.Args parsing
	os.Args = append([]string{"cmd"}, args...)

	// Call cli.Execute()
	execErr := cli.Execute()

	// Close write end of pipe and wait for reader goroutine
	w.Close()
	output := <-outputCh
	r.Close()

	// Extract hash from last non-empty line of output
	hash := extractLastLine(output)

	return hash, execErr
}

// pnpmCPU maps runtime.GOARCH to pnpm's --cpu value (Node.js process.arch).
func pnpmCPU() string {
	if runtime.GOARCH == "amd64" {
		return "x64"
	}

	return runtime.GOARCH
}

// extractLastLine returns the last non-empty line from the output.
func extractLastLine(output string) string {
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}

	return ""
}

func TestIntegration(t *testing.T) {
	requireNix(t)

	tests := []struct {
		name           string
		fetcherVersion int
		pnpmPkg        string
		owner          string
		repo           string
		rev            string
		srcHash        string
		expectedHash   string
		hashByPlatform map[string]string
	}{
		{
			name:           "[正常系] fetcher v1 - synchrony のハッシュが一致する",
			fetcherVersion: 1,
			pnpmPkg:        "pnpm_9",
			owner:          "relative",
			repo:           "synchrony",
			rev:            "2.4.5",
			srcHash:        "sha256-nJ6H1SZAQCG6U3BPEPmm+BGQa8Af+Vb1E+Lv8lIqDBE=",
			expectedHash:   "sha256-PfgCw2FUEY0OfErfyPnMCLUlO8b4UC/Q5mIG7lezT/w=",
		},
		{
			name:           "[正常系] fetcher v2 - nrm のハッシュが一致する",
			fetcherVersion: 2,
			pnpmPkg:        "pnpm_10",
			owner:          "pana",
			repo:           "nrm",
			rev:            "v2.1.0",
			srcHash:        "sha256-2P0dSZa17A3NslNatCx1edLnrcDtGGpOlk6srcvjL1Y=",
			expectedHash:   "sha256-PENYS5xO2LwT3+TGl/wU2r0ALEj/JQfbkpf/0MJs0uw=",
		},
		{
			name:           "[正常系] fetcher v3 - ccusage のハッシュが一致する",
			fetcherVersion: 3,
			pnpmPkg:        "pnpm_10",
			owner:          "ryoppippi",
			repo:           "ccusage",
			rev:            "v18.0.5",
			srcHash:        "sha256-GopiyaY8lfrgV2tRDSy+qC5AndxIHtGbsAJ51mRi8mU=",
			hashByPlatform: map[string]string{
				"darwin-arm64": "sha256-VIlwcVlhObqj3WKbB2qhGqDByisQvMrzquQZ5V1NgNE=",
				"darwin-amd64": "sha256-YXtCOh8RnSDMuA9DRvkHOTxIJnqY5+coMlkuGobG1QY=",
				"linux-amd64":  "sha256-54d38QVrs2J+i/XH/uzNoFhjqkgTMtZMhbIw9YCYjNA=",
				"linux-arm64":  "sha256-X852T/pffK1XP0ztK0rPzhyVWHSVQxhtgMuD3hlrMWE=",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pnpmPath := getPnpmPath(t, tt.pnpmPkg)
			storePath := fetchSource(t, tt.owner, tt.repo, tt.rev, tt.srcHash)
			srcPath := copyToWritable(t, storePath)

			args := []string{
				"--fetcher-version", strconv.Itoa(tt.fetcherVersion),
				"--pnpm-path", pnpmPath,
				"--pnpm-flag", "--os=" + runtime.GOOS,
				"--pnpm-flag", "--cpu=" + pnpmCPU(),
				"--hash=",
				srcPath,
			}

			output, err := executeCLI(t, args)
			if err != nil {
				t.Fatalf("cli.Execute() returned error: %v", err)
			}

			expected := tt.expectedHash
			if expected == "" {
				platform := runtime.GOOS + "-" + runtime.GOARCH
				var ok bool
				expected, ok = tt.hashByPlatform[platform]
				if !ok {
					t.Fatalf("no expected hash for platform %s; actual hash: %s (add this to hashByPlatform)", platform, output)
				}
			}

			if output != expected {
				t.Errorf("hash mismatch:\n  got:  %s\n  want: %s", output, expected)
			}
		})
	}
}

func TestIntegrationHashFlag(t *testing.T) {
	requireNix(t)

	pnpmPath := getPnpmPath(t, "pnpm_9")
	storePath := fetchSource(
		t,
		"relative",
		"synchrony",
		"2.4.5",
		"sha256-nJ6H1SZAQCG6U3BPEPmm+BGQa8Af+Vb1E+Lv8lIqDBE=",
	)

	t.Run("[正常系] --hash フラグで正しいハッシュを渡すと成功する", func(t *testing.T) {
		srcPath := copyToWritable(t, storePath)

		args := []string{
			"--fetcher-version", "1",
			"--pnpm-path", pnpmPath,
			"--pnpm-flag", "--os=" + runtime.GOOS,
			"--pnpm-flag", "--cpu=" + pnpmCPU(),
			"--hash", "sha256-PfgCw2FUEY0OfErfyPnMCLUlO8b4UC/Q5mIG7lezT/w=",
			srcPath,
		}

		_, err := executeCLI(t, args)
		if err != nil {
			t.Errorf("expected no error with correct hash, got: %v", err)
		}
	})

	t.Run("[異常系] --hash フラグで不正なハッシュを渡すとエラーになる", func(t *testing.T) {
		srcPath := copyToWritable(t, storePath)

		args := []string{
			"--fetcher-version", "1",
			"--pnpm-path", pnpmPath,
			"--pnpm-flag", "--os=" + runtime.GOOS,
			"--pnpm-flag", "--cpu=" + pnpmCPU(),
			"--hash", "sha256-INVALIDHASH",
			srcPath,
		}

		_, err := executeCLI(t, args)
		if err == nil {
			t.Error("expected error with wrong hash, got nil")
		}
	})
}

func TestIntegrationErrors(t *testing.T) {
	requireNix(t)

	t.Run("[異常系] pnpm-lock.yaml が存在しないとエラーになる", func(t *testing.T) {
		emptyDir := t.TempDir()

		pnpmPath := getPnpmPath(t, "pnpm_9")

		args := []string{
			"--fetcher-version", "1",
			"--pnpm-path", pnpmPath,
			"--hash=",
			emptyDir,
		}

		_, err := executeCLI(t, args)
		if err == nil {
			t.Error("expected error when pnpm-lock.yaml is missing, got nil")
		}
	})

	t.Run("[異常系] 無効な fetcher version を渡すとエラーになる", func(t *testing.T) {
		emptyDir := t.TempDir()

		pnpmPath := getPnpmPath(t, "pnpm_9")

		args := []string{
			"--fetcher-version", "0",
			"--pnpm-path", pnpmPath,
			"--hash=",
			emptyDir,
		}

		_, err := executeCLI(t, args)
		if err == nil {
			t.Error("expected error with invalid fetcher version, got nil")
		}
	})
}
