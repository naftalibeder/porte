package encode

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"detaku/lib"
	"detaku/types"

	"golang.org/x/exp/slices"
)

func CopyOrEncodeVideo(fileInfo types.FileInfo, srcPath string, tmpDir string, fileNameNoExt string) (destPath string, err error) {
	if fileInfo.MediaKind != types.Video {
		return "", errors.New("file is not a video")
	}

	// Copy or encode video and audio with ffmpeg.

	vidInfo := fileInfo.VidInfo
	cmdArgs := []string{
		"-i", srcPath,
		"-f", "mp4",
	}

	if !vidInfo.NeedsReEncode {
		cmdArgs = append(cmdArgs,
			"-c:v", "copy",
			"-c:a", "copy",
		)
	} else {
		cmdArgs = append(cmdArgs,
			"-c:v", "libx264",
			"-pix_fmt", "yuv420p",
			"-c:a", "aac",
			"-b:a", "128k",
			"-preset", "slow",
			"-crf", "17",
		)
	}

	destPath = filepath.Join(tmpDir, fileNameNoExt) + ".mp4"
	cmdArgs = append(cmdArgs, destPath)
	cmd := exec.Command((lib.FfmpegBin), cmdArgs...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Join(err, fmt.Errorf(string(out)))
	}

	return destPath, nil
}

// Returns the codecs used to encode the video at srcPath and whether they are
// supported in an mp4 container, as well as the duration.
func GetVidInfo(srcPath string) (vidInfo types.VidInfo, err error) {
	vidInfo = types.VidInfo{}

	out, err := exec.Command(lib.FfprobeBin, srcPath).CombinedOutput()
	if err != nil {
		return types.VidInfo{}, fmt.Errorf("error running ffprobe: %s, %s", out, err)
	}
	info := string(out)

	// Get video codec and mp4 compatibility info.

	m := regexp.MustCompile(`Video: (\w+)\b`)
	VidInfo := m.FindStringSubmatch(info)
	if len(VidInfo) > 0 {
		vidInfo.VidCodec = VidInfo[1]
	}
	vidInfo.IsVidCompat = slices.Contains(
		[]string{
			// "mpeg1video", // Disabled due to an error opening the result in QuickTime.
			"mpeg2video",
			"mpeg4",
			"h264",
		},
		vidInfo.VidCodec,
	)

	// Get audio codec and mp4 compatibility info.

	m = regexp.MustCompile(`Audio: (\w+)\b`)
	audCodecs := m.FindStringSubmatch(info)
	if len(audCodecs) > 0 {
		vidInfo.AudCodec = audCodecs[1]
	}
	vidInfo.IsAudCompat = slices.Contains(
		[]string{
			"h264",
			"aac",
			"mp1",
			"mp2",
			"mp3",
			"qcelp",
			"twinvq",
			"vorbis",
			"alac",
		},
		vidInfo.AudCodec,
	)

	// Determine if file can simply be repackaged.

	if !vidInfo.IsVidCompat || !vidInfo.IsAudCompat {
		vidInfo.NeedsReEncode = true
	}

	// Get video duration.

	args := []string{
		"-v", "quiet",
		"-show_entries",
		"format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		srcPath,
	}
	out, err = exec.Command(lib.FfprobeBin, args...).CombinedOutput()
	if err != nil {
		return types.VidInfo{}, fmt.Errorf("error running ffprobe: %s, %s", out, err)
	}
	durStr := strings.TrimSpace(string(out))
	dur, err := strconv.ParseFloat(durStr, 64)
	if err != nil {
		return types.VidInfo{}, err
	}
	vidInfo.DurationSec = dur

	return vidInfo, nil
}
