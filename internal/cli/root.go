package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/lockfile"
	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/pnpm"
	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/store"
)

var rootCmd = &cobra.Command{
	Use:   "nix-prefetch-pnpm-deps [source-dir]",
	Short: "prefetch dependencies for pnpm",
	Args:  cobra.ExactArgs(1),
	RunE:  run,
}

func init() {
	fetcherVersionFlag.Register(rootCmd)
	pnpmPathFlag.Register(rootCmd)
	workspaceFlag.Register(rootCmd)
	pnpmFlagFlag.Register(rootCmd)
	hashFlag.Register(rootCmd)
	quietFlag.Register(rootCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func initPnpm(osFs afero.Fs, pnpmPath string) (*pnpm.Pnpm, error) {
	if pnpmPath != "" {
		p, pnpmErr := pnpm.New(osFs, pnpmPath)
		return p, pnpmErr
	}

	p, pnpmErr := pnpm.WithPathEnvVar(osFs)
	return p, pnpmErr
}

func computeStoreHash(osFs afero.Fs, storePath string, fetcherVersion int) (string, error) {
	// Write .fetcher-version to the store directory before Normalize,
	// because Normalize sets dirs to 0o555 (read-only).
	// For v3+, .fetcher-version is written to a separate output directory in computeHashWithTarball.
	//nolint:mnd // fetcherVersion 2 is the only version that writes .fetcher-version to the store
	if fetcherVersion == 2 {
		fetcherVersionPath := filepath.Join(storePath, ".fetcher-version")
		versionContent := fmt.Sprintf("%d\n", fetcherVersion)
		writeErr := afero.WriteFile(osFs, fetcherVersionPath, []byte(versionContent), 0o444)
		if writeErr != nil {
			return "", fmt.Errorf("failed to write .fetcher-version: %w", writeErr)
		}
	}

	// Normalize the pnpm store (remove tmp/projects, normalize JSON, set permissions)
	if normalizeErr := store.Normalize(osFs, store.NormalizeOptions{
		StorePath:      storePath,
		FetcherVersion: fetcherVersion,
	}); normalizeErr != nil {
		return "", normalizeErr
	}

	//nolint:mnd // fetcherVersion 3+ uses tarball-based output
	if fetcherVersion >= 3 {
		return computeHashWithTarball(osFs, storePath, fetcherVersion)
	}

	hash, hashErr := store.Hash(osFs, storePath)
	if hashErr != nil {
		return "", hashErr
	}

	return hash, nil
}

func computeHashWithTarball(osFs afero.Fs, storePath string, fetcherVersion int) (string, error) {
	// Create temporary output directory for .fetcher-version and tarball
	outDir, err := os.MkdirTemp("", "nix-prefetch-pnpm-out-")
	if err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}
	defer os.RemoveAll(outDir)

	// Write .fetcher-version file
	fetcherVersionPath := filepath.Join(outDir, ".fetcher-version")
	versionContent := fmt.Sprintf("%d\n", fetcherVersion)
	//nolint:mnd // read-only file permissions
	writeErr := afero.WriteFile(
		osFs,
		fetcherVersionPath,
		[]byte(versionContent),
		0o444,
	)
	if writeErr != nil {
		return "", fmt.Errorf("failed to write .fetcher-version: %w", writeErr)
	}

	// Create reproducible tarball
	tarballPath := filepath.Join(outDir, "pnpm-store.tar.zst")
	if tarballErr := store.CreateTarball(osFs, storePath, tarballPath); tarballErr != nil {
		return "", tarballErr
	}

	// Hash the output directory (containing .fetcher-version and tarball)
	hash, hashErr := store.Hash(osFs, outDir)
	if hashErr != nil {
		return "", hashErr
	}

	return hash, nil
}

func run(_ *cobra.Command, args []string) error {
	fetcherVersion, err := fetcherVersionFlag.GetIntE()
	if err != nil {
		return err
	}

	pnpmPath := pnpmPathFlag.GetString()
	workspaces := workspaceFlag.GetStringSlice()
	pnpmFlags := pnpmFlagFlag.GetStringSlice()
	expectedHash := hashFlag.GetString()
	quiet := quietFlag.GetBool()

	srcPath := args[0]
	osFs := afero.NewOsFs()

	// Verify lockfile exists and is valid
	lockfilePath := filepath.Join(srcPath, "pnpm-lock.yaml")
	if _, loadErr := lockfile.Load(osFs, lockfilePath); loadErr != nil {
		return loadErr
	}

	// Create pnpm instance from explicit path or PATH env var
	p, pnpmErr := initPnpm(osFs, pnpmPath)
	if pnpmErr != nil {
		return pnpmErr
	}

	// Create temp directory for pnpm store
	storePath, err := os.MkdirTemp("", "nix-prefetch-pnpm-deps-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(storePath)

	// Run pnpm install to fetch dependencies into the store
	if installErr := p.Install(pnpm.InstallOptions{
		StorePath:  storePath,
		Workspaces: workspaces,
		Registry:   os.Getenv("NIX_NPM_REGISTRY"),
		ExtraFlags: pnpmFlags,
		WorkingDir: srcPath,
	}); installErr != nil {
		return installErr
	}

	// Normalize store and compute NAR hash
	hash, hashErr := computeStoreHash(osFs, storePath, fetcherVersion)
	if hashErr != nil {
		return hashErr
	}

	// Verify against expected hash if provided
	if expectedHash != "" && hash != expectedHash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash, hash)
	}

	if !quiet {
		fmt.Fprintln(os.Stdout, hash)
	}

	return nil
}
