package detaku

import (
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"detaku/console"
	"detaku/types"
	"detaku/utils"
)

type AnalyzeDirResult struct {
	ImgFileInfoMap   types.FileInfoMap
	VidFileInfoMap   types.FileInfoMap
	SupplFileInfoMap types.FileInfoMap
}

type AnalyzeFileJob struct {
	Path string
}

type AnalyzeFileResult struct {
	Path          string
	MediaKind     types.MediaKind
	MediaFileInfo types.FileInfo
	SupplFileInfo types.FileInfo
	Ext           string
	Err           error
}

func analyzeDir(srcDir string) (AnalyzeDirResult, error) {
	// Get total file count.

	sectionStart := time.Now()
	console.Update(console.PhaseCounting, [][]string{
		{"", "- In progress (may take several minutes)..."},
	})

	out, err := exec.Command("bash", "-c", fmt.Sprintf("find '%s' -type f | wc -l", srcDir)).CombinedOutput()
	if err != nil {
		return AnalyzeDirResult{}, fmt.Errorf("failed to get file count: %s, %s", string(out), err)
	}
	totalFileCt, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return AnalyzeDirResult{}, fmt.Errorf("failed to convert file count to int: %s", err)
	}

	console.Update(console.PhaseCounting, [][]string{
		{"", fmt.Sprintf("- Found %d files", totalFileCt)},
		{"", "- " + console.GetElapsedStr(sectionStart) + " elapsed"},
	})

	// Get a more precise list of usable files.

	usableFilesMap := map[string]bool{}
	_ = filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fileName := d.Name()
		if strings.HasPrefix(fileName, ".") {
			return nil
		}

		usableFilesMap[path] = true
		return nil
	})

	// Set up worker pool to handle file analysis jobs.

	jobCt := len(usableFilesMap)
	jobs := make(chan AnalyzeFileJob, jobCt)
	results := make(chan AnalyzeFileResult, jobCt)

	// Initialize all workers.
	const workerCt = 10
	for i := 0; i < workerCt; i++ {
		go runAnalyzeFileJob(jobs, results)
	}
	defer close(jobs)

	// Populate jobs.
	for path := range usableFilesMap {
		job := AnalyzeFileJob{
			Path: path,
		}
		jobs <- job
	}

	// Read results from file analysis.

	sectionStart = time.Now()

	imgFileInfoMap := types.FileInfoMap{}
	imgExtCtMap := types.ExtCtMap{}

	vidFileInfoMap := types.FileInfoMap{}
	vidExtCtMap := types.ExtCtMap{}
	vidTotalDurationSec := 0
	vidTotalDuration := ""
	vidsNeedReEncodeCt := 0

	// A lookup table recording the existence of each supplementary json file.
	// Each file is keyed by its full path.
	var supplFileInfoMap = types.FileInfoMap{}

	walkedFileCt := 0
	for i := 0; i < jobCt; i++ {
		result := <-results
		walkedFileCt++

		// Add result to counter maps.

		if result.MediaKind == types.Image {
			imgFileInfoMap[result.Path] = result.MediaFileInfo
			imgExtCtMap[result.Ext]++
		} else if result.MediaKind == types.Video {
			vidFileInfoMap[result.Path] = result.MediaFileInfo
			vidExtCtMap[result.Ext]++
			vidTotalDurationSec += int(result.MediaFileInfo.VidInfo.DurationSec)
			dur, _ := time.ParseDuration(fmt.Sprintf("%ds", vidTotalDurationSec))
			vidTotalDuration = dur.String()
			if result.MediaFileInfo.VidInfo.NeedsReEncode {
				vidsNeedReEncodeCt++
			}
		}
		supplFileInfoMap[result.Path] = result.SupplFileInfo

		// Tell us about it.

		imgExtsSorted := utils.SortedListFromCt(imgExtCtMap)
		imgExtsDisp := ""
		if len(imgExtCtMap) > 0 {
			imgExtsDisp = fmt.Sprintf("(%s)", imgExtsSorted)
		}

		vidExtsSorted := utils.SortedListFromCt(vidExtCtMap)
		vidExtsDisp := ""
		if len(vidExtCtMap) > 0 {
			vidExtsDisp = fmt.Sprintf("(%s)", vidExtsSorted)
		}

		console.Update(console.PhaseAnalyzing, [][]string{
			{"", fmt.Sprintf("- Analyzed %d/%d files", walkedFileCt, totalFileCt)},
			{"", fmt.Sprintf("- Images: %d %s", len(imgFileInfoMap), imgExtsDisp)},
			{"", fmt.Sprintf("- Videos: %d %s (%s, %d need re-encoding)", len(vidFileInfoMap), vidExtsDisp, vidTotalDuration, vidsNeedReEncodeCt)},
			{"", fmt.Sprintf("- Supplementary files: %d", len(supplFileInfoMap))},
			{"", "- " + console.GetElapsedStr(sectionStart) + " elapsed"},
		})
	}

	result := AnalyzeDirResult{
		ImgFileInfoMap:   imgFileInfoMap,
		VidFileInfoMap:   vidFileInfoMap,
		SupplFileInfoMap: supplFileInfoMap,
	}
	return result, nil
}
