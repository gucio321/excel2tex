#!/bin/bash

SEP="X"
TITLE="XXXXX"
N=$1

help() {
        cat <<EOF
Usage of excel2tex.sh:

-t, --title:    Set Table's title
                Default: $TITLE
-s, --sep:      Set custom "Table Separator" (aka column type in latex)
                Default is $SEP
                NOTE: You might also want this in your preamble:
                \\newcolumntype{L}{>{\\raggedright\\arraybackslash}X}
                \\newcolumntype{Y}{>{\\centering\\arraybackslash}X}
-y:             Alias to -s Y
-n:             Set number of table columns.
                NOTE: by default uses 1st program argument if no other options specified
                Default: $N
-h, --help:     Show this message and exit.
EOF
}

while [[ $# -gt 0 ]]; do
  case $1 in
    -y|--center)
      SEP="Y"
      shift # past argument
      ;;
    -s|--sep)
      SEP="$2"
      shift # past argument
      shift # past value
      ;;
    -t|--title)
      TITLE="$2"
      shift # past argument
      shift # past value
      ;;
    -h|--help)
      help
      exit 0
      ;;
    -n)
      N="$2"
      shift # past argument
      shift # past value
      ;;
    -*|--*)
      echo "Unknown option $1"
      help
      exit 1
      ;;
    *)
      POSITIONAL_ARGS+=("$1") # save positional arg
      shift # past argument
      ;;
  esac
done

set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters

echo -e "\\\begin{table}[ht]\n\\\caption{$TITLE}\n\\\centering\n \\\begin{tabularx}{\\\textwidth}{|$(printf "$SEP|%.0s" $(seq 1 $N))}\n \\\\hline \n$(xclip -o -selection clipboard |sed -e 's/\t\+/ \& /g'  -e '1s/\( *\)\([^&]\+[^&^ ]\)\( *\)/\1\\\\textbf{\2}\3/g' -e 's/$/ \\\\\\\\ \\\\hline/g')\n\\\end{tabularx}\n\\\end{table}" |xclip
