#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import json
import os

# Input and output file paths
input_file = "cmd/HDXT/configs/Enron_USENIX_filecnt_sorted.json"
output_file = "cmd/HDXT/configs/Enron_USENIX_w2_keywords_2.txt"

# Ensure input file exists
if not os.path.exists(input_file):
    print(f"Error: Input file not found {input_file}")
    exit(1)

# Read JSON file
print(f"Reading file {input_file}...")
with open(input_file, 'r', encoding='utf-8') as f:
    data = json.load(f)

# Find the first keyword with the highest occurrence count
max_count = 0
first_keyword = None
for keyword, count in data.items():
    if count > max_count:
        max_count = count
        first_keyword = keyword

if not first_keyword:
    print("Error: Keyword not found")
    exit(1)

print(f"First keyword (highest count): {first_keyword}, count: {max_count}")

# Extract keywords with different occurrence counts, taking only the first for each count
second_keywords = []
current_count = None
for keyword, count in data.items():
    if count != current_count:
        second_keywords.append(keyword)
        current_count = count

print(f"Found {len(second_keywords)} keywords with different occurrence counts")

# Generate query dataset
print(f"Generating query dataset to {output_file}...")
with open(output_file, 'w', encoding='utf-8') as f:
    for second_keyword in second_keywords:
        query = f"{second_keyword}#{first_keyword}\n"
        f.write(query)

print(f"Query dataset generation complete! Generated {len(second_keywords)} queries") 