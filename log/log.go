package log

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"porte/types"
)

type LogOutput struct {
	Entries []LogEntry
}

type DateSrc = string

const (
	DateSrcExifTag  DateSrc = "exifTag"
	DateSrcImgTitle DateSrc = "fileName"
)

type LogEntry struct {
	SrcPath                  string
	DestPath                 string
	SupplFilePath            string
	Outcome                  types.Outcome
	ConvertingStartedAt      time.Time
	ConvertingEndedAt        time.Time
	ConvertingDurationSec    float32
	MediaKind                types.MediaKind
	DateSrc                  DateSrc
	DateSrcExifTagName       string
	DateSrcImgTitleSearchStr string
	UsedDateTag              types.ExifDateTag
	AllExifTags              types.ExifTags
	SupplExifTags            types.ExifTags
	VidInfo                  types.VidInfo
	Errors                   []string
}

type PrettyLogEntry struct {
	LogEntry
	AllExifTags   map[string]any
	SupplExifTags map[string]any
}

var (
	outFilePath string
	output      LogOutput
)

func Start(dir string) string {
	outFilePath = filepath.Join(dir, "log.json")
	output = LogOutput{}
	return outFilePath
}

func AddEntry(entry LogEntry) {
	output.Entries = append(output.Entries, entry)
	write()
}

func write() error {
	prettyEntries := []PrettyLogEntry{}
	for _, e := range output.Entries {
		allExifTags := map[string]any{}
		for n, t := range e.AllExifTags.Misc {
			allExifTags[n] = t.Value
		}
		for n, t := range e.AllExifTags.Dates {
			allExifTags[n] = t.Date
		}
		for n, t := range e.AllExifTags.Geo {
			allExifTags[n] = t.Value
		}

		supplExifTags := map[string]any{}
		for n, t := range e.SupplExifTags.Misc {
			supplExifTags[n] = t.Value
		}
		for n, t := range e.SupplExifTags.Dates {
			supplExifTags[n] = t.Date
		}
		for n, t := range e.SupplExifTags.Geo {
			supplExifTags[n] = t.Value
		}

		prettyEntry := PrettyLogEntry{
			LogEntry:      e,
			AllExifTags:   allExifTags,
			SupplExifTags: supplExifTags,
		}
		prettyEntries = append(prettyEntries, prettyEntry)
	}

	bt, err := json.MarshalIndent(prettyEntries, "", "  ")
	if err != nil {
		return err
	}

	outPath := filepath.Join(outFilePath)
	err = os.WriteFile(outPath, bt, 0644)
	if err != nil {
		return err
	}

	return nil
}
