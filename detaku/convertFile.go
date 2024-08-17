package detaku

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"detaku/encode"
	"detaku/exif"
	"detaku/log"
	"detaku/types"
	"detaku/utils"
)

func runConvertFileJob(jobs <-chan ConvertFileJob, results chan<- ConvertFileResult) {
	for job := range jobs {
		result := convertFile(job)
		results <- result
	}
}

func convertFile(job ConvertFileJob) (result ConvertFileResult) {
	srcPath := job.SrcPath
	fileInfo := job.FileInfo
	supplFileInfoMap := job.SupplFileInfoMap
	subDirs := job.DestSubDirs

	logEntry := log.LogEntry{}
	absSrcPath, _ := filepath.Abs(srcPath)
	logEntry.SrcPath = absSrcPath
	logEntry.MediaKind = fileInfo.MediaKind
	logEntry.ConvertingStartedAt = time.Now()

	if fileInfo.MediaKind == types.Video {
		logEntry.VidInfo = fileInfo.VidInfo
	}

	defer func() {
		logEntry.ConvertingEndedAt = time.Now()
		logEntry.ConvertingDurationSec = float32(logEntry.ConvertingEndedAt.Sub(logEntry.ConvertingStartedAt).Seconds())
		result.LogEntry = logEntry
	}()

	canSaveFile := true

	// Set up directory structure.

	tmpWorkingDir, err := os.MkdirTemp(subDirs.Tmp, "")
	if err != nil {
		result := ConvertFileResult{
			SrcPath:  srcPath,
			LogEntry: logEntry,
			Err:      err,
		}
		return result
	}
	defer os.RemoveAll(tmpWorkingDir)

	// Parse the file's path components.

	srcNameExt := filepath.Base(srcPath)
	srcExt := filepath.Ext(srcPath)
	srcName := strings.TrimSuffix(srcNameExt, srcExt)

	// Extract all exif tags from the file.

	exifTags, err := exif.GetAllExifTags(srcPath)
	if err != nil {
		logEntry.Errors = append(logEntry.Errors, fmt.Sprintf("Error getting all exif tags: %s", err))

		result := ConvertFileResult{
			SrcPath:  srcPath,
			LogEntry: logEntry,
			Err:      err,
		}
		return result
	}
	logEntry.AllExifTags = exifTags

	// Extract all exif tags from a supplementary file, if available.

	supplFilePath, supplExifTags, err := exif.GetSupplementaryExifTags(srcPath, supplFileInfoMap)
	if err != nil {
		logEntry.Errors = append(logEntry.Errors, fmt.Sprintf("Error getting supplementary exif tags: %s", err))
	}
	logEntry.SupplFilePath = supplFilePath
	logEntry.SupplExifTags = supplExifTags

	// Find the earliest available date tag.

	earliestDateTag := types.ExifDateTag{
		Name: "",
		Date: time.Now(),
	}
	foundDate := false
	dateTags := []types.ExifDateTag{}
	for _, t := range exifTags.Dates {
		// Avoid using exif date tags that describe system activity, like
		// FileModificationDateTime.
		if !strings.Contains(t.Name, "File") {
			dateTags = append(dateTags, t)
		}
	}
	for _, t := range supplExifTags.Dates {
		dateTags = append(dateTags, t)
	}
	if len(dateTags) > 0 {
		earliestDateTag = dateTags[0]
		for _, t := range dateTags {
			if t.Date.Before(earliestDateTag.Date) {
				earliestDateTag = t
			}
		}
		foundDate = true
		logEntry.DateSrc = log.DateSrcExifTag
		logEntry.DateSrcExifTagName = earliestDateTag.Name
	}

	if !foundDate {
		searchStr := fileInfo.Name + " " + supplExifTags.Misc["ImageTitle"].Value
		dateFromTitle, err := utils.GetDateFromStr(searchStr)
		if err == nil {
			// If the filename or image title indicates an earlier date than any other,
			// use it. This could mean an older photo was digitized at a later date, causing
			// the exif data to be incorrect.
			earliestDateTag = types.ExifDateTag{
				Name: "CreateDate",
				Date: dateFromTitle,
			}
			foundDate = true
			logEntry.DateSrc = log.DateSrcImgTitle
			logEntry.DateSrcImgTitleSearchStr = searchStr
		}
	}

	if !foundDate {
		canSaveFile = false
		logEntry.Errors = append(logEntry.Errors, "No earliest date found in file, supplementary file, or filename")
	} else {
		logEntry.UsedDateTag = earliestDateTag
	}

	// Find all geo tags.

	geoTags := []types.ExifStrTag{}
	for _, t := range exifTags.Geo {
		geoTags = append(geoTags, t)
	}
	for _, t := range supplExifTags.Geo {
		geoTags = append(geoTags, t)
	}

	// Set the input filename as the title tag to preserve it (since the output filename
	// will have a datestamp before the original title).

	title := srcName

	// Copy from the source file to a temporary file to begin safely making modifications.

	tmpPath := filepath.Join(tmpWorkingDir, "1") + srcExt
	_, err = exec.Command("cp", srcPath, tmpPath).CombinedOutput()
	if err != nil {
		logEntry.Errors = append(logEntry.Errors, err.Error())

		result := ConvertFileResult{
			SrcPath:  srcPath,
			LogEntry: logEntry,
			Err:      err,
		}
		return result
	}

	// Fix an incorrect extension, like a jpg named cat.png, by first duplicating the
	// file with a corrected extension.

	tmpPathNext := ""
	fixedExt := exif.GetExifFileExt(exifTags.Misc, srcExt)
	if srcExt != fixedExt {
		tmpPathNext = filepath.Join(tmpWorkingDir, "2") + fixedExt
		out, err := exec.Command("cp", tmpPath, tmpPathNext).CombinedOutput()
		if err != nil {
			logEntry.Errors = append(logEntry.Errors, fmt.Sprintf("Error copying file with new extension: %s", string(out)))

			result := ConvertFileResult{
				SrcPath:  srcPath,
				LogEntry: logEntry,
				Err:      err,
			}
			return result
		}
	}
	if tmpPathNext != "" {
		tmpPath = tmpPathNext
	}

	// Normalize a video by repackaging or copying.

	tmpPathNext = ""
	if fileInfo.MediaKind == types.Video {
		tmpPathNext, err = encode.CopyVideo(fileInfo, tmpPath, tmpWorkingDir, "3")
		if err != nil {
			canSaveFile = false
			logEntry.Errors = append(logEntry.Errors, fmt.Sprintf("Error copying or encoding video: %s", err))
		}
	}
	if tmpPathNext != "" {
		tmpPath = tmpPathNext
	}

	// Set the file's tags.

	tmpPathNext = ""
	if canSaveFile {
		tmpPathNext = filepath.Join(tmpWorkingDir, "4"+filepath.Ext(tmpPath))
		tagsArg := exif.SetExifTagsArg{
			TagsPath: srcPath,
			Title:    title,
			Date:     earliestDateTag.Date,
			Geo:      geoTags,
		}
		err = exif.SetExifTags(tmpPath, tmpPathNext, tagsArg)
		if err != nil {
			canSaveFile = false
			logEntry.Errors = append(logEntry.Errors, fmt.Sprintf("Error setting exif tags: %s", err))
		}
	}
	if tmpPathNext != "" {
		tmpPath = tmpPathNext
	}

	// Write the file to the success or fail directory with the appropriate name.

	copyFromPath := ""
	copyToPath := ""
	outcome := types.OutcomeSuccess

	if canSaveFile {
		copyFromPath = tmpPath
		destFileName := earliestDateTag.Date.Format(utils.FileNameFmt) + utils.FileNamePartSep + srcName + filepath.Ext(tmpPath)
		copyToPath = utils.GetValidPath(subDirs.Success, destFileName)
		outcome = types.OutcomeSuccess
	} else {
		copyFromPath = srcPath
		destFileName := srcNameExt
		copyToPath = utils.GetValidPath(subDirs.Fail, destFileName)
		outcome = types.OutcomeFail
	}

	cmd := exec.Command("cp", copyFromPath, copyToPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logEntry.Errors = append(logEntry.Errors, fmt.Sprintf("Error copying file to final directory: %s, %s", err.Error(), string(out)))
	}

	// Add to html entry.

	absDestPath, _ := filepath.Abs(copyToPath)
	logEntry.DestPath = absDestPath
	logEntry.Outcome = outcome

	result = ConvertFileResult{
		SrcPath:  srcPath,
		LogEntry: logEntry,
		Err:      nil,
	}
	return result
}
