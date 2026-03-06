#!/bin/bash
# Generate storage overhead data for all databases

export CGO_CFLAGS="-I/opt/homebrew/opt/gmp/include"
export CGO_LDFLAGS="-L/opt/homebrew/opt/gmp/lib"

mkdir -p result/Storage

echo "Generating storage overhead data for all databases..."

# Crime_USENIX_REV
echo "Crime_USENIX_REV:"
go run ./cmd/storage/main.go -db Crime_USENIX_REV 2>/dev/null | grep -A100 "CSV Format" | tail -n +3 > result/Storage/crime_storage.csv

# Enron_USENIX
echo "Enron_USENIX:"
go run ./cmd/storage/main.go -db Enron_USENIX 2>/dev/null | grep -A100 "CSV Format" | tail -n +3 > result/Storage/enron_storage.csv

# Wiki_USENIX
echo "Wiki_USENIX:"
go run ./cmd/storage/main.go -db Wiki_USENIX 2>/dev/null | grep -A100 "CSV Format" | tail -n +3 > result/Storage/wiki_storage.csv

echo "Done! Results saved to result/Storage/"
