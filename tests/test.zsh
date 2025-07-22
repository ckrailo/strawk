#!/bin/zsh

set -o errexit
set -o nounset
set -o pipefail

for testfile in tests/**/*.awk; do
  testname=$(basename $testfile | sed 's/.awk$//')
  echo "running test $testfile..."
  infile=$(echo $testfile | sed 's/.awk$/.in/')
  outfile=$(echo $testfile | sed 's/.awk$/.out/')
  # flags=$(cat "$testfile:A:h/flags")
  # ./bin/strawk -f "$testfile" $(echo $flags) "$infile" > ./bin/output
  ./bin/strawk -f "$testfile" "$infile" > ./bin/output
  if ! diff ./bin/output "$outfile" > /dev/null; then
    echo "ERROR: test $testname failed!"
  fi
done
