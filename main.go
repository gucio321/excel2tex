package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/cosiner/flag"
	"github.com/kpango/glg"
	"golang.design/x/clipboard"
)

const (
	fingerprint      = "%%excel2tex%%"
	DefaultTitle     = "XXXXX"
	DefaultSeparator = "c"
)

var commitHash string = "(unknown)"

func init() {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	commitHash = buildInfo.Main.Version
}

func texHeader() string {
	return fmt.Sprintf(`%[1]s Code generated with https://github.com/gucio321/excel2tex %s: %s`,
		fingerprint, commitHash, strings.Join(os.Args, " "))
}

type Row []string

type Table struct {
	Rows            []Row
	LatexColumnType string
	Title           string
	BoldFirstRow    bool
	BoldFirstColumn bool
	LongTable       bool
	NoPreamble      bool
	NoPostamble     bool
	Trim            bool
	Label           string
}

func NewTable() *Table {
	return &Table{
		Rows:         make([]Row, 0),
		Title:        DefaultTitle,
		BoldFirstRow: true,
	}
}

func parseExcelInput(data []byte) (*Table, error) {
	result := NewTable()

	// we can use e.g. csv
	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = '\t'

CSSRead:
	for {
		row, err := reader.Read()
		switch {
		case errors.Is(err, io.EOF):
			break CSSRead
		case err != nil:
			return nil, fmt.Errorf("Unexpected error while parsing excel data: %v", err)
		}

		result.Rows = append(result.Rows, row)
	}

	return result, nil
}

func (t *Table) EncodeLatexTable() string {
	result := t.normalTable()
	if t.LongTable {
		result = t.longTable()
	}

	return strings.ReplaceAll(result, "\n\n", "\n")
}

// normalTable encodes table with table and tabularx
func (t *Table) normalTable() string {
	colTypes := strings.Repeat(fmt.Sprintf("|%s", t.LatexColumnType), len(t.Rows[0]))
	colTypes += "|"

	preamble := fmt.Sprintf(
		`%[1]s
\begin{table}[H]
\caption{%[2]s} %[4]s
\centering
 \begin{tabularx}{\textwidth}{%[3]s}
 \hline `, texHeader(), t.Title, t.colTypesStr(), t.label())

	postamble := `\end{tabularx}
\end{table}`

	if t.NoPreamble {
		preamble = ""
	}

	if t.NoPostamble {
		postamble = ""
	}

	return fmt.Sprintf(
		`%[1]s
		%[2]s %[3]s`,
		preamble, t.mergeRows().String(), postamble,
	)
}

// longTable encodes table with table and tabularx
func (t *Table) longTable() string {
	preamble := fmt.Sprintf(`%[1]s
\begin{longtable}{%[3]s} %% Column alignment and table borders
\caption{%[2]s} %[4]s \\
\hline \endfirsthead %% Header for the first page
\hline \endhead %% Header for subsequent pages
\hline %% Footer for each page`,
		texHeader(),
		t.Title,
		t.colTypesStr(),
		t.label(),
	)

	postamble := `\end{longtable}`

	if t.NoPreamble {
		preamble = ""
	}

	if t.NoPostamble {
		postamble = ""
	}

	return fmt.Sprintf(
		`
%[1]s
%% Table content
%[2]s %[3]s`,
		preamble, t.mergeRows().String(), postamble,
	)
}

// rows returns a copy of t.Rows but with applied various modifiers from Table
func (t *Table) rows() []Row {
	rows := t.Rows

	if t.Trim {
		for i := 0; i < len(rows[0]); i++ {
			cell := rows[0][i]
			if cell == "" {
				for j := 0; j < len(rows); j++ {
					rows[j] = append(rows[j][:i], rows[j][i+1:]...)
				}

				i-- // we lose one column, so we want to recheck the same index
			}
		}
	}

	if t.BoldFirstRow {
		for i, cell := range rows[0] {
			rows[0][i] = "\\textbf{" + cell + "}"
		}
	}

	if t.BoldFirstColumn {
		for i := 0; i < len(rows); i++ {
			rows[i][0] = "\\textbf{" + rows[i][0] + "}"
		}
	}

	return rows
}

func (t *Table) mergeRows() *strings.Builder {
	rowsStr := &strings.Builder{}
	rows := t.rows()

	for _, row := range rows {
		for i, cell := range row {
			fmt.Fprintf(rowsStr, "%s", cell)
			if i < len(row)-1 {
				fmt.Fprint(rowsStr, " & ")
			}
		}

		fmt.Fprintln(rowsStr, " \\\\ \\hline")
	}

	return rowsStr
}

func (t *Table) colTypesStr() string {
	colTypes := strings.Repeat(fmt.Sprintf("|%s", t.LatexColumnType), len(t.Rows[0]))
	colTypes += "|"
	return colTypes
}

func (t *Table) label() string {
	if t.Label == "" {
		return ""
	}

	return fmt.Sprintf("\\label{tbl:%s}", t.Label)
}

type flags struct {
	Title string `names:"-T, --title" usage:"Title of the table" default:"XXXXX"`
	Label string `names:"-l, --label" usage:"Label for the table"`
	Trim  bool   `names:"-t, --trim" usage:"Trim empty columns (useful if you copy only some specified columns e.g. A and C) (NOTE: considers the first (header) row only!)" default:"false"`

	NoFirstRowBold      bool `names:"-nb, --no-header-bold" usage:"Do not bold first row (header)" default:"false"`
	BoldFirstColumn     bool `names:"-bc, --bold-first-column" usage:"Bold first column" default:"false"`
	NoPreamblePostamble bool `names:"-npp, -do --data-only" usage:"Do not generate latex preamble and postamble. Will return only tble body. Ignores title, label, column type, table type e.t.c. Useful for introducing data fixes." default:"false"`

	ColType string `names:"-s" usage:"Separator for table columns (latex table columns type)" default:"c"`

	Legacy bool `name:"--legacy" usage:"Use tabularx instead of longtable. (-s X recomended)" default:"false"`

	Force bool `names:"-f, --force" usage:"Skip any data checks (when possible)." default:"false"`

	Version bool `names:"-v, --version" usage:"Print version and exit" default:"false"`
}

func main() {
	glg.Infof("Welcome to %s %s", glg.Cyan("excel2tex"), glg.Yellow(commitHash))

	var f flags

	flag.ParseStruct(&f)

	glg.Debug("Parsed flags")

	if f.Version {
		fmt.Println(commitHash)
		os.Exit(0)
	}

	if err := clipboard.Init(); err != nil {
		glg.Fatalf("Error while initializing clipboard: %v", err)
	}

	glg.Debug("Clipboard initialized")

	glg.Debug("Reading excel table from clipboard")
	excelTableData := clipboard.Read(clipboard.FmtText)
	glg.Debugf("Got data from clipboard. %d bytes.", len(excelTableData))

	glg.Debug("Validating excel data")
	glg.Debug("Checking, if data aren't actually latex table")
	if !f.Force {
		if strings.HasPrefix(string(excelTableData), fingerprint) {
			glg.Fatalf("Data from clipboard seems to be already latex table (copy again from excel). Use -f to force processing.")
		}
	}

	interFormat, err := parseExcelInput(excelTableData)
	if err != nil {
		glg.Fatalf("Error while parsing excel data: %v", err)
	}

	glg.Debug("Setting properties for internally-processed table")
	interFormat.Title = f.Title
	interFormat.LatexColumnType = f.ColType
	interFormat.LongTable = !f.Legacy
	interFormat.BoldFirstRow = !f.NoFirstRowBold
	interFormat.BoldFirstColumn = f.BoldFirstColumn
	interFormat.NoPreamble = f.NoPreamblePostamble
	interFormat.NoPostamble = f.NoPreamblePostamble
	interFormat.Trim = f.Trim
	interFormat.Label = f.Label

	glg.Debug("Generating latex table")
	latexTable := interFormat.EncodeLatexTable()
	glg.Debug("Writing latex table to clipboard")
	// a small trick here: wait for data to be writen and then exit
	go clipboard.Write(clipboard.FmtText, []byte(latexTable))
	<-clipboard.Watch(context.Background(), clipboard.FmtText)
	glg.Success("Latex table copied to clipboard")
}
