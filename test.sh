#!/usr/bin/env bash
echo 'Running go test in each directory.'
for tf in $(find . -name '*_test.go'); do
  echo ''
  d=$(dirname "$tf")
  echo "go test $d/*"
  go test "$d"/*.go
done

