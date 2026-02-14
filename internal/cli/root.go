package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/lockfile"
	"github.com/cffnpwr/nix-prefetch-pnpm-deps/internal/logger"
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
	preInstallCommandFlag.Register(rootCmd)
	hashFlag.Register(rootCmd)
	quietFlag.Register(rootCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func initPnpm(osFs afero.Fs, logger logger.Logger, pnpmPath string) (*pnpm.Pnpm, error) {
	if pnpmPath != "" {
		p, pnpmErr := pnpm.New(osFs, logger, pnpmPath)
		return p, pnpmErr
	}

	p, pnpmErr := pnpm.WithPathEnvVar(osFs, logger)
	return p, pnpmErr
}

// validateLockfileVersion verifies lockfile version is compatible with pnpm version.
func validateLockfileVersion(l *lockfile.Lockfile, p *pnpm.Pnpm) error {
	lockfileVer, lockfileVerErr := l.MajorVersion()
	if lockfileVerErr != nil {
		return lockfileVerErr
	}

	pnpmVer, pnpmVerErr := p.MajorVersion()
	if pnpmVerErr != nil {
		return pnpmVerErr
	}
	if lockfileVer > pnpmVer {
		return fmt.Errorf(
			"lockfileVersion %s in pnpm-lock.yaml is too new for the provided pnpm version %d",
			l.LockfileVersion,
			pnpmVer,
		)
	}

	return nil
}

func computeStoreHash(
	osFs afero.Fs,
	logger logger.Logger,
	storePath string,
	fetcherVersion int,
) (string, error) {
	logger.Debugf("use fetcher version %d", fetcherVersion)

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
	logger.Debug("normalize pnpm store for reproducible hashing")
	normalizeErr := store.Normalize(osFs, store.NormalizeOptions{
		StorePath:      storePath,
		FetcherVersion: fetcherVersion,
	})
	if normalizeErr != nil {
		return "", normalizeErr
	}

	//nolint:mnd // fetcherVersion 3+ uses tarball-based output
	if fetcherVersion >= 3 {
		return computeHashWithTarball(osFs, logger, storePath, fetcherVersion)
	}

	logger.Debugf("compute hash of pnpm store at %s", storePath)
	hash, hashErr := store.Hash(osFs, storePath)
	if hashErr != nil {
		return "", hashErr
	}

	logger.Debugf("computed hash: %s", hash)
	return hash, nil
}

func computeHashWithTarball(
	osFs afero.Fs,
	logger logger.Logger,
	storePath string,
	fetcherVersion int,
) (string, error) {
	logger.Debug("creating tarball of pnpm store for hashing")

	// Create temporary output directory for .fetcher-version and tarball
	outDir, err := afero.TempDir(osFs, "", "nix-prefetch-pnpm-out-")
	if err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}
	defer func() { _ = osFs.RemoveAll(outDir) }()
	logger.Debugf("created temporary output directory at %s", outDir)

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
	logger.Debugf("created tarball of pnpm store at %s", tarballPath)

	// Hash the output directory (containing .fetcher-version and tarball)
	hash, hashErr := store.Hash(osFs, outDir)
	if hashErr != nil {
		return "", hashErr
	}

	logger.Debugf("computed hash with tarball: %s", hash)
	return hash, nil
}

//nolint:cyclop,funlen // run function is the main command logic
func run(_ *cobra.Command, args []string) error {
	fetcherVersion, err := fetcherVersionFlag.GetIntE()
	if err != nil {
		return err
	}

	pnpmPath := pnpmPathFlag.GetString()
	workspaces := workspaceFlag.GetStringSlice()
	pnpmFlags := pnpmFlagFlag.GetStringSlice()
	preInstallCommands := preInstallCommandFlag.GetStringSlice()
	expectedHash := hashFlag.GetString()
	quiet := quietFlag.GetBool()

	level := slog.LevelInfo
	if quiet {
		level = slog.LevelError
	}
	logger := logger.New(level)
	defer logger.Close()

	logger.Debugf("fetcher version: %d", fetcherVersion)
	logger.Debugf("pnpm path: %s", pnpmPath)
	logger.Debugf("workspaces: %v", workspaces)
	logger.Debugf("extra pnpm flags: %v", pnpmFlags)
	logger.Debugf("pre-install commands: %v", preInstallCommands)
	logger.Debugf("expected hash: %s", expectedHash)

	srcPath := args[0]
	osFs := afero.NewOsFs()

	// Verify lockfile exists and is valid
	lockfilePath := filepath.Join(srcPath, "pnpm-lock.yaml")
	lf, loadErr := lockfile.Load(osFs, lockfilePath)
	if loadErr != nil {
		logger.Fatalf("failed to load pnpm-lock.yaml: %w", loadErr)
	}
	logger.Infof("loaded pnpm-lock.yaml from %s", lockfilePath)

	// Create pnpm instance from explicit path or PATH env var
	p, pnpmErr := initPnpm(osFs, logger, pnpmPath)
	if pnpmErr != nil {
		logger.Fatalf("failed to initialize pnpm: %w", pnpmErr)
	}
	logger.Debugf("initialized pnpm with path: %s", p.Path())

	// Validate lockfile version against pnpm version
	verErr := validateLockfileVersion(lf, p)
	if verErr != nil {
		logger.Fatalf("invalid lockfile version: %w", verErr)
	}
	logger.Debugf(
		"validated lockfile version %s is compatible with pnpm version %s",
		lf.LockfileVersion,
		func() string { v, _ := p.Version(); return v }(),
	)

	// Create temp directory for pnpm store
	storePath, err := afero.TempDir(osFs, "", "nix-prefetch-pnpm-deps-")
	if err != nil {
		logger.Fatalf("failed to create temp directory: %w", err)
	}
	defer func() { _ = osFs.RemoveAll(storePath) }()
	logger.Debugf("created temporary directory for pnpm store at %s", storePath)

	// Run pnpm install to fetch dependencies into the store
	installOpts := pnpm.InstallOptions{
		StorePath:          storePath,
		Workspaces:         workspaces,
		Registry:           os.Getenv("NIX_NPM_REGISTRY"),
		ExtraFlags:         pnpmFlags,
		PreInstallCommands: preInstallCommands,
		WorkingDir:         srcPath,
	}
	installErr := p.Install(osFs, installOpts)
	if installErr != nil {
		logger.Fatalf("failed to install dependencies: %w", installErr)
	}
	logger.Infof("successfully installed dependencies to pnpm store at %s", storePath)

	// Normalize store and compute NAR hash
	hashStepLogger := logger.StepLogger(slog.LevelInfo, "compute NAR hash")
	hash, hashErr := computeStoreHash(osFs, logger, storePath, fetcherVersion)
	if hashErr != nil {
		hashStepLogger.Fail(hashErr)
		logger.Fatalf("failed to compute NAR hash: %w", hashErr)
	}
	hashStepLogger.Done()

	// Verify against expected hash if provided
	if expectedHash != "" {
		if expectedHash != hash {
			logger.Fatalf("hash mismatch:\n  expected %s\n  got %s", expectedHash, hash)
		}
		return nil
	}

	// Close logger (stop TUI) before printing hash directly to stdout.
	_ = logger.Close()

	// Print NAR hash
	fmt.Fprintln(os.Stdout, hash)

	return nil
}
