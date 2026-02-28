# ConjunctiveSSE

Reference implementation for the paper:

> **[PAPER TITLE]**
> [AUTHORS]
> *[VENUE, YEAR]*

This repository implements four Conjunctive Searchable Symmetric Encryption (SSE) schemes that support secure multi-keyword AND queries over encrypted databases.

| Scheme | Description |
|--------|-------------|
| **ODXT** | Original Dynamic Cross-Tags — baseline conjunctive SSE with TSet + XSet |
| **HDXT** | Hybrid Dynamic Cross-Tags — combines Mitra (single-keyword) and Auhme (conjunctive) sub-schemes |
| **FDXT** | Forward-secure Dynamic Cross-Tags — counter-based forward security |
| **SDSSE-CQ** | Structured Dynamic SSE with Conjunctive Queries — uses the [Aura](https://github.com/ZBCccc/Aura) framework |

## Prerequisites

### System dependencies

**libpbc** (pairing-based cryptography library) — required to build any scheme:

```bash
# macOS
brew install pbc

# Ubuntu / Debian
sudo apt-get install libpbc-dev

# From source: https://crypto.stanford.edu/pbc/download.html
```

**Go 1.23+**

```bash
# https://go.dev/dl/
go version  # verify
```

### Dataset (MongoDB)

All schemes read plaintext data from MongoDB. The easiest way to get the datasets is via Docker:

```bash
# Build the image (restores Crime, Enron, and Wiki datasets into MongoDB)
docker build -t conjunctive-sse .

# Run MongoDB with the datasets loaded
docker run -d -p 27017:27017 --name sse-mongo conjunctive-sse
```

Alternatively, point `mongo_uri` in the config files at any MongoDB instance that has the required collections loaded.

**Required MongoDB schema** — each database must have two collections:

- `keyword_ids` — documents of the form `{ "k": "<keyword>", "val_set": ["<id1>", "<id2>", ...] }`
- `id_keywords` — documents of the form `{ "id": "<doc_id>", "val_st": ["<kw1>", "<kw2>", ...] }`

Available datasets: `Crime_USENIX_REV`, `Enron_USENIX`, `Wiki_USENIX`

## Quick Start

```bash
git clone https://github.com/<YOUR_ORG>/ConjunctiveSSE
cd ConjunctiveSSE

# Install Go dependencies
go mod download

# Copy and edit the config for the scheme you want to run
cp cmd/SDSSE-CQ/configs/config.example.json cmd/SDSSE-CQ/configs/config.json

# Run ciphertext generation + search (phase "cs")
go run cmd/SDSSE-CQ/main.go
```

## Running Each Scheme

All schemes are driven by a `config.json` in their respective `cmd/<SCHEME>/configs/` directory.

```bash
go run cmd/ODXT/main.go
go run cmd/HDXT/main.go
go run cmd/FDXT/main.go
go run cmd/SDSSE-CQ/main.go
```

### Config file reference

```jsonc
{
    "db":        "Enron_USENIX",           // MongoDB database name
    "phase":     "cs",                     // "c" = setup only, "s" = search only, "cs" = both
    "group":     "Enron_USENIX_w2_keywords_2.txt",  // query file under configs/
    "del_rate":  0,                        // deletion rate % (ODXT only)
    "mongo_uri": "mongodb://localhost:27017"         // MongoDB connection URI
}
```

Pre-generated query files for each dataset are already included under `cmd/<SCHEME>/configs/`.

### Distributed mode (gRPC)

ODXT, HDXT, and FDXT support a client-server split over gRPC. Start the server first, then run the client with `-server`:

```bash
# Terminal 1 — server
go run cmd/ODXT/server/main.go

# Terminal 2 — client
go run cmd/ODXT/client/main.go -server localhost:50051 -db Enron_USENIX -mongo mongodb://localhost:27017
```

## Reproducing Paper Results

The benchmark results in the paper were produced with the following setup:

1. Start MongoDB with the Docker image above
2. For each scheme and dataset, run the **ciphertext generation phase** (`"phase": "c"`) first
3. Then run the **search phase** (`"phase": "s"`) with the corresponding query file

Results are written as CSV files to `result/<Phase>/<Scheme>/<Dataset>/`.

The encryption keys are fixed (`"0123456789123456"`) for reproducibility. To use random keys instead, pass `randomKey=true` in `DBSetup` / `Init`.

## Citation

If you use this code in your research, please cite:

```bibtex
@inproceedings{[CITE_KEY],
  title     = {[PAPER TITLE]},
  author    = {[AUTHORS]},
  booktitle = {[VENUE]},
  year      = {[YEAR]}
}
```

## License

MIT — see [LICENSE](LICENSE).
