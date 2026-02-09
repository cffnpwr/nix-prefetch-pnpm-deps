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
	if normalizeErr := store.Normalize(osFs, store.NormalizeOptions{
		StorePath:      storePath,
		FetcherVersion: fetcherVersion,
	}); normalizeErr != nil {
		return "", normalizeErr
	}

	hash, hashErr := store.Hash(osFs, storePath)
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
