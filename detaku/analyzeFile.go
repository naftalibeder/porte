package detaku

import (
	"path/filepath"
	"strings"

	"detaku/encode"
	"detaku/exif"
	"detaku/types"
)

func runAnalyzeFileJob(jobs <-chan AnalyzeFileJob, results chan<- AnalyzeFileResult) {
	for job := range jobs {
		result := analyzeFile(job)
		results <- result
	}
}

func analyzeFile(job AnalyzeFileJob) (result AnalyzeFileResult) {
	path := job.Path

	name := filepath.Base(path)

	var mediaKind types.MediaKind = types.Unknown
	var mediaFileInfo types.FileInfo
	var supplFileInfo types.FileInfo
	var ext string

	mimeType, _ := exif.GetExifMimeType(path)
	if strings.HasPrefix(mimeType, "image") {
		mediaKind = types.Image
	} else if strings.HasPrefix(mimeType, "video") {
		mediaKind = types.Video
	}

	ext = strings.ToLower(filepath.Ext(name))
	if mediaKind == types.Image {
		mediaFileInfo = types.FileInfo{
			Path:      path,
			Name:      name,
			MediaKind: mediaKind,
			MIMEType:  mimeType,
		}
	} else if mediaKind == types.Video {
		vidInfo, err := encode.GetVidInfo(path)
		if err != nil {
			result := AnalyzeFileResult{
				Path: path,
				Err:  err,
			}
			return result
		}
		mediaFileInfo = types.FileInfo{
			Path:      path,
			Name:      name,
			MediaKind: mediaKind,
			MIMEType:  mimeType,
			VidInfo:   vidInfo,
		}
	} else {
		supplFileInfo = types.FileInfo{
			Path:      path,
			Name:      name,
			MediaKind: mediaKind,
		}
	}

	result = AnalyzeFileResult{
		Path:          path,
		MediaKind:     mediaKind,
		MediaFileInfo: mediaFileInfo,
		SupplFileInfo: supplFileInfo,
		Ext:           ext,
	}
	return result
}
