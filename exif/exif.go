package exif

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"porte/lib"
	"porte/types"
	"porte/utils"
)

type googleInfo struct {
	Title        string
	Description  string
	ImageViews   string
	CreationTime struct {
		Timestamp string
		Formatted string
	}
	PhotoTakenTime struct {
		Timestamp string
		Formatted string
	}
	PhotoLastModifiedTime struct {
		Timestamp string
		Formatted string
	}
	GeoData struct {
		Latitude      float32
		Longitude     float32
		Altitude      float32
		LatitudeSpan  float32
		LongitudeSpan float32
	}
	GeoDataExif struct {
		Latitude      float32
		Longitude     float32
		Altitude      float32
		LatitudeSpan  float32
		LongitudeSpan float32
	}
	People [](struct {
		name string
	})
	Url                string
	GooglePhotosOrigin struct {
		PhotosDesktopUploader struct {
		}
	}
}

// Returns all exif tags from the file at srcPath.
func GetAllExifTags(srcPath string) (tags types.ExifTags, err error) {
	tags = types.ExifTags{
		Misc:  map[string]types.ExifStrTag{},
		Dates: map[string]types.ExifDateTag{},
		Geo:   map[string]types.ExifStrTag{},
	}

	// Get misc tags.

	cmd := exec.Command(
		lib.ExiftoolBin,
		"-api", "TimeZone=UTC",
		"-d", utils.ExifToolDateFmt,
		"-c", "%.6f",
		"-j",
		srcPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return types.ExifTags{}, err
	}

	rawExifTagMap := parseJSONResponse(out)
	for n, t := range rawExifTagMap {
		tags.Misc[n] = types.ExifStrTag{
			Name:  n,
			Value: fmt.Sprint(t),
		}
	}

	// Get date tags.

	cmdArgs := []string{
		"-api", "TimeZone=UTC",
		"-time:all",
		"-d", utils.ExifToolDateFmt,
		"-c", "%.6f",
		"-j",
		srcPath,
	}
	cmd = exec.Command(lib.ExiftoolBin, cmdArgs...)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return types.ExifTags{}, err
	}

	rawExifTagMap = parseJSONResponse(out)
	for n, t := range rawExifTagMap {
		v := fmt.Sprint(t)
		d, err := time.Parse(utils.GoParseExifToolDateFmt, v)
		if err != nil {
			continue
		}

		if d.Unix() == 0 {
			continue
		}

		// Avoid using tags that describe system activity, like
		// FileModificationDateTime.
		if strings.Contains(n, "File") {
			continue
		}

		// Avoid using tags that are just standalone times without a date component.
		if !strings.Contains(n, "Date") {
			continue
		}

		// Avoid using tags that describe dates unrelated to when the file was
		// captured, like the creation date of the color profile.
		if n == "ProfileDateTime" {
			continue
		}

		tags.Dates[n] = types.ExifDateTag{
			Name: n,
			Date: d,
		}
	}

	// Get geo tags.

	cmd = exec.Command(lib.ExiftoolBin, "-a", "-gps:all", "-c", "%.6f", "-j", srcPath)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return types.ExifTags{}, err
	}

	rawExifTagMap = parseJSONResponse(out)
	for n, t := range rawExifTagMap {
		tags.Geo[n] = types.ExifStrTag{
			Name:  n,
			Value: fmt.Sprint(t),
		}
	}

	return tags, err
}

// Returns the mime type from the exif data of the file at srcPath.
func GetExifMimeType(srcPath string) (mimeType string, err error) {
	cmd := exec.Command(lib.ExiftoolBin, "-MIMEType", "-j", srcPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// JSON return value from exiftool of the shape [{ "TagName": TagValue }].
	var data [](map[string]string)
	err = json.Unmarshal(out, &data)
	if err != nil {
		return "", err
	}
	rawExifTagMap := data[0] // {exiftool name: value}

	mimeType = rawExifTagMap["MIMEType"]
	return mimeType, nil
}

// Returns a valid extension (including the . prefix) based on the exif data in tags,
// or the original extension if a different one isn't found.
func GetExifFileExt(miscTags map[string]types.ExifStrTag, ext string) string {
	t, exists := miscTags["FileTypeExtension"]
	if !exists {
		return ext
	}

	return fmt.Sprintf(".%s", t.Value)
}

// Finds the file in supplFileInfoMap most likely to match the file at srcPath. If one
// exists, returns the file's path and the exif tags it contains.
func GetSupplementaryExifTags(srcPath string, supplFileInfoMap types.FileInfoMap) (filePath string, tags types.ExifTags, err error) {
	// Try to find a corresponding Google json file.

	filePath, err = utils.GetSupplementaryFilePath(srcPath, supplFileInfoMap)
	if err != nil {
		return "", types.ExifTags{}, err
	}

	jsonBt, err := os.ReadFile(filePath)
	if err != nil {
		return "", types.ExifTags{}, err
	}

	// Read the json file.

	var googleInfo *googleInfo
	err = json.Unmarshal(jsonBt, &googleInfo)
	if err != nil {
		return "", types.ExifTags{}, err
	}
	if googleInfo == nil {
		return "", types.ExifTags{}, errors.New("supplementary file is empty")
	}

	// Build a return payload.

	tags = types.ExifTags{
		Misc: map[string]types.ExifStrTag{
			"ImageTitle": {
				Name:  "ImageTitle",
				Value: googleInfo.Title,
			},
		},
		Dates: map[string]types.ExifDateTag{
			"DateTime": {
				Name: "DateTime",
				Date: unixToDate(googleInfo.PhotoLastModifiedTime.Timestamp),
			},
			"DateTimeOriginal": {
				Name: "DateTimeOriginal",
				Date: unixToDate(googleInfo.PhotoTakenTime.Timestamp),
			},
			"DateTimeDigitized": {
				Name: "DateTimeDigitized",
				Date: unixToDate(googleInfo.PhotoTakenTime.Timestamp),
			},
		},
		Geo: map[string]types.ExifStrTag{
			"GPSLatitude": {
				Name:  "GPSLatitude",
				Value: fmt.Sprint(googleInfo.GeoData.Latitude),
			},
			"GPSLongitude": {
				Name:  "GPSLongitude",
				Value: fmt.Sprint(googleInfo.GeoData.Longitude),
			},
			"GPSAltitude": {
				Name:  "GPSAltitude",
				Value: fmt.Sprint(googleInfo.GeoData.Altitude),
			},
		},
	}

	// Clean up bad values.

	for n, t := range tags.Dates {
		if t.Date.Unix() == 0 {
			delete(tags.Dates, n)
		}
	}
	for n, t := range tags.Geo {
		if t.Value == "" {
			delete(tags.Geo, n)
		}
		f, err := strconv.ParseFloat(t.Value, 32)
		if err != nil {
			continue
		}
		if f == 0 {
			delete(tags.Geo, n)
		}
	}

	// Report.

	return filePath, tags, nil
}

type SetExifTagsArg struct {
	// The path to a source file from which all tags should be copied.
	TagsPath string
	Title    string
	Date     time.Time
	Geo      []types.ExifStrTag
}

// Copies the file at srcPath to a new file at destPath, copying all tags from the
// file at tags.TagsPath and then setting all date-related exif tags to tags.Date,
// the title tag to tags.Title, etc.
func SetExifTags(srcPath string, destPath string, tags SetExifTagsArg) error {
	cmdArgs := []string{}
	cmdArgs = append(cmdArgs, "-TagsFromFile", tags.TagsPath)
	cmdArgs = append(cmdArgs, fmt.Sprintf("-Title=%s", tags.Title))
	cmdArgs = append(cmdArgs, fmt.Sprintf("-AllDates=%s", tags.Date.Format(utils.GoParseExifToolDateFmt)))
	for _, t := range tags.Geo {
		cmdArgs = append(cmdArgs, fmt.Sprintf("-%s=%s", t.Name, t.Value))
	}
	cmdArgs = append(cmdArgs, "-o", destPath, srcPath)

	cmd := exec.Command(lib.ExiftoolBin, cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Join(err, fmt.Errorf(string(out)))
	}

	return nil
}

// Parses the json response from exiftool and returns a map of the shape
// {exiftool name: value}. Errors are not handled, in order to return a
// map, even if empty.
func parseJSONResponse(out []byte) map[string]interface{} {
	// JSON return value from exiftool of the shape [{ "TagName": TagValue }].
	var data [](map[string]interface{})
	err := json.Unmarshal(out, &data)
	if err != nil {
		return map[string]interface{}{}
	}
	if len(data) == 0 {
		return map[string]interface{}{}
	}

	m := data[0]
	delete(m, "SourceFile")
	return m
}

func unixToDate(u string) time.Time {
	uInt, _ := strconv.ParseInt(u, 10, 64)
	t := time.Unix(uInt, 0).UTC()
	return t
}
