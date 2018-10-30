package cmd

import (
	"testing"
)

func TestCheckFile(t *testing.T) {
	filename := "util.go"
	falseFile := "not_a_file.go"
	f, _ := checkFile(filename)
	if f == nil {
		t.Errorf("Expected the file %v to exist", filename)
	}

	f, _ = checkFile(falseFile)
	if f != nil {
		t.Errorf("Expected the file %v to not exist", falseFile)
	}
}

func TestNewFileName(t *testing.T) {
	expected := "test.go"

	fn := getNewFileName("test.go", "go")
	if fn != expected {
		t.Errorf("Expected the filename to be %s but instead got %s", expected, fn)
	}

	fn = getNewFileName("test", "go")
	if fn != expected {
		t.Errorf("Expected the filename to be %s but instead got %s", expected, fn)
	}
}
