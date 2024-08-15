package main

import (
	"fmt"
	"os"
	"path/filepath"

	"detaku/detaku"
)

func main() {
	srcDir, destDir, err := parseArgs(os.Args)
	if err != nil {
		fmt.Printf("Error validating arguments: %s\n", err)
		return
	}

	err = detaku.Run(srcDir, destDir)
	if err != nil {
		fmt.Printf("Error converting directory: %s\n", err)
		return
	}
}

func parseArgs(args []string) (srcDir string, destDir string, err error) {
	if len(args) < 2 {
		return "", "", fmt.Errorf("not enough arguments (expected `bin srcpath destpath`)")
	}

	srcDir = args[1]
	_, err = os.Stat(srcDir)
	if err != nil {
		return "", "", fmt.Errorf("'%s' does not appear to be a valid source directory", srcDir)
	}

	if len(args) == 3 {
		destDir = args[2]
	} else {
		srcDirEnclosing, srcDirBase := filepath.Split(srcDir)
		destDir = filepath.Join(srcDirEnclosing, fmt.Sprintf("%s_Export", srcDirBase))
	}
	_, err = os.Stat(destDir)
	if err == nil {
		return "", "", fmt.Errorf("destination '%s' already exists", destDir)
	}

	err = os.MkdirAll(destDir, 0777)
	if err != nil {
		return "", "", fmt.Errorf("failed to create export directory in '%s'", destDir)
	}

	return srcDir, destDir, nil
}
