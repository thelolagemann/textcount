package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/afero"
)

var (
	mainChars int = 2017
	mainLines int = 4
	mainFile  *os.File
	mainBytes []byte = []byte(` Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nullam maximus suscipit fringilla. Phasellus eu accumsan est. Donec eu purus vel turpis imperdiet malesuada. Donec in mauris mauris. Aliquam posuere, sapien non aliquet pulvinar, nisi est placerat est, ullamcorper ornare sapien nulla ut neque. Curabitur viverra ligula eu metus pulvinar dignissim. Nam ut lobortis quam, eu luctus lectus. Donec nibh lectus, ultricies vitae nunc vel, gravida ultricies orci. Nunc elementum sapien tortor, in venenatis libero tempor vitae. Maecenas non lacus at arcu viverra sagittis. Nam lobortis nunc in elit fermentum ullamcorper. Sed dui nibh, scelerisque volutpat felis non, porta vehicula ante. Ut tristique ullamcorper tempor. Ut id diam in nibh blandit ultrices a quis tellus. Etiam at auctor massa. 

  Vivamus maximus condimentum urna, nec blandit dolor viverra iaculis. Donec at volutpat nisl. Nulla tincidunt quam eros, ut pharetra ipsum placerat sit amet. Sed justo elit, efficitur sit amet risus in, tempus aliquet nibh. Etiam augue urna, aliquet sit amet ultricies in, malesuada quis odio. Sed dui sem, gravida cursus tempor a, tristique vitae nulla. Suspendisse posuere scelerisque orci, a iaculis ex ultricies non. Aliquam scelerisque ligula non turpis aliquam varius. Mauris purus sapien, dapibus ac laoreet a, finibus nec mauris. Sed felis lacus, aliquam vel ex ornare, imperdiet consequat dui. Donec lobortis ex nisi, vel sodales nunc ornare ac. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Sed accumsan in lectus at euismod. Maecenas et lorem pretium, pretium arcu ut, facilisis tortor. 

   In ut erat ac felis volutpat vulputate ac in justo. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Proin egestas sem ex. Integer pharetra velit non sagittis hendrerit. Cras accumsan quam eget lorem ultrices finibus. Nulla rhoncus accumsan nulla et porttitor. Etiam et tellus lacinia, laoreet ligula molestie, tristique lectus. `)
)

func init() {
	var err error

	//ignorer, err = ignore.CompileIgnoreFile(filepath.Join(folder, ".gitignore"))
	if err != nil {
		panic(err)
	}
}

func TestMain(t *testing.T) {
	os.Args = []string{"textcount", "./"}
	main()
}

func TestMain_NoDirectory(t *testing.T) {
	os.Args = []string{"textcount", "-folder=./invalid-directory"}
	assertLogFatal(t, main)
}

func TestMain_NoAccessFilePaths(t *testing.T) {
	fs = afero.NewMemMapFs()
	fs.Mkdir("inaccessible", 0000)
	afero.WriteFile(fs, "inaccessible/beans", []byte("beans"), 0000)
	os.Args = []string{"textcount", "-folder=./inaccessible"}
	assertLogFatal(t, main)
}

func TestCountFile(t *testing.T) {
	fs = afero.NewMemMapFs()
	afero.WriteFile(fs, "testcount1", mainBytes[128:], 0644)
	afero.WriteFile(fs, "testcount2", mainBytes, 0644)

	tests := []struct {
		input     string
		wantChars int
		wantLines int
	}{
		{input: "testcount1", wantChars: mainChars - 128, wantLines: mainLines},
		{input: "testcount2", wantChars: mainChars, wantLines: mainLines},
	}

	for _, tc := range tests {
		c, l, err := countFile(tc.input)
		if err != nil {
			t.Error(err)
		}
		if c != tc.wantChars {
			t.Errorf("expected: %v, got: %v", mainChars-128, c)
		}
		if l != tc.wantLines {
			t.Errorf("expected: %v, got: %v", mainLines, l)
		}
	}

}

func TestCountFile_OpenInvalidFile(t *testing.T) {
	fs = afero.NewMemMapFs()
	_, _, err := countFile("invalid-file")

	if err == nil {
		t.Error("expecting error, got none")
	}
}

func TestCountCharsFromBytes(t *testing.T) {
	if n := countCharsFromBytes(mainBytes); n != mainChars {
		t.Errorf("expecting char length of %v, got %v", mainChars, n)
	}

}

func TestCountLinesFromBytes(t *testing.T) {
	if n := countLinesFromBytes(mainBytes); n != mainLines {
		t.Errorf("expecting line length of %v, got %v", mainLines, n)
	}
}

func TestGetFilePaths(t *testing.T) {
	folder = "./"
	if _, err := getFilePaths(); err != nil {
		t.Errorf("listing files resulted in an error: %v", err)
	}

	folder = "./invalid-directory"
	if _, err := getFilePaths(); err == nil {
		t.Error("expecting invalid directory error, got none")
	}
}

func TestLoadIgnorer(t *testing.T) {
	folder = "./"
	if err := loadIgnorer(); err != nil {
		t.Error(err)
	}
}

func TestLoadIgnorer_InvalidFile(t *testing.T) {
	folder = "./invalid-directory"
	if err := loadIgnorer(); err == nil {
		t.Error("expecting getIgnore to return error for invalid directory")
	}
}

func TestShouldSkip(t *testing.T) {
	for _, skip := range skipFiles {
		if !shouldSkip(skip) {
			t.Errorf("%v didn't skip", skip)
		}
	}

	if shouldSkip("invalidskip") {
		t.Error("invalid skip")
	}
}

func TestShouldSkip_NotSkip(t *testing.T) {
	skips := []string{"notskippedfile", "anothernotskippedfile"}
	for _, skip := range skips {
		if shouldSkip(skip) {
			t.Errorf("%v skipped", skip)
		}
	}
}

func BenchmarkCountFile(b *testing.B) {
	fs = afero.NewMemMapFs()
	afero.WriteFile(fs, "testcount", mainBytes, 0644)
	for n := 0; n < b.N; n++ {
		countFile("testcount")
	}
}

func BenchmarkGetFilesPaths(b *testing.B) {
	folder = "./"

	for n := 0; n < b.N; n++ {
		getFilePaths()
	}
}

func BenchmarkGetIgnore(b *testing.B) {
	folder = "./"
	for n := 0; n < b.N; n++ {
		loadIgnorer()
	}
}

func BenchmarkShouldSkip(b *testing.B) {
	for n := 0; n < b.N; n++ {
		shouldSkip("go.mod")
	}
}

func BenchmarkCountLinesFromBytes(b *testing.B) {
	for n := 0; n < b.N; n++ {
		countLinesFromBytes(mainBytes)
	}
}

func BenchmarkCountCharsFromBytes(b *testing.B) {
	for n := 0; n < b.N; n++ {
		countCharsFromBytes(mainBytes)
	}
}

func assertLogFatal(t *testing.T, f func()) {
	origLogFatal := logFatalF
	defer func() { logFatalF = origLogFatal }()

	errors := []string{}
	logFatalF = func(format string, args ...interface{}) {
		if len(args) > 0 {
			errors = append(errors, fmt.Sprintf(format, args))
		} else {
			errors = append(errors, format)
		}
	}
	f()
	if len(errors) != 1 {
		t.Errorf("expecting logFatal, got %v", len(errors))
	}
}
