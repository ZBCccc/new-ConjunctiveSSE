# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## Project Overview

ConjunctiveSSE is a research implementation of four Conjunctive Searchable Symmetric Encryption (SSE) schemes that support secure multi-keyword AND queries over encrypted databases:

- **ODXT**: Original Dynamic Cross-Tags (baseline with TSet + XSet)
- **HDXT**: Hybrid Dynamic Cross-Tags (combines Mitra single-keyword + Auhme conjunctive sub-schemes)
- **FDXT**: Forward-secure Dynamic Cross-Tags (counter-based forward security)
- **SDSSE-CQ**: Structured Dynamic SSE with Conjunctive Queries (uses Aura framework)

## Architecture

### Directory Structure

```
cmd/<SCHEME>/
  ├── client/main.go    # gRPC client entry point
  ├── server/main.go    # gRPC server entry point
  └── configs/          # Config files and query keyword files

pkg/<SCHEME>/
  ├── <SCHEME>.go       # Core encryption/search logic
  ├── client/           # gRPC client implementation
  ├── server/           # gRPC server implementation
  └── proto/            # Protocol buffer definitions

pkg/Database/           # MongoDB operations
pkg/utils/              # Shared cryptographic utilities (PBC, AUHME, etc.)
```

### Execution Modes

Each scheme supports two execution modes:

1. **Standalone mode**: Client and server logic run in the same process (legacy, not actively maintained)
2. **Distributed mode (gRPC)**: Client-server split over gRPC (current architecture)

### Data Flow

1. **Setup/Update Phase**: Client reads plaintext from MongoDB → encrypts data → sends ciphertext to server via gRPC
2. **Search Phase**: Client generates search tokens → sends to server → server searches encrypted index → returns results

## Building and Running

### Prerequisites

**System dependencies:**
```bash
# macOS
brew install pbc gmp

# Ubuntu/Debian
sudo apt-get install libpbc-dev libgmp-dev
```

**Go version:** 1.23+

**Environment setup (macOS with Homebrew):**
```bash
export CGO_CFLAGS="-I/opt/homebrew/opt/gmp/include -I/usr/local/include"
export CGO_LDFLAGS="-L/opt/homebrew/opt/gmp/lib -L/usr/local/lib"
```

Add these to your `~/.zshrc` or `~/.bashrc` to make them permanent.

**MongoDB:** Required for plaintext data. Use Docker image or local instance:
```bash
docker build -t conjunctive-sse .
docker run -d -p 27017:27017 --name sse-mongo conjunctive-sse
```

### Running a Scheme (Distributed Mode)

**Start server:**
```bash
go run cmd/<SCHEME>/server/main.go
```

**Run client:**
```bash
# ODXT/HDXT/FDXT
go run cmd/<SCHEME>/client/main.go -server localhost:50051 -db <DB_NAME> -mongo mongodb://localhost:27017

# SDSSE-CQ
go run cmd/SDSSE-CQ/client/main.go -server localhost:50051 -db <DB_NAME> -mongo mongodb://localhost:27017 -phase cs -group keywords_2.txt
```

**Available databases:** `Crime_USENIX_REV`, `Enron_USENIX`, `Wiki_USENIX`

### Client Flags

- `-server`: gRPC server address (default: `localhost:50051`)
- `-db`: MongoDB database name
- `-mongo`: MongoDB URI (default: `mongodb://localhost:27017`)
- `-phase`: (SDSSE-CQ only) `c` = setup, `s` = search, `cs` = both
- `-group`: (SDSSE-CQ only) Query file name in `configs/`

### Testing

```bash
# Run all tests (requires pbc/gmp installed)
go test ./...

# Run specific package tests
go test ./pkg/utils/...
```

## Protocol Buffers

When modifying `.proto` files, regenerate Go code:

```bash
protoc --go_out=. --go-grpc_out=. pkg/<SCHEME>/proto/<scheme>.proto
```

## MongoDB Schema

Each database requires two collections:

- `keyword_ids`: `{ "k": "<keyword>", "val_set": ["<id1>", "<id2>", ...] }`
- `id_keywords`: `{ "id": "<doc_id>", "val_st": ["<kw1>", "<kw2>", ...] }`

## Results Output

Benchmark results are written to CSV files:
- Setup/Update: `result/Update/<SCHEME>/<DB_NAME>/<timestamp>.csv`
- Search: `result/Search/<SCHEME>/<DB_NAME>/<timestamp>.csv`

## Key Implementation Details

### HDXT Phases

HDXT has three distinct phases:
1. **Setup**: Processes first half of `id_keywords` collection using `hdxt.Setup()`
2. **Update**: Processes second half using `hdxt.Update()` (via gRPC)
3. **Search**: Executes conjunctive queries

### Encryption Keys

Keys are hardcoded to `"0123456789123456"` for reproducibility. To use random keys, pass `randomKey=true` to `Init()`.

### PBC (Pairing-Based Cryptography)

ODXT and FDXT use the `github.com/Nik-U/pbc` library for pairing operations. The `pkg/utils/pbc/` wrapper provides helper functions for PRF-to-Zr conversions and group operations.

### SDSSE-CQ Integration

SDSSE-CQ uses the external Aura framework (`github.com/ZBCccc/Aura`). The client wraps Aura's `Client` type and communicates with the server via gRPC.

## Common Issues

### Build Failures

If you see `fatal error: 'gmp.h' file not found` or `'pbc/pbc.h' file not found`:
- Ensure `libpbc-dev` and `libgmp-dev` are installed
- On macOS, verify Homebrew paths: `brew --prefix pbc` and `brew --prefix gmp`
- Set CGO flags if needed:
  ```bash
  export CGO_CFLAGS="-I/usr/local/include"
  export CGO_LDFLAGS="-L/usr/local/lib"
  ```

### MongoDB Connection

If MongoDB connection fails, verify:
- MongoDB is running: `docker ps` or `systemctl status mongod`
- URI is correct: `mongodb://localhost:27017`
- Database and collections exist

## Development Notes

- The codebase is research-oriented; focus on correctness over production-readiness
- Each scheme is largely independent; changes to one scheme rarely affect others
- The `pkg/utils/` package contains shared cryptographic primitives used across schemes
- gRPC communication uses streaming for Setup phase (HDXT) to handle large ciphertext batches
