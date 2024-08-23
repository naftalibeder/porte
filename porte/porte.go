package porte

import (
	"time"

	"porte/console"
	"porte/lib"
	"porte/log"
)

func Run(srcDir string, destDir string) error {
	// Set up environment.

	logFilePath := log.Start(destDir)
	console.Start()
	totalStart := time.Now()

	// Validate and install dependencies.

	_, err := lib.GetLibs()
	if err != nil {
		return err
	}

	// Analyze all files.

	srcInfo, err := analyzeDir(srcDir)
	if err != nil {
		return err
	}

	// Convert all files.

	err = convertDir(srcInfo, destDir, logFilePath, totalStart)
	if err != nil {
		return err
	}

	return nil
}
