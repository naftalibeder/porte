package types

import "time"

// Files.

type MediaKind = string

const (
	Image   MediaKind = "image"
	Video   MediaKind = "video"
	Unknown MediaKind = "unknown"
)

type FileInfo struct {
	Path      string
	Name      string
	MediaKind MediaKind
	MIMEType  string
	VidInfo   VidInfo
}

// Map of file path to file info.
type FileInfoMap = map[string]FileInfo

// Exif. See https://exiftool.org/TagNames/EXIF.html.

type ExifTags struct {
	Misc  map[string]ExifStrTag
	Dates map[string]ExifDateTag
	Geo   map[string]ExifStrTag
}

type ExifStrTag struct {
	Name  string
	Value string
}

type ExifDateTag struct {
	Name string
	Date time.Time
}

type VidInfo struct {
	VidCodec             string
	IsVidCompat          bool
	AudCodec             string
	IsAudCompat          bool
	CanBeRePackagedInMP4 bool
	DurationSec          float64
}

// Results.

type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeFail    Outcome = "fail"
	OutcomeSkip    Outcome = "skip"
)

// Map of file extension to occurrence count.
type ExtCtMap map[string]int

// Map of video codec to occurrence count.
type VidCodecCtMap map[string]int

// Map of audio codec to occurrence count.
type AudCodecCtMap map[string]int

type ExtCount struct {
	Ext   string
	Count int
}
