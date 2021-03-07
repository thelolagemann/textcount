package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/afero"

	"golang.org/x/tools/godoc/util"
)

var (
	skipFiles []string
	ignorer   ignore.IgnoreParser
	folder    string
	err       error
	chars     int
	lines     int
	skip      bool
	skipped   *string

	fs afero.Fs

	logFatalF = log.Fatalf
)

func init() {
	flag.StringVar(&folder, "folder", "./", "the folder to count")
	skipped = flag.String("skipped", "go.mod,go.sum", "comma seperated list of files to skip counting")

	fs = afero.NewOsFs()
}

func main() {
	flag.Parse()
	for _, skipped := range strings.Split(*skipped, ",") {
		skipFiles = append(skipFiles, skipped)
	}
	loadIgnorer()
	if _, err := fs.Stat(folder); err != nil {
		logFatalF("error accessing %v, %v", folder, err)
		return
	}

	files, err := getFilePaths()
	if err != nil {
		logFatalF("couldn't load file paths, %v", err)
		return
	}

	for _, file := range files {
		c, l, err := countFile(file)
		if err != nil {
			fmt.Println("error counting file", err)
			continue
		}
		chars += c
		lines += l
	}
	fmt.Printf("lines\t%v\nchars\t%v\nfiles\t%v\n", lines, chars, len(files))

}

// countFile counts the characters and lines of
// valid text files
func countFile(path string) (chars int, lines int, err error) {
	info, err := fs.Open(path)
	if err != nil {
		return chars, lines, err
	}

	bytes, err := ioutil.ReadAll(info)
	info.Close()
	if err != nil {
		return chars, lines, err
	}

	if isText := util.IsText(bytes[:64]); isText {
		lines += countLinesFromBytes(bytes)
		chars += countCharsFromBytes(bytes)
	}

	return chars, lines, nil
}

// countCharsFromBytes counts the UTF-8 characters
// found in b.
func countCharsFromBytes(b []byte) int {
	return len(string(b))
}

// countLinesFromBytes counts the occurences of newline
// seperators in b.
func countLinesFromBytes(b []byte) int {
	return bytes.Count(b, []byte{'\n'})
}

// loadIgnorer loads a gitignore file and stores it in
// ignorer, returning an error if it failed.
func loadIgnorer() error {
	var err error
	ignorer, err = ignore.CompileIgnoreFile(filepath.Join(folder, ".gitignore"))
	if err != nil {
		return err
	}
	return nil
}

// getFilePaths gets a list of files in the folder
// location. It does not check that the location
// is existing as it's assumed that this is done
// once at init to avoid fs requests.
func getFilePaths() ([]string, error) {
	var files []string
	if err := afero.Walk(fs, folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !shouldSkip(path) && !info.IsDir() {
			files = append(files, path)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

// shouldSkip returns true if f is in the list
// of skipFiles, and file skipping is enabled.
func shouldSkip(f string) bool {
	for _, skip := range skipFiles {
		if strings.HasSuffix(f, skip) {
			return true
		}
	}

	ign, ig := ignorer.(*ignore.GitIgnore)

	if ig && ign != nil {
		return ignorer.MatchesPath(strings.Replace(f, folder, "", 1))
	}

	return false
}
