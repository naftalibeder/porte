package porte

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"porte/console"
	"porte/log"
	"porte/types"
)

type ConvertFileJob struct {
	SrcPath          string
	FileInfo         types.FileInfo
	SupplFileInfoMap types.FileInfoMap
	DestSubDirs      ConvertDestSubDirs
}

type ConvertFileResult struct {
	SrcPath  string
	LogEntry log.LogEntry
	Err      error
}

type ConvertDestSubDirs struct {
	Tmp     string
	Success string
	Fail    string
}

func convertDir(srcInfo AnalyzeDirResult, destDir string, logFilePath string, totalStart time.Time) error {
	// Set up destination directory structure.

	destSubDirs := ConvertDestSubDirs{
		Tmp:     filepath.Join(destDir, ".tmp"),
		Success: filepath.Join(destDir, "success"),
		Fail:    filepath.Join(destDir, "fail"),
	}

	if err := os.MkdirAll(destSubDirs.Tmp, 0777); err != nil {
		return err
	}
	if err := os.MkdirAll(destSubDirs.Success, 0777); err != nil {
		return err
	}
	if err := os.MkdirAll(destSubDirs.Fail, 0777); err != nil {
		return err
	}

	defer os.RemoveAll(destSubDirs.Tmp)

	// Convert all images and videos in the source directory.

	err := convertSubPhase(
		srcInfo.ImgFileInfoMap,
		srcInfo.SupplFileInfoMap,
		destSubDirs,
		console.PhaseConvertingImgs,
	)
	if err != nil {
		return err
	}

	err = convertSubPhase(
		srcInfo.VidFileInfoMap,
		srcInfo.SupplFileInfoMap,
		destSubDirs,
		console.PhaseConvertingVids,
	)
	if err != nil {
		return err
	}

	// Tell us about it.

	console.Update(console.PhaseComplete, [][]string{
		{"", fmt.Sprintf("- Files exported to '%s'", destDir)},
		{"", fmt.Sprintf("- Log saved to '%s'", logFilePath)},
		{"", "- " + console.GetElapsedStr(totalStart) + " total elapsed"},
	})

	return nil
}

func convertSubPhase(mediaFileInfoMap types.FileInfoMap, supplFileInfoMap types.FileInfoMap, destSubDirs ConvertDestSubDirs, consolePhase int) error {
	progressCt := 0
	totalCt := len(mediaFileInfoMap)
	successCt := 0
	failCt := 0

	if totalCt == 0 {
		console.Update(consolePhase, [][]string{
			{"", "- 0 files"},
		})
		return nil
	}

	// Set up worker pool to handle file analysis jobs.

	jobCt := len(mediaFileInfoMap)
	jobs := make(chan ConvertFileJob, jobCt)
	results := make(chan ConvertFileResult, jobCt)

	// Initialize all workers.
	const workerCt = 10
	for i := 0; i < workerCt; i++ {
		go runConvertFileJob(jobs, results)
	}
	defer close(jobs)

	// Populate jobs.
	for path, fileInfo := range mediaFileInfoMap {
		job := ConvertFileJob{
			SrcPath:          path,
			FileInfo:         fileInfo,
			SupplFileInfoMap: supplFileInfoMap,
			DestSubDirs:      destSubDirs,
		}
		jobs <- job
	}

	// Read results from file conversions.

	sectionStart := time.Now()
	console.Update(consolePhase, [][]string{
		{"", "- Starting..."},
		{"", "- " + console.GetElapsedStr(sectionStart) + " elapsed"},
	})

	for i := 0; i < jobCt; i++ {
		result := <-results
		progressCt++

		console.Update(consolePhase, [][]string{
			{"", fmt.Sprintf("- '%s'", result.SrcPath)},
			{"", fmt.Sprintf("- Converting %d of %d", progressCt, totalCt)},
			{"", fmt.Sprintf("- %d success, %d fail", successCt, failCt)},
			{"", "- " + console.GetElapsedStr(sectionStart) + " elapsed"},
		})
		log.AddEntry(result.LogEntry)

		if result.LogEntry.Outcome == types.OutcomeSuccess {
			successCt++
		} else if result.LogEntry.Outcome == types.OutcomeFail {
			failCt++
		}
	}

	return nil
}
