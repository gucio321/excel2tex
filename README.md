# Installation

Download an executable from Releases section and put somewhere on your system (e.g. inside of $pATH).

Optionally to get latest code do `go install github.com/gucio321/excel2tex/v2@latest`

# Usage

1. Open LibreOffice Calc or Excel
2. Select cells you'd like to put in you latex and Ctrl+C-copy them.
![Select cells](./images/select-cells.png)
3. Run excel2tex programm in your terminal or by double-click on windows (You can specify additional options - see [here](#command-line-arguments)).
4. Paste the output in your latex document and build it.
![Paste in latex](./images/paste-in-latex.png)

## Command Line Arguments

```console
$ excel2tex -h
2024-11-10 17:04:09	[INFO]:	Welcome to excel2tex
Usage of excel2tex:
  -bc
    	Bold first column.
  -long
    	Use longtable instead of table and tabularx (recomended -s c)
  -nb
    	Do not bold first row.
  -s string
    	Separator for table columns (latex table columns type) (default "X")
  -t string
    	Title of the table (default "XXXXX")
```

# Legal Notes

:warning: This project is not affiliated with LibreOffice nor Microsoft Excel. It is a personal project made for educational purposes only. Use it at your own risk.

Excel is a registered trademark of Microsoft Corporation in the United States and/or other countries.
