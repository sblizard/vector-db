# Vector DB Development Roadmap

## Phase 1 – Set Up Your Ship

**Goal:** prepare the Go project skeleton and tools

### Initialize Go module

```bash
go mod init github.com/yourname/vectordb
go mod tidy
```

### Create the core folder structure

```bash
mkdir -p cmd/server internal/{api,engine,index,storage,math,config,util} pkg/vector data scripts test
```

### Add .gitignore

```bash
echo -e "data/\n*.bin\n*.db\n*.log\n" > .gitignore
```

### Create README.md
Describe your vision, goals, and architecture summary.

### Add dependencies

```bash
go get github.com/dgraph-io/badger/v4       # metadata KV store
go get gonum.org/v1/gonum/blas              # BLAS for dot products
go get github.com/gorilla/mux               # HTTP router
```

---

## Phase 2 – Lay the Keel (Scaffolding and Config)

**Goal:** basic runnable Go server with config file

1. Create `internal/config/config.go`
   - Load JSON config (`config.json`)
   - Provide defaults for dim, num_subspaces, etc.

2. Create `cmd/server/main.go`
   - Load config
   - Initialize logger
   - Start HTTP server with placeholder routes

3. Implement `internal/api/router.go` and `handler.go`
   - Endpoints: `/insert`, `/search`, `/stats`
   - For now: return stub JSON like `{"status":"ok"}`

4. Run:
   ```bash
   go run ./cmd/server
   ```

5. Commit:
   ```bash
   git add .
   git commit -m "Initial Go scaffolding with config + HTTP server"
   ```

---

## Phase 3 – Build the Engine Core

**Goal:** brute-force search that actually works

1. Implement `internal/math/dot.go`
   - Write basic dot product and cosine similarity functions.

2. Implement `internal/engine/query_engine.go`
   - Accept query, loop over vectors (simple array), compute cosine similarity.

3. Implement `internal/storage/fileio.go`
   - Store/load raw vectors as binary files (`vectors.bin`).

4. Implement `internal/storage/metadata.go`
   - Use BadgerDB to map IDs → vector offsets.

5. Wire up `POST /insert` and `POST /search` handlers to these functions.

6. Test by inserting a few vectors and searching them manually.

7. Commit.

---

## Phase 4 – Divide the Vector Sea (Subspaces / IVF)

**Goal:** add clustering and subspace routing

1. Implement `internal/index/kmeans.go`
   - Basic K-means to find `M` centroids.

2. Implement `internal/index/centroid.go`
   - Functions to find nearest centroid for a vector.

3. Implement `internal/index/subspace.go`
   - Struct that holds `[][]float32` vectors for each cluster.

4. Implement `internal/index/ivf_index.go`
   - Handles inserting vectors into the nearest subspace
   - Routes query to nearest centroids (top-L)
   - Runs brute-force inside those subspaces concurrently using goroutines

5. Modify `query_engine.go` to use `IVFIndex` instead of global brute-force.

6. Add parallelization via `sync.WaitGroup`.

7. Commit and benchmark.

---

## Phase 5 – Persistence and Performance

**Goal:** make it persistent and fast as hell

1. Implement `internal/storage/mmap.go` – load vector data with `syscall.Mmap`.
2. Pre-normalize vectors at insert time (`normalize.go`).
3. Add `internal/engine/heap.go` for Top-K merge.
4. Add timing metrics and logs for query latency.
5. Benchmark against flat brute-force baseline.
6. Commit with results.

---

## Phase 6 – Extras and Polish

**Goal:** usability and scaling

1. Implement `GET /stats` → show cluster sizes and memory usage.
2. Add CLI tools in `scripts/` for data generation & benchmarking.
3. Add tests in `test/` for insert/search correctness.
4. Add README instructions for running and example curl requests.
5. Optional: add Dockerfile for easy deployment.
6. Commit.

---

## Phase 7 – Chart Future Waters

**Next-level upgrades after the prototype works**

- Add quantization (float32 → uint8)
- Replace brute-force subspace search with HNSW graphs
- Add WAL & background flushing for durability
- Shard subspaces across nodes for distributed search
