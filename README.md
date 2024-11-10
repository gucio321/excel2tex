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
$ excel2tex --help
Usage of excel2tex.sh:

-t, --title:    Set Table's title
                Default: XXXXX
-s, --sep:      Set custom "Table Separator" (aka column type in latex)
                Default is X
                NOTE: You might also want this in your preamble:
                \newcolumntype{L}{>{\raggedright\arraybackslash}X}
                \newcolumntype{Y}{>{\centering\arraybackslash}X}
-y:             Alias to -s Y
-n:             Set number of table columns.
                NOTE: by default uses 1st program argument if no other options specified
                Default: --help
-h, --help:     Show this message and exit.
```

# Legal Notes

:warning: This project is not affiliated with LibreOffice nor Microsoft Excel. It is a personal project made for educational purposes only. Use it at your own risk.

Excel is a registered trademark of Microsoft Corporation in the United States and/or other countries.
