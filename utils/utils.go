package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"detaku/types"

	"github.com/markusmobius/go-dateparser"
)

const (
	// The string to join the output file date, original name, and count suffix.
	FileNamePartSep = "_"

	// The Go format to use for output files named by date and time.
	FileNameFmt = "2006-01-02" + FileNamePartSep + "15-04-05"

	// The format for date arguments sent to exiftool.
	ExifToolDateFmt = "%Y-%m-%dT%H:%M:%S"

	// The Go format for parsing date values from exiftool queries.
	GoParseExifToolDateFmt = "2006-01-02T15:04:05"
)

// Finds an available path in destDir, trying incrementing suffixes if needed.
func GetValidPath(destDir string, destFileName string) (destPath string) {
	base := filepath.Base(destFileName)
	ext := filepath.Ext(destFileName)
	name := strings.TrimSuffix(base, ext)

	incr := 0

	for destPath == "" {
		destName := fmt.Sprintf("%s%s", name, ext)
		if incr > 0 {
			destName = fmt.Sprintf("%s%s%d%s", name, FileNamePartSep, incr, ext)
		}

		_, err := os.Stat(filepath.Join(destDir, destName))
		if err != nil {
			destPath = filepath.Join(destDir, destName)
		}

		incr++
	}

	return destPath
}

// Finds the file in supplFileInfoMap most likely to match the file at srcPath.
//
// The supplFileInfoMap arg is needed because an image and its corresponding info file
// might be located in different directories, in an export containing multiple archives.
//
// This needs to be expanded to handle these disconnected info files. We currently
// avoid searching for them because a match for just the file name, e.g. IMG_0823.jpg,
// could potentially match multiple files.
func GetSupplementaryFilePath(srcPath string, supplFileInfoMap types.FileInfoMap) (filePath string, err error) {
	fullPath := srcPath + ".json"
	info, exists := supplFileInfoMap[fullPath]
	if exists {
		return info.Path, nil
	}

	return "", fmt.Errorf("no file at '%s'", fullPath)
}

// Parses s and returns a date, if one is represented.
func GetDateFromStr(s string) (date time.Time, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic when checking for date in '%s': %s", s, r)
		}
	}()

	ext := filepath.Ext(s)
	s = strings.TrimSuffix(s, ext)

	// Check if a date might be incorrectly found in an arbitrary number string.
	re := regexp.MustCompile(`\d{9}`)
	exists := re.MatchString(s)
	if exists {
		return time.Time{}, fmt.Errorf("'%s' contains a long numeric string which probably does not represent a date", s)
	}

	// Search the string for a date.
	parser := dateparser.Parser{}
	dates, err := parser.SearchWithLanguage(nil, "en", s)
	if err != nil {
		return time.Time{}, err
	}

	if len(dates) == 0 {
		return time.Time{}, fmt.Errorf("no dates found in '%s'", s)
	}
	d := dates[0]
	return d.Date.Time, nil
}

// Returns a comma-separated list from the provided map, ordered by each
// entry's int value.
func SortedListFromCt(m map[string]int) string {
	items := []types.SortableItem{}

	for val, ct := range m {
		items = append(items, types.SortableItem{
			Value: val,
			Count: ct,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Count > items[j].Count
	})

	disp := ""
	for i, item := range items {
		if i > 0 {
			disp += ", "
		}
		disp += fmt.Sprintf("%s %d", item.Value, item.Count)
	}
	return disp
}
