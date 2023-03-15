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
)

// Run is the primary entrypoint to matrs cli tool.
// This is where the matrfile path is resolved, compiled and executed
func Run() {
	fs := flag.NewFlagSet("matr", flag.ExitOnError)
	fs.StringVar(&matrFilePath, "matrfile", "./Matrfile.go", "path to Matrfile")
	fs.BoolVar(&cleanFlag, "clean", false, "clean the matr cache")
	fs.BoolVar(&helpFlag, "h", false, "Display usage info")
	fs.BoolVar(&versionFlag, "v", false, "Display version")
	fs.BoolVar(&noCacheFlag, "no-cache", false, "Don't use the matr cache")
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if cleanFlag {
		if err := clean(matrFilePath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if helpFlag {
		fs.Usage()
		return
	}

	if versionFlag {
		fmt.Printf("\nmatr version: %s\n\n", Version)
		return
	}

	matrCachePath, err := build(matrFilePath, noCacheFlag)
	if err != nil {
		fs.Usage()
		os.Stderr.WriteString(err.Error() + "\n")
		return
	}

	if err := run(matrCachePath, fs.Args()...); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		return
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
	var cmds []parser.Command

	absPath, err := filepath.Abs(path)
	if err != nil {
		return cmds, err
	}

	matrFilePath, err := getMatrfilePath(absPath)
	if err != nil {
		return cmds, err
	}

	cmds, err = parser.Parse(matrFilePath)
	if err != nil {
		return cmds, err
	}

	return cmds, nil
}

func run(matrCachePath string, args ...string) error {
	if _, err := os.Stat(filepath.Join(matrCachePath, "matr")); err != nil {
		return errors.New("matrfile has not been compiled")
	}
	c := exec.Command(filepath.Join(matrCachePath, "matr"), args...)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	return c.Run()
}

func build(matrFilePath string, noCache bool) (string, error) {
	absPath, err := filepath.Abs(matrFilePath)
	if err != nil {
		return "", err
	}

	newHash, err := getSha256(absPath)
	if err != nil {
		return "", err
	}

	matrCachePath := filepath.Join(filepath.Dir(absPath), defaultCacheFolder)

	oldHash, err := os.ReadFile(filepath.Join(matrCachePath, "matrfile.sha256"))
	if err == nil && !noCache {
		if ok := bytes.Equal(oldHash, newHash); ok {
			return matrCachePath, nil
		}
	}

	if dir, err := os.Stat(matrCachePath); err != nil || !dir.IsDir() {
		if err := os.Mkdir(matrCachePath, 0777); err != nil {
			return "", err
		}
	}

	if err := os.WriteFile(filepath.Join(matrCachePath, "matrfile.sha256"), newHash, 0644); err != nil {
		return "", err
	}

	if !symlinkValid(matrCachePath) {
		os.Remove(filepath.Join(matrCachePath, defaultMatrFile))
		if err := os.Symlink(absPath, filepath.Join(matrCachePath, defaultMatrFile)); err != nil {
			if os.IsExist(err) {
				return "", err
			}
		}
	}

	f, err := os.OpenFile(filepath.Join(matrCachePath, "main.go"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return "", err
	}
	defer f.Close()

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
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
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
