package lib

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	ExiftoolBin = ""
	FfmpegBin   = ""
	FfprobeBin  = ""

	libCacheDir = ""
)

//go:embed install.sh
var InstallScript string

func GetLibs() (out string, err error) {
	globalCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	libCacheDir = filepath.Join(globalCacheDir, "porte", "lib")
	_ = os.MkdirAll(libCacheDir, 0777)

	ExiftoolDir := filepath.Join(libCacheDir, "exiftool")
	ExiftoolBin = filepath.Join(libCacheDir, "exiftool/exiftool")
	FfmpegDir := filepath.Join(libCacheDir, "ffmpeg")
	FfmpegBin = filepath.Join(libCacheDir, "ffmpeg/ffmpeg")
	FfprobeDir := filepath.Join(libCacheDir, "ffprobe")
	FfprobeBin = filepath.Join(libCacheDir, "ffprobe/ffprobe")

	cmd := exec.Command("sh", "-c", InstallScript)
	cmd.Env = append(cmd.Env,
		"CACHE_DIR="+libCacheDir,
		"EXIFTOOL_DIR="+ExiftoolDir,
		"EXIFTOOL_BIN="+ExiftoolBin,
		"FFMPEG_DIR="+FfmpegDir,
		"FFMPEG_BIN="+FfmpegBin,
		"FFPROBE_DIR="+FfprobeDir,
		"FFPROBE_BIN="+FfprobeBin,
	)
	outBt, err := cmd.CombinedOutput()
	out = string(outBt)
	if err != nil {
		return out, fmt.Errorf("error installing dependencies: %s, %s", out, err)
	}

	return out, nil
}
