#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import json
import os

# Input and output file paths
input_file = "cmd/HDXT/configs/Enron_USENIX_filecnt.json"
output_file = "cmd/HDXT/configs/Enron_USENIX_filecnt_sorted.json"

# Ensure input file exists
if not os.path.exists(input_file):
    print(f"Error: Input file not found {input_file}")
    exit(1)

# Read JSON file
print(f"Reading file {input_file}...")
with open(input_file, 'r', encoding='utf-8') as f:
    data = json.load(f)

# Sort by value (keyword occurrence count) in ascending order
print("Sorting by keyword occurrence count in ascending order...")
sorted_data = {k: v for k, v in sorted(data.items(), key=lambda item: item[1])}

# Write sorted data to new file
print(f"Writing sorted data to {output_file}...")
with open(output_file, 'w', encoding='utf-8') as f:
    json.dump(sorted_data, f, indent=2, ensure_ascii=False)

print(f"Sorting complete! Results saved to {output_file}")

# Output some statistics
print(f"Total keywords: {len(data)}")
min_key = min(sorted_data.items(), key=lambda x: x[1])
max_key = max(sorted_data.items(), key=lambda x: x[1])
print(f"Keyword with minimum count: {min_key[0]} ({min_key[1]} times)")
print(f"Keyword with maximum count: {max_key[0]} ({max_key[1]} times)") 