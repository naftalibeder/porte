package lib

import (
	"fmt"
	"os"
	"testing"
)

func TestGetLibs(t *testing.T) {
	_ = os.RemoveAll(libCacheDir)

	out, err := GetLibs()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Finished installing external libraries:\n---\n%s---\n", out)
}
