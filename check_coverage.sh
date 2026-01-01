#!/bin/bash
set -e

# Extract total coverage (excluding cmd)
go test -coverprofile=coverage.out $(go list ./... | grep -v cmd)
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

echo "Total Coverage: $TOTAL_COVERAGE%"

if python3 -c "import sys; sys.exit(0 if float('$TOTAL_COVERAGE') >= 65.0 else 1)"; then
  echo "Coverage achieved!"
  exit 0
else
  echo "Coverage is below 65%!"
  exit 1
fi
