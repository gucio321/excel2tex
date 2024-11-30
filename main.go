package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kpango/glg"
	"golang.design/x/clipboard"
)

const fingerprint = "%%excel2tex%%"

var commitHash string = "(unknown)"

func texHeader() string {
	return fmt.Sprintf(`%[1]s Code generated with https://github.com/gucio321/excel2tex %s: %s`,
		fingerprint, commitHash, strings.Join(os.Args, " "))
}

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
	BoldFirstColumn bool
	LongTable       bool
	NoPreamble      bool
	NoPostamble     bool
	Trim            bool
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
	if t.LongTable {
		return t.longTable()
	}

	return t.normalTable()
}

// normalTable encodes table with table and tabularx
func (t *Table) normalTable() string {
	colTypes := strings.Repeat(fmt.Sprintf("|%s", t.LatexColumnType), len(t.Rows[0]))
	colTypes += "|"

	preamble := fmt.Sprintf(
		`%[1]s
\begin{table}[H]
\caption{%[2]s}
\centering
 \begin{tabularx}{\textwidth}{%[3]s}
 \hline 
 `, texHeader(), t.Title, t.colTypesStr())

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
%[2]s
%[3]s`,
		preamble, t.mergeRows().String(), postamble,
	)
}

// longTable encodes table with table and tabularx
func (t *Table) longTable() string {
	preamble := fmt.Sprintf(`%[1]s
\begin{longtable}{%[3]s} %% Column alignment and table borders
\caption{%[2]s} \\

%% Header for the first page
\hline
%%\multicolumn{3}{|c|}{Table Header} \\
%%\hline
\endfirsthead

%% Header for subsequent pages
\hline
%%\multicolumn{3}{|c|}{Table Header (continued)} \\
%%\hline
\endhead

%% Footer for each page
\hline
%%\endfoot

%% Footer for the last page
%%\hline
%%\endlastfoot
`, texHeader(), t.Title, t.colTypesStr())

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
%[2]s
%[3]s`,
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

func main() {
	glg.Infof("Welcome to %s %s", glg.Cyan("excel2tex"), glg.Yellow(commitHash))

	title := flag.String("t", DefaultTitle, "Title of the table")
	colType := flag.String("s", DefaultSeparator, "Separator for table columns (latex table columns type)")
	long := flag.Bool("long", false, "Use longtable instead of table and tabularx (recomended -s c)")
	noFirstRowBold := flag.Bool("nb", false, "Do not bold first row.")
	boldFirstColumn := flag.Bool("bc", false, "Bold first column.")
	noPreamblePostamble := flag.Bool("npp", false, "Do not generate latex preamble and postamble. Will return only tble body. Ignores title. Useful to replace only the table body.")
	trim := flag.Bool("trim", false, "Trim empty columns (useful if you copy only some specified columns e.g. A and C) (NOTE: considers the first (header) row!)")
	force := flag.Bool("f", false, "Skip any data checks (when possible).")
	version := flag.Bool("v", false, "Print version and exit")

	flag.Parse()

	glg.Debug("Parsed flags")

	if *version {
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
	if !*force {
		if strings.HasPrefix(string(excelTableData), fingerprint) {
			glg.Fatalf("Data from clipboard seems to be already latex table (copy again from excel). Use -f to force processing.")
		}
	}

	interFormat, err := parseExcelInput(excelTableData)
	if err != nil {
		glg.Fatalf("Error while parsing excel data: %v", err)
	}

	glg.Debug("Setting properties for internally-processed table")
	interFormat.Title = *title
	interFormat.LatexColumnType = *colType
	interFormat.LongTable = *long
	interFormat.BoldFirstRow = !*noFirstRowBold
	interFormat.BoldFirstColumn = *boldFirstColumn
	interFormat.NoPreamble = *noPreamblePostamble
	interFormat.NoPostamble = *noPreamblePostamble
	interFormat.Trim = *trim

	glg.Debug("Generating latex table")
	latexTable := interFormat.EncodeLatexTable()
	glg.Debug("Writing latex table to clipboard")
	// a small trick here: wait for data to be writen and then exit
	go clipboard.Write(clipboard.FmtText, []byte(latexTable))
	<-clipboard.Watch(context.Background(), clipboard.FmtText)
	glg.Success("Latex table copied to clipboard")
}
