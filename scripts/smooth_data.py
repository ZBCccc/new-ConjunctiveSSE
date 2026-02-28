#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys
import csv
import pandas as pd
import argparse
from pathlib import Path

def smooth_data(input_file):
    """
    Smooth data in CSV file:
    - If clientTime > 80, set it to the first value less than 80 nearby
    - If serverTime > 70, set it to the first value less than 70 nearby
    - If clientTime or serverTime is modified, update totalTime to clientTime+serverTime

    Args:
        input_file: Input CSV file path
    """
    # Get output file name
    input_path = Path(input_file)
    output_dir = input_path.parent
    output_file = output_dir / f"processed_{input_path.name}"

    print(f"Processing file: {input_file}")
    print(f"Output file: {output_file}")

    # Read CSV file
    df = pd.read_csv(input_file, sep=',', skipinitialspace=True)

    # Create a copy of original data for comparison
    original_df = df.copy()

    # Smooth clientTime column
    for i in range(len(df)):
        if df.at[i, 'clientTime'] > 70:
            # Find the first value less than 70 nearby
            found = False
            # Search forward first
            for j in range(i-1, -1, -1):
                if df.at[j, 'clientTime'] <= 70:
                    df.at[i, 'clientTime'] = df.at[j, 'clientTime']
                    found = True
                    break
            # If not found forward, search backward
            if not found:
                for j in range(i+1, len(df)):
                    if df.at[j, 'clientTime'] <= 70:
                        df.at[i, 'clientTime'] = df.at[j, 'clientTime']
                        break

    # Smooth serverTime column
    for i in range(len(df)):
        if df.at[i, 'serverTime'] > 80:
            # Find the first value less than 80 nearby
            found = False
            # Search forward first
            for j in range(i-1, -1, -1):
                if df.at[j, 'serverTime'] <= 80:
                    df.at[i, 'serverTime'] = df.at[j, 'serverTime']
                    found = True
                    break
            # If not found forward, search backward
            if not found:
                for j in range(i+1, len(df)):
                    if df.at[j, 'serverTime'] <= 80:
                        df.at[i, 'serverTime'] = df.at[j, 'serverTime']
                        break

    # Update totalTime
    for i in range(len(df)):
        if (df.at[i, 'clientTime'] != original_df.at[i, 'clientTime'] or
            df.at[i, 'serverTime'] != original_df.at[i, 'serverTime']):
            df.at[i, 'totalTime'] = df.at[i, 'clientTime'] + df.at[i, 'serverTime']

    # Save processed data
    df.to_csv(output_file, index=False)

    # Count modified rows
    client_modified = sum(df['clientTime'] != original_df['clientTime'])
    server_modified = sum(df['serverTime'] != original_df['serverTime'])
    total_modified = sum((df['clientTime'] != original_df['clientTime']) |
                         (df['serverTime'] != original_df['serverTime']))

    print(f"Modification statistics:")
    print(f"  clientTime modified rows: {client_modified}")
    print(f"  serverTime modified rows: {server_modified}")
    print(f"  Total modified rows: {total_modified}")
    print(f"Processing complete. Results saved to: {output_file}")

def main():
    parser = argparse.ArgumentParser(description='Smooth CSV data file')
    parser.add_argument('input_file', help='Input CSV file path')
    args = parser.parse_args()

    if not os.path.exists(args.input_file):
        print(f"Error: File '{args.input_file}' does not exist")
        sys.exit(1)

    smooth_data(args.input_file)

if __name__ == "__main__":
    main() 