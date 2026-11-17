.PHONY: all frontend backend clean run dev

# Default target
all: frontend backend

# Build the Vue.js frontend (writes to ../dist for go:embed)
frontend:
	@echo "Building Frontend..."
	cd wsjtx-web-ui && npm install && npm run build

# Build the Go backend
backend:
	@echo "Building Backend..."
	go build -o wsjtx-web main.go

# Clean build artifacts (Vite outDir is ../dist, not wsjtx-web-ui/dist)
clean:
	@echo "Cleaning..."
	rm -f wsjtx-web
	rm -rf dist/assets
	# Keep tracked placeholder dist/index.html for go:embed

# Two-terminal local dev: run `make dev` in one terminal after starting Go,
# or use the recipe documented in README / AGENTS.md.
dev:
	@echo "Starting Vite on :5173 (proxies /ws → :8080)..."
	@echo "Ensure Go is running: go run main.go   (or -udp :2238 if GridTracker owns :2237)"
	cd wsjtx-web-ui && npm run dev

# Run the production binary (build first with `make all`)
run: all
	@echo "Starting application..."
	./wsjtx-web
