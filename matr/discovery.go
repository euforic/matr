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
	// TODO: clean up this shit show
	// create a new flagset
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
	var err error
	var cmds []parser.Command

	matrFilePath, err = filepath.Abs(matrFilePath)
	if err != nil {
		return cmds, err
	}

	matrFilePath, err = getMatrfilePath(matrFilePath)
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
	c := exec.Command(filepath.Join(matrCachePath, "matr"), args...)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	return c.Run()
}

func build(matrFilePath string, noCache bool) (string, error) {
	// get absolute path to matrfile
	matrFilePath, err := filepath.Abs(matrFilePath)
	if err != nil {
		return "", err
	}

	matrCachePath := filepath.Join(filepath.Dir(matrFilePath), ".matr")

	// check if the matrfile has changed
	newHash, err := getSha256(matrFilePath)
	if err != nil {
		return "", err
	}

	// read the hash from the matrfileSha256 file
	oldHash, err := os.ReadFile(filepath.Join(matrCachePath, "matrfile.sha256"))
	if err == nil && !noCache {
		// if the hash is the same, we can skip the build
		if ok := bytes.Equal(oldHash, newHash); ok {
			return matrCachePath, nil
		}
	}

	// check if the cache folder exists
	if dir, err := os.Stat(matrCachePath); err != nil || !dir.IsDir() {
		if err := os.Mkdir(matrCachePath, 0777); err != nil {
			return "", err
		}
	}

	// if the file doesn't exist, create it
	if err := os.WriteFile(filepath.Join(matrCachePath, "matrfile.sha256"), []byte(newHash), 0644); err != nil {
		return "", err
	}

	if !symlinkValid(matrCachePath) {
		os.Remove(filepath.Join(matrCachePath, defaultMatrFile))
		if err := os.Symlink(matrFilePath, filepath.Join(matrCachePath, defaultMatrFile)); err != nil {
			if os.IsExist(err) {
				return "", err
			}
		}
	}

	// create the main.go file in the matr cache folder
	// for the generated code to write to
	f, err := os.OpenFile(filepath.Join(matrCachePath, "main.go"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return "", err
	}
	defer f.Close()

	cmds, err := parseMatrfile(matrFilePath)
	if err != nil {
		return "", err
	}

	if err := generate(cmds, f); err != nil {
		return "", err
	}

	// TODO: check if we need to rebuild
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

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func getMatrfilePath(matrFilePath string) (string, error) {
	matrFilePath, err := filepath.Abs(matrFilePath)
	if err != nil {
		return "", err
	}

	fp, err := os.Stat(matrFilePath)
	if err != nil {
		return "", errors.New("unable to find Matrfile: " + matrFilePath)
	}

	if !fp.IsDir() {
		return matrFilePath, nil
	}

	matrFilePath = filepath.Join(matrFilePath, "Matrfile")

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
