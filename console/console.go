package console

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/term"
)

type phase = int

const (
	PhaseCounting       phase = 0
	PhaseAnalyzing      phase = 1
	PhaseConvertingImgs phase = 2
	PhaseConvertingVids phase = 3
	PhaseComplete       phase = 4
)

const (
	headerColWd = 5
)

var (
	writer = bufio.NewWriter(os.Stdout)
	termFD int
	width  int
	output map[phase]([][]string)
)

func Start() {
	if testing.Testing() {
		return
	}

	termFD = int(os.Stdin.Fd())
	output = map[int][][]string{}
}

func Update(phase int, entry [][]string) {
	if testing.Testing() {
		return
	}

	wd, _, err := term.GetSize(termFD)
	if err != nil {
		log.Fatal(err)
	}
	width = wd

	retreat := true

	if phase == PhaseCounting {
		output[PhaseCounting] = entry
	} else if phase == PhaseAnalyzing {
		output[PhaseAnalyzing] = entry
	} else if phase == PhaseConvertingImgs {
		output[PhaseConvertingImgs] = entry
	} else if phase == PhaseConvertingVids {
		output[PhaseConvertingVids] = entry
	} else if phase == PhaseComplete {
		output[PhaseComplete] = entry
		retreat = false
	}

	print(phase, retreat)
}

func print(phase int, retreat bool) {
	checkIfCompleted := func(p int) string {
		if phase >= p {
			return "[x]"
		} else {
			return "[ ]"
		}
	}

	rows := [][]string{}
	rows = append(rows, []string{checkIfCompleted(PhaseCounting), "Counting files"})
	rows = append(rows, output[PhaseCounting]...)
	rows = append(rows, []string{checkIfCompleted(PhaseAnalyzing), "Analyzing files"})
	rows = append(rows, output[PhaseAnalyzing]...)
	rows = append(rows, []string{checkIfCompleted(PhaseConvertingImgs), "Converting images"})
	rows = append(rows, output[PhaseConvertingImgs]...)
	rows = append(rows, []string{checkIfCompleted(PhaseConvertingVids), "Converting videos"})
	rows = append(rows, output[PhaseConvertingVids]...)
	rows = append(rows, []string{checkIfCompleted(PhaseComplete), "Complete"})
	rows = append(rows, output[PhaseComplete]...)

	write(rows, retreat)
}

func write(rows [][]string, retreat bool) {
	lineCt := 0
	widths := []int{headerColWd, width - headerColWd - 1}

	// Functions.

	addNewLn := func() {
		start := "\r"
		clear := "\033[K"
		fmt.Fprint(writer, "\n"+start+clear)
		lineCt++
	}
	addRowDivider := func() {
		fmt.Fprint(writer, "+"+strings.Repeat("-", width-2)+"+")
		addNewLn()
	}
	addCellDivider := func(i int, t string) {
		ct := widths[i] - len(t)
		if ct >= 2 {
			fmt.Fprint(writer, strings.Repeat(" ", ct-1))
		}
	}

	// Content.

	addRowDivider()
	for _, cols := range rows {
		if cols[0] == "-" {
			addRowDivider()
		} else {
			fmt.Fprint(writer, "| ")
			for i, cell := range cols {
				fmt.Fprint(writer, cell)
				addCellDivider(i, cell)
			}
			fmt.Fprint(writer, "|")
			addNewLn()
		}
	}
	addRowDivider()

	// Print and clean up.

	writer.Flush()

	if retreat {
		fmt.Printf("\033[%dA", lineCt)
	}
}

func GetElapsedStr(start time.Time) string {
	return time.Since(start).Truncate(time.Second).String()
}
