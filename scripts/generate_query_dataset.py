#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import json
import os

# Input and output file paths
input_file = "cmd/HDXT/configs/Enron_USENIX_filecnt_sorted.json"
output_file = "cmd/HDXT/configs/Enron_USENIX_keywords_w1_generated.txt"

# Ensure input file exists
if not os.path.exists(input_file):
    print(f"Error: Input file not found {input_file}")
    exit(1)

# Read JSON file
print(f"Reading file {input_file}...")
with open(input_file, 'r', encoding='utf-8') as f:
    data = json.load(f)

# Find the first keyword with occurrence count of 10
first_keyword = None
for keyword, count in data.items():
    if count == 10:
        first_keyword = keyword
        break

if not first_keyword:
    print("Error: Keyword with count 10 not found")
    exit(1)

print(f"First keyword (count = 10): {first_keyword}")

# Extract keywords with occurrence count > 10, taking only the first for each count
second_keywords = []
current_count = None
for keyword, count in data.items():
    if count > 10:
        if count != current_count:
            second_keywords.append(keyword)
            current_count = count

print(f"Found {len(second_keywords)} keywords with different occurrence counts (>10)")

# Generate query dataset
print(f"Generating query dataset to {output_file}...")
with open(output_file, 'w', encoding='utf-8') as f:
    for second_keyword in second_keywords:
        query = f"{first_keyword}#{second_keyword}\n"
        f.write(query)

print(f"Query dataset generation complete! Generated {len(second_keywords)} queries") 