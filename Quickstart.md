# Build and run
make build
./bin/litmuschaos-mcp-server

# Development with hot reload
make dev

# Cross-platform builds
make build-all

# Docker deployment
make docker-build && make docker-run
