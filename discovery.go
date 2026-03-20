package matr

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/euforic/matr/parser"
)

const (
	defaultMatrFile    = "Matrfile.go"
	defaultCacheFolder = ".matr"
)

var (
	matrFilePath string
	helpFlag     bool
	versionFlag  bool
	cleanFlag    bool
	noCacheFlag  bool
	timeoutFlag  time.Duration
)

// Run is the primary entrypoint to matr's CLI tool.
func Run() {
	fs := flag.NewFlagSet("matr", flag.ExitOnError)
	fs.StringVar(&matrFilePath, "matrfile", "./Matrfile.go", "path to Matrfile")
	fs.BoolVar(&cleanFlag, "clean", false, "clean the matr cache")
	fs.BoolVar(&helpFlag, "h", false, "display usage info")
	fs.BoolVar(&versionFlag, "v", false, "display version")
	fs.BoolVar(&noCacheFlag, "no-cache", false, "don't use the matr cache")
	fs.DurationVar(&timeoutFlag, "timeout", defaultTaskTimeout, "task execution timeout; 0 disables the timeout")
	if err := fs.Parse(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if cleanFlag {
		if err := clean(matrFilePath); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if helpFlag {
		fs.Usage()
		return
	}

	if versionFlag {
		_, _ = fmt.Fprintf(os.Stdout, "\nmatr version: %s\n\n", Version)
		return
	}

	matrCachePath, err := build(matrFilePath, noCacheFlag)
	if err != nil {
		fs.Usage()
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := run(matrCachePath, timeoutFlag, fs.Args()...); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func clean(matrfilePath string) error {
	matrfilePath, err := getMatrfilePath(matrfilePath)
	if err != nil {
		return err
	}

	cachePath := filepath.Join(filepath.Dir(matrfilePath), defaultCacheFolder)
	return os.RemoveAll(cachePath)
}

func parseMatrfile(path string) ([]parser.Command, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	matrFilePath, err := getMatrfilePath(absPath)
	if err != nil {
		return nil, err
	}

	return parser.Parse(matrFilePath)
}

func run(matrCachePath string, timeout time.Duration, args ...string) error {
	if _, err := os.Stat(filepath.Join(matrCachePath, "matr")); err != nil {
		return errors.New("matrfile has not been compiled")
	}
	c := exec.Command(filepath.Join(matrCachePath, "matr"), args...)
	c.Env = append(os.Environ(), timeoutEnvVar+"="+timeout.String())
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	return c.Run()
}

func build(matrFilePath string, noCache bool) (string, error) {
	absPath, err := filepath.Abs(matrFilePath)
	if err != nil {
		return "", err
	}

	newHash, err := buildHash(absPath)
	if err != nil {
		return "", err
	}

	matrCachePath := filepath.Join(filepath.Dir(absPath), defaultCacheFolder)
	oldHash, err := os.ReadFile(filepath.Join(matrCachePath, "matrfile.sha256"))
	if err == nil && !noCache && bytes.Equal(oldHash, newHash) {
		if _, statErr := os.Stat(filepath.Join(matrCachePath, "matr")); statErr == nil {
			return matrCachePath, nil
		}
	}

	if dir, err := os.Stat(matrCachePath); err != nil || !dir.IsDir() {
		if err := os.Mkdir(matrCachePath, 0o777); err != nil {
			return "", err
		}
	}

	if err := os.WriteFile(filepath.Join(matrCachePath, "matrfile.sha256"), newHash, 0o644); err != nil {
		return "", err
	}

	if !symlinkValid(matrCachePath) {
		_ = os.Remove(filepath.Join(matrCachePath, defaultMatrFile))
		if err := os.Symlink(absPath, filepath.Join(matrCachePath, defaultMatrFile)); err != nil && !os.IsExist(err) {
			return "", err
		}
	}

	f, err := os.OpenFile(filepath.Join(matrCachePath, "main.go"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	cmds, err := parseMatrfile(absPath)
	if err != nil {
		return "", err
	}

	if err := generate(cmds, f); err != nil {
		return "", err
	}

	cmd := exec.Command("go", "build", "-tags", "matr", "-o", filepath.Join(matrCachePath, "matr"),
		filepath.Join(matrCachePath, "Matrfile.go"),
		filepath.Join(matrCachePath, "main.go"),
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return matrCachePath, cmd.Run()
}

func getSha256(path string) ([]byte, error) {
	return hashFiles(path)
}

func buildHash(matrfilePath string) ([]byte, error) {
	inputs, err := buildHashInputs(matrfilePath)
	if err != nil {
		return nil, err
	}
	return hashFiles(inputs...)
}

func hashFiles(paths ...string) ([]byte, error) {
	h := sha256.New()
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(h, f); err != nil {
			_ = f.Close()
			return nil, err
		}
		if err := f.Close(); err != nil {
			return nil, err
		}
	}
	return h.Sum(nil), nil
}

func buildHashInputs(matrfilePath string) ([]string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, errors.New("unable to resolve build inputs")
	}

	return []string{
		matrfilePath,
		execPath,
	}, nil
}

func getMatrfilePath(mfpath string) (string, error) {
	absPath, err := filepath.Abs(mfpath)
	if err != nil {
		return "", err
	}
	fp, err := os.Stat(absPath)
	if err != nil {
		return "", errors.New("unable to find Matrfile: " + absPath)
	}

	if !fp.IsDir() {
		return absPath, nil
	}

	matrFilePath := filepath.Join(absPath, "Matrfile")
	if _, err = os.Stat(matrFilePath + ".go"); err == nil {
		return matrFilePath + ".go", nil
	}
	if _, err := os.Stat(matrFilePath); err == nil {
		return matrFilePath, nil
	}

	return "", errors.New("unable to find Matrfile")
}

func symlinkValid(path string) bool {
	pth, err := os.Readlink(filepath.Join(path, "Matrfile.go"))
	if err != nil {
		return false
	}
	if _, err := os.Stat(pth); err != nil {
		return false
	}
	return true
}
