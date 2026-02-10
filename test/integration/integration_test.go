//go:build integration

package integration_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const nixpkgsRev = "a1bab9e494f5f4939442a57a58d0449a109593fe"

type src struct {
	owner   string
	repo    string
	rev     string
	srcHash string
}

var (
	_, testFile, _, _ = runtime.Caller(0)
	projectRoot       = filepath.Join(filepath.Dir(testFile), "../..")

	binaryPath string
	nixPath    string

	pnpmPaths = map[string]string{
		"pnpm_9":  "",
		"pnpm_10": "",
	}
	srcPaths = map[string]string{
		"synchrony": "",
		"nrm":       "",
		"ccusage":   "",
	}

	srcs = map[string]src{
		"synchrony": {
			owner:   "relative",
			repo:    "synchrony",
			rev:     "2.4.5",
			srcHash: "sha256-nJ6H1SZAQCG6U3BPEPmm+BGQa8Af+Vb1E+Lv8lIqDBE=",
		},
		"nrm": {
			owner:   "pana",
			repo:    "nrm",
			rev:     "v2.1.0",
			srcHash: "sha256-2P0dSZa17A3NslNatCx1edLnrcDtGGpOlk6srcvjL1Y=",
		},
		"ccusage": {
			owner:   "ryoppippi",
			repo:    "ccusage",
			rev:     "v18.0.5",
			srcHash: "sha256-GopiyaY8lfrgV2tRDSy+qC5AndxIHtGbsAJ51mRi8mU=",
		},
	}
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

// TestMain builds the CLI binary before running tests.
func testMain(m *testing.M) int {
	var err error

	// Get nix path
	fmt.Fprintln(os.Stderr, "Getting nix path...")
	nixPath, err = getNixPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get nix path: %v\n", err)
		return 1
	}

	// Get pnpm paths
	fmt.Fprintln(os.Stderr, "Getting pnpm paths...")
	for pkg := range pnpmPaths {
		fmt.Fprintf(os.Stderr, "  Getting %s\n", pkg)
		pnpmPath, getPnpmErr := getPnpmPath(nixPath, pkg)
		if getPnpmErr != nil {
			fmt.Fprintf(os.Stderr, "failed to get pnpm path for %s: %v\n", pkg, getPnpmErr)
			return 1
		}

		pnpmPaths[pkg] = pnpmPath
	}

	// Get source paths
	fmt.Fprintln(os.Stderr, "Getting source paths...")
	for name, srcInfo := range srcs {
		fmt.Fprintf(os.Stderr, "  Getting source for %s\n", name)
		storePath, fetchErr := fetchSource(nixPath, srcInfo)
		if fetchErr != nil {
			fmt.Fprintf(os.Stderr, "failed to fetch source for %s: %v\n", name, fetchErr)
			return 1
		}

		srcPaths[name] = storePath
	}

	tmpDir, err := os.MkdirTemp("", "integration-test-*")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		return 1
	}

	binaryPath = filepath.Join(tmpDir, "nix-prefetch-pnpm-deps")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = projectRoot

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n%s\n", err, out)
		return 1
	}

	return m.Run()
}

// getNixPath returns the nix binary path.
func getNixPath() (string, error) {
	p, err := exec.LookPath("nix")
	if err != nil {
		return "", fmt.Errorf("nix is not installed or not in PATH: %w", err)
	}

	return p, nil
}

// nixBuild runs "nix build <args> --no-link --print-out-paths" and returns the store path.
func nixBuild(nixPath string, args ...string) (string, error) {
	cmdArgs := make([]string, 0, len(args)+4)
	cmdArgs = append(cmdArgs, "build")
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "--no-link", "--print-out-paths")

	cmd := exec.Command(nixPath, cmdArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("nix build failed: %w\nstderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// getPnpmPath returns the pnpm binary path from pinned nixpkgs.
func getPnpmPath(nixPath, pnpmPkg string) (string, error) {
	storePath, err := nixBuild(
		nixPath,
		fmt.Sprintf("github:NixOS/nixpkgs/%s#%s", nixpkgsRev, pnpmPkg),
	)
	if err != nil {
		return "", fmt.Errorf("failed to build pnpm package %s: %w", pnpmPkg, err)
	}

	return storePath + "/bin/pnpm", nil
}

// fetchSource fetches a GitHub source via nix fetchFromGitHub.
func fetchSource(nixPath string, src src) (string, error) {
	nixpkgsTarball := fmt.Sprintf(
		"https://github.com/NixOS/nixpkgs/archive/%s.tar.gz",
		nixpkgsRev,
	)
	expr := fmt.Sprintf(
		`(import (builtins.fetchTarball "%s") {}).fetchFromGitHub `+
			`{ owner = "%s"; repo = "%s"; rev = "%s"; hash = "%s"; }`,
		nixpkgsTarball, src.owner, src.repo, src.rev, src.srcHash,
	)

	return nixBuild(nixPath, "--impure", "--expr", expr)
}

// copyToWritable copies the nix store source to a writable temp directory.
// Nix store paths are read-only, but pnpm install needs to create node_modules.
func copyToWritable(t *testing.T, storePath string) string {
	t.Helper()

	dst := t.TempDir()

	cmd := exec.Command("cp", "-a", storePath+"/.", dst)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to copy source: %v\n%s", err, out)
	}

	// Make writable (nix store files are read-only)
	cmd = exec.Command("chmod", "-R", "u+w", dst)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to chmod: %v\n%s", err, out)
	}

	return dst
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

// executeCLI runs the CLI binary as a subprocess and returns (last line of stdout, error).
func executeCLI(t *testing.T, args []string) (string, error) {
	t.Helper()

	if binaryPath == "" {
		t.Fatal("binaryPath is not set")
	}
	cmd := exec.Command(binaryPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("%w\nstderr: %s", err, stderr.String())
	}

	hash := extractLastLine(stdout.String())

	return hash, err
}

// pnpmCPU maps runtime.GOARCH to pnpm's --cpu value (Node.js process.arch).
func pnpmCPU() string {
	goarch := runtime.GOARCH
	if goarch == "amd64" {
		return "x64"
	}

	return goarch
}

type hash interface {
	hash(platform string) string
}

type singleHash string

func (h singleHash) hash(_ string) string {
	return string(h)
}

type hashByPlatform map[string]string

func (h hashByPlatform) hash(platform string) string {
	return h[platform]
}

func TestIntegration(t *testing.T) {
	tests := []struct {
		name           string
		fetcherVersion int
		pnpmPath       string
		srcPath        string
		additionalArgs []string
		wantHash       hash
		wantErr        bool
	}{
		{
			name:           "[正常系] fetcher v1 synchronysのハッシュが一致する",
			fetcherVersion: 1,
			pnpmPath:       pnpmPaths["pnpm_9"],
			srcPath:        srcPaths["synchrony"],
			wantHash:       singleHash("sha256-PfgCw2FUEY0OfErfyPnMCLUlO8b4UC/Q5mIG7lezT/w="),
		},
		{
			name:           "[正常系] fetcher v2 nrmのハッシュが一致する",
			fetcherVersion: 2,
			pnpmPath:       pnpmPaths["pnpm_10"],
			srcPath:        srcPaths["nrm"],
			wantHash:       singleHash("sha256-PENYS5xO2LwT3+TGl/wU2r0ALEj/JQfbkpf/0MJs0uw="),
		},
		{
			name:           "[正常系] fetcher v3 ccusageのハッシュが一致する",
			fetcherVersion: 3,
			pnpmPath:       pnpmPaths["pnpm_10"],
			srcPath:        srcPaths["ccusage"],
			additionalArgs: []string{
				"--pnpm-flag", "--cpu=" + pnpmCPU(),
				"--pnpm-flag", "--os=" + runtime.GOOS,
			},
			wantHash: hashByPlatform{
				"amd64-linux":  "sha256-54d38QVrs2J+i/XH/uzNoFhjqkgTMtZMhbIw9YCYjNA=",
				"arm64-linux":  "sha256-X852T/pffK1XP0ztK0rPzhyVWHSVQxhtgMuD3hlrMWE=",
				"amd64-darwin": "sha256-YXtCOh8RnSDMuA9DRvkHOTxIJnqY5+coMlkuGobG1QY=",
				"arm64-darwin": "sha256-VIlwcVlhObqj3WKbB2qhGqDByisQvMrzquQZ5V1NgNE=",
			},
		},
		{
			name:           "[異常系] 不正なfetcherバージョンを渡すとエラーになる",
			fetcherVersion: 99,
			pnpmPath:       pnpmPaths["pnpm_9"],
			srcPath:        srcPaths["synchrony"],
			wantErr:        true,
		},
		{
			name:           "[異常系] 不正なpnpmパスを渡すとエラーになる",
			fetcherVersion: 1,
			pnpmPath:       "/invalid/pnpm/path",
			srcPath:        srcPaths["synchrony"],
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantHash := ""
			if tt.wantHash != nil {
				wantHash = tt.wantHash.hash(fmt.Sprintf("%s-%s", runtime.GOARCH, runtime.GOOS))
			}
			srcPath := copyToWritable(t, tt.srcPath)
			args := []string{
				"--fetcher-version", strconv.Itoa(tt.fetcherVersion),
				"--pnpm-path", tt.pnpmPath,
			}
			args = append(args, tt.additionalArgs...)
			args = append(args, srcPath)

			gotHash, err := executeCLI(t, args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("executeCLI() error: %v", err)
			}
			if gotHash != wantHash {
				t.Fatalf("unexpected hash:\n got:  %s\n want: %s", gotHash, wantHash)
			}
		})
	}
}

func TestIntegrationHashFlag(t *testing.T) {
	tests := []struct {
		name     string
		hashFlag string
		wantErr  bool
	}{
		{
			name:     "[正常系] 正しいハッシュを渡すと成功する",
			hashFlag: "sha256-PfgCw2FUEY0OfErfyPnMCLUlO8b4UC/Q5mIG7lezT/w=",
			wantErr:  false,
		},
		{
			name:     "[異常系] 不正なハッシュを渡すとエラーになる",
			hashFlag: "sha256-INVALIDHASH",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pnpmPath := pnpmPaths["pnpm_9"]
			srcPath := copyToWritable(t, srcPaths["synchrony"])

			args := []string{
				"--fetcher-version", "1",
				"--pnpm-path", pnpmPath,
				"--hash", tt.hashFlag,
				srcPath,
			}

			_, err := executeCLI(t, args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("executeCLI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
