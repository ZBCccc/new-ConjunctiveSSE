#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import json
import pymongo
import argparse
import sys
from collections import Counter
from tqdm import tqdm

def main():
    # Parse command line arguments
    parser = argparse.ArgumentParser(description='Count the number of file IDs for each keyword in MongoDB')
    parser.add_argument('--host', default='localhost', help='MongoDB host address')
    parser.add_argument('--port', type=int, default=27017, help='MongoDB port')
    parser.add_argument('--db', default='Enron_USENIX', help='MongoDB database name')
    parser.add_argument('--collection', default='id_keywords', help='MongoDB collection name')
    parser.add_argument('--output', default='cmd/HDXT/configs/Enron_USENIX_filecnt.json', help='Output JSON file path')
    parser.add_argument('--batch-size', type=int, default=1000, help='Batch size')
    args = parser.parse_args()

    try:
        # Connect to MongoDB
        client = pymongo.MongoClient(f"mongodb://{args.host}:{args.port}/",
                                    serverSelectionTimeoutMS=5000)

        # Check if connection is successful
        client.server_info()
        print(f"Successfully connected to MongoDB: {args.host}:{args.port}")

        # Select database and collection
        db = client[args.db]
        collection = db[args.collection]

        # Get total document count
        total_docs = collection.count_documents({})
        print(f"Collection '{args.collection}' has {total_docs} documents")

        if total_docs == 0:
            print("Warning: Collection is empty, no data to process")
            return

        # Initialize keyword counter
        keyword_counter = Counter()

        # Process large data with batching and progress bar
        print("Counting keyword file count...")
        cursor = collection.find({}, batch_size=args.batch_size)

        for doc in tqdm(cursor, total=total_docs, desc="Processing progress"):
            # Ensure document contains required fields
            if "id" in doc and "keywords" in doc:
                # Get keyword list
                keywords = doc["keywords"]

                # Count each keyword (deduplicate, each ID counted only once)
                unique_keywords = set(keywords)
                for keyword in unique_keywords:
                    keyword_counter[keyword] += 1

        # Convert counter to dictionary
        keyword_count_dict = dict(keyword_counter)

        # Save result to JSON file
        with open(args.output, "w", encoding="utf-8") as f:
            json.dump(keyword_count_dict, f, ensure_ascii=False, indent=2)

        print(f"Keyword file count saved to {args.output}")
        print(f"Total keywords counted: {len(keyword_count_dict)}")

    except pymongo.errors.ServerSelectionTimeoutError:
        print(f"Error: Cannot connect to MongoDB server {args.host}:{args.port}")
        sys.exit(1)
    except pymongo.errors.OperationFailure as e:
        print(f"MongoDB operation failed: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"Error occurred: {e}")
        sys.exit(1)
    finally:
        if 'client' in locals():
            client.close()
            print("MongoDB connection closed")

if __name__ == "__main__":
    main()
