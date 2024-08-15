package detaku

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"detaku/exif"
	"detaku/utils"

	"gopkg.in/yaml.v3"
)

type TestDesc struct {
	Success []TestFileDesc `yaml:"success"`
	Fail    []TestFileDesc `yaml:"fail"`
	Skip    []TestFileDesc `yaml:"skip"`
}

type TestFileDesc struct {
	FileName string   `yaml:"name"`
	Tags     TestTags `yaml:"tags"`
}

type TestTags struct {
	Misc  []TestTag `yaml:"misc"`
	Dates []TestTag `yaml:"dates"`
	Geo   []TestTag `yaml:"geo"`
}

type TestTag struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

var (
	//go:embed fixtures
	fixturesDir embed.FS
)

func TestRun(t *testing.T) {
	globalCacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Fatalf("Error getting user cache directory: %s", err)
	}

	fixtureEntries, err := fixturesDir.ReadDir("fixtures")
	if err != nil {
		t.Fatalf("Error finding fixtures directory: %s", err)
	}
	fmt.Printf("Found %d test fixtures\n", len(fixtureEntries))

	for _, fixtureEntry := range fixtureEntries {
		inDir := filepath.Join("fixtures", fixtureEntry.Name())

		outDirRoot := filepath.Join(globalCacheDir, "detaku", "tests", fixtureEntry.Name())
		_ = os.RemoveAll(outDirRoot)
		err := os.MkdirAll(outDirRoot, 0777)
		if err != nil {
			t.Fatalf("Error creating test directory: %s", err)
		}

		outDirSuccess := filepath.Join(outDirRoot, "success")
		err = os.Mkdir(outDirSuccess, 0777)
		if err != nil {
			t.Fatalf("Error creating subdirectory: %s", err)
		}

		// Parse expected description.

		expectedBt, err := os.ReadFile(filepath.Join(inDir, "expected.yaml"))
		if err != nil {
			t.Fatalf("Error reading expected description: %s", err)
		}

		var expected TestDesc
		err = yaml.Unmarshal(expectedBt, &expected)
		if err != nil {
			t.Fatalf("Error parsing yaml: %s", err)
		}

		// Parse and convert all files in the directory.

		fmt.Printf("Processing directory '%s'\n", inDir)
		err = Run(inDir, outDirRoot)
		if err != nil {
			t.Fatalf("Error handling directory '%s': %s", inDir, err)
		}

		fmt.Printf("Finished processing into destination directory '%s':\n", outDirRoot)
		err = filepath.Walk(outDirRoot, func(path string, info fs.FileInfo, err error) error {
			fmt.Printf("- %s\n", path)
			return nil
		})
		if err != nil {
			t.Fatalf("Error walking test directory: %s\n", err)
		}

		// Verify output of all files.

		tmpFiles, err := os.ReadDir(outDirRoot)
		if err != nil {
			t.Fatalf("Error reading temporary directory: %s", err)
		}

		if len(tmpFiles) == 0 {
			t.Fatal("Output directory is empty\n")
		}

		for _, expecteds := range expected.Success {
			// Verify file exists.

			expectedPath := filepath.Join(outDirSuccess, expecteds.FileName)
			fmt.Printf("Expecting file: '%s'\n", expectedPath)

			actualEntries, _ := os.ReadDir(outDirSuccess)
			actualFile, err := os.Stat(expectedPath)
			if err != nil {
				t.Fatalf("Error finding '%s'; instead found %s\n", expectedPath, actualEntries)
			}

			filePath := filepath.Join(outDirSuccess, actualFile.Name())
			fmt.Printf("Found file: '%s'\n", filePath)

			// Verify number of tags in file.

			exifTags, err := exif.GetAllExifTags(filePath)
			if err != nil {
				t.Fatalf("Error getting exif tags at '%s': %s\n", filePath, err)
			}

			fmt.Printf("Found %d tags\n", len(exifTags.Dates)+len(exifTags.Misc))

			// Verify values of file tags.

			for _, expected := range expecteds.Tags.Misc {
				actual, exists := exifTags.Misc[expected.Name]
				if !exists {
					t.Fatalf("- Unable to find value for expected tag %s\n", expected.Name)
				}
				if actual.Value != expected.Value {
					t.Fatalf("- For tag %s, expected %s but got %s\n", expected.Name, expected.Value, actual.Value)
				}
				fmt.Printf("- For tag %s, found expected value %s\n", expected.Name, actual.Value)
			}

			for _, expected := range expecteds.Tags.Dates {
				actual, exists := exifTags.Dates[expected.Name]
				if !exists {
					t.Fatalf("- Unable to find value for expected tag %s\n", expected.Name)
				}
				expectedDate, err := time.Parse(utils.GoParseExifToolDateFmt, expected.Value)
				if err != nil {
					t.Fatalf("- Invalid date for expected tag %s: %s\n", expected.Name, expected.Value)
				}
				if !actual.Date.Equal(expectedDate) {
					t.Fatalf("- For tag %s, expected %s but got %s\n", expected.Name, expectedDate, actual.Date)
				}
				fmt.Printf("- For tag %s, found expected date %s\n", expected.Name, actual.Date)
			}

			for _, expected := range expecteds.Tags.Geo {
				actual, exists := exifTags.Geo[expected.Name]
				if !exists {
					t.Fatalf("- Unable to find value for expected tag %s\n", expected.Name)
				}
				if actual.Value != expected.Value {
					t.Fatalf("- For tag %s, expected %s but got %s\n", expected.Name, expected.Value, actual.Value)
				}
				fmt.Printf("- For tag %s, found expected value %s\n", expected.Name, actual.Value)
			}
		}
	}
}
