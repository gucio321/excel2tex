package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/kpango/glg"
	"golang.design/x/clipboard"
)

const (
	DefaultTitle     = "XXXXX"
	DefaultSeparator = "X"
)

type Row []string

type Table struct {
	Rows            []Row
	LatexColumnType string
	Title           string
	BoldFirstRow    bool
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

func (t *Table) encodeLatexTable() string {
	rows := t.Rows
	if t.BoldFirstRow {
		for i, cell := range rows[0] {
			rows[0][i] = "\\textbf{" + cell + "}"
		}
	}

	rowsStr := &strings.Builder{}
	for _, row := range rows {
		for i, cell := range row {
			fmt.Fprintf(rowsStr, "%s", cell)
			if i < len(row)-1 {
				fmt.Fprint(rowsStr, " & ")
			}
		}

		fmt.Fprintln(rowsStr, " \\\\ \\hline")
	}

	colTypes := strings.Repeat(fmt.Sprintf("|%s", t.LatexColumnType), len(rows[0]))
	colTypes += "|"

	return fmt.Sprintf(
		`\begin{table}[ht]
\caption{%[1]s}
\centering
 \begin{tabularx}{\textwidth}{%[3]s}
 \hline 
%[2]s
\end{tabularx}
\end{table}`,
		t.Title,
		rowsStr.String(),
		colTypes,
	)
}

func main() {
	glg.Infof("Welcome to %s", glg.Cyan("excel2tex"))

	title := flag.String("t", DefaultTitle, "Title of the table")
	colType := flag.String("s", DefaultSeparator, "Separator for table columns (latex table columns type)")
	flag.Parse()
	glg.Debug("Parsed flags")

	if err := clipboard.Init(); err != nil {
		glg.Fatalf("Error while initializing clipboard: %v", err)
	}

	glg.Debug("Clipboard initialized")

	glg.Debug("Reading excel table from clipboard")
	excelTableData := clipboard.Read(clipboard.FmtText)
	glg.Debugf("Got data from clipboard. %d bytes.", len(excelTableData))

	interFormat, err := parseExcelInput(excelTableData)
	if err != nil {
		glg.Fatalf("Error while parsing excel data: %v", err)
	}

	glg.Debug("Setting properties for internally-processed table")
	interFormat.Title = *title
	interFormat.LatexColumnType = *colType

	glg.Debug("Generating latex table")
	latexTable := interFormat.encodeLatexTable()
	glg.Debug("Writing latex table to clipboard")
	// a small trick here: wait for data to be writen and then exit
	go clipboard.Write(clipboard.FmtText, []byte(latexTable))
	<-clipboard.Watch(context.Background(), clipboard.FmtText)
	glg.Success("Latex table copied to clipboard")
}
