package utils

import (
	"fmt"
	"testing"

	"detaku/types"
)

func TestGetSupplementaryFilePath(t *testing.T) {
	type Iter struct {
		srcPath          string
		supplFileInfoMap types.FileInfoMap
	}

	var iters = []Iter{
		{
			srcPath: "./foo/picture.jpg",
			supplFileInfoMap: types.FileInfoMap{
				"./foo/picture.jpg.json": {Path: "./foo/picture.jpg.json"},
			},
		},
	}

	for _, iter := range iters {
		filePath, err := GetSupplementaryFilePath(iter.srcPath, iter.supplFileInfoMap)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Printf("Found supplementary file for '%s' at '%s'\n", iter.srcPath, filePath)
	}
}

func TestGetDateFromStr(t *testing.T) {
	type Iter struct {
		str        string
		expectDate bool
	}

	var iters = []Iter{
		{str: "june-5-2012", expectDate: true},
		{str: "03042020", expectDate: true},
		{str: "", expectDate: false},
		{str: "IMG_5863.JPG", expectDate: false},
		{str: "photo20", expectDate: false},
	}

	for _, iter := range iters {
		foundDate := false

		d, err := GetDateFromStr(iter.str)
		if err != nil {
			foundDate = false
		}

		if foundDate != iter.expectDate {
			fmt.Printf("Found date '%s' in string '%s'\n", d, iter.str)
		}
	}
}
