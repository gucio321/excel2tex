#!/bin/bash
help() {
        cat <<EOF
Usage of excel2tex.sh:
EOF
}

SEP="X"
TITLE="XXXXX"
N=$1

while [[ $# -gt 0 ]]; do
  case $1 in
    -y|--center)
      SEP="Y"
      shift # past argument
      ;;
    -s|--separator)
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
