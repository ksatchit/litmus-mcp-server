# LitmusChaos MCP Server (Go Implementation)

A comprehensive Model Context Protocol (MCP) server for LitmusChaos 3.x, built in Go, enabling AI assistants like Claude to interact with your chaos engineering platform. This server provides a complete interface for managing chaos experiments, infrastructures, environments, and resilience probes through natural language interactions.

## Features

### 🧪 **Chaos Experiment Management**
- List, create, and manage chaos experiments
- Execute experiments on-demand or via schedules
- Stop running experiments with granular control
- Support for custom fault configurations and weightings

### 🏗️ **Infrastructure Operations**
- Register and manage chaos infrastructures (formerly agents/delegates)
- Monitor infrastructure health and status
- Generate installation manifests
- Support for both namespace and cluster-scoped deployments

### 🌍 **Environment Organization**
- Create and manage environments (PROD/NON_PROD)
- Organize infrastructures by environment
- Environment-based filtering and operations

### 📊 **Experiment Execution Tracking**
- Detailed experiment run history and status
- Real-time execution monitoring
- Fault-level success/failure tracking
- Resiliency score calculations

### 🔍 **Resilience Probes**
- HTTP, Command, Kubernetes, and Prometheus probes
- Plug-and-play probe architecture
- Steady-state validation during chaos

### 📚 **ChaosHub Integration**
- Browse available chaos faults
- Multiple hub support (Git and Remote)
- Fault categorization and discovery

### 📈 **Statistics & Analytics**
- Comprehensive experiment and infrastructure statistics
- Resiliency score distributions
- Run status breakdowns

## Prerequisites

- Go 1.21 or higher
- Access to a LitmusChaos 3.x Chaos Center
- Valid project credentials

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/litmuschaos-mcp-server-go.git
cd litmuschaos-mcp-server-go

# Build the binary
make build

# Or install directly
make install
```

### Using Go Install

```bash
go install github.com/yourusername/litmuschaos-mcp-server-go@latest
```

### Using Docker

```bash
# Build the Docker image
make docker-build

# Run with Docker
docker run --rm -it \
  -e CHAOS_CENTER_ENDPOINT=http://your-chaos-center:8080 \
  -e LITMUS_PROJECT_ID=your-project-id \
  -e LITMUS_ACCESS_TOKEN=your-token \
  litmuschaos-mcp-server:latest
```

## Configuration

### Environment Variables

```bash
# Required Configuration
export CHAOS_CENTER_ENDPOINT=http://your-chaos-center:8080
export LITMUS_PROJECT_ID=your-project-id
export LITMUS_ACCESS_TOKEN=your-access-token

# Optional Defaults
export DEFAULT_INFRA_ID=your-default-infrastructure-id
export DEFAULT_ENVIRONMENT_ID=production
```

### Getting Your Credentials

1. **Chaos Center Endpoint**: URL of your LitmusChaos installation
2. **Project ID**: Found in your Chaos Center project settings
3. **Access Token**: Generate from Chaos Center → Settings → Access Tokens

## Usage

### With Claude Desktop

Add to your Claude Desktop MCP configuration:

```json
{
  "mcpServers": {
    "litmuschaos": {
      "command": "/path/to/litmuschaos-mcp-server",
      "env": {
        "CHAOS_CENTER_ENDPOINT": "http://localhost:8080",
        "LITMUS_PROJECT_ID": "your-project-id",
        "LITMUS_ACCESS_TOKEN": "your-token"
      }
    }
  }
}
```

### Standalone Usage

```bash
# Using environment variables
./bin/litmuschaos-mcp-server

# Or with make
make run
```

## Development

### Setup Development Environment

```bash
# Clone and setup
git clone https://github.com/yourusername/litmuschaos-mcp-server-go.git
cd litmuschaos-mcp-server-go

# Install dependencies
make deps

# Run with hot reload (requires air)
make dev
```

### Development Commands

```bash
# Build the project
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Run all checks
make check

# Clean build artifacts
make clean

# Build for all platforms
make build-all
```

### Project Structure

```
.
├── main.go              # Main server implementation
├── handlers.go          # Tool implementation handlers (part 1)
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
├── Makefile            # Build automation
├── Dockerfile          # Container build
├── .air.toml           # Development hot reload config
├── .golangci.yml       # Linter configuration
├── bin/                # Binary output directory
└── tmp/                # Development temporary files
```

## Available Tools

The server provides 17 comprehensive tools for chaos engineering operations:

### Experiment Management
- `list_chaos_experiments` - List all chaos experiments with filtering
- `get_chaos_experiment` - Get detailed experiment information
- `create_chaos_experiment` - Create new chaos experiments with custom faults
- `run_chaos_experiment` - Execute experiments immediately
- `stop_chaos_experiment` - Stop running experiments

### Execution Monitoring
- `list_experiment_runs` - List experiment execution history
- `get_experiment_run_details` - Get detailed run information with logs

### Infrastructure Management
- `list_chaos_infrastructures` - List all registered infrastructures
- `get_infrastructure_details` - Get detailed infrastructure information
- `register_chaos_infrastructure` - Register new Kubernetes infrastructures

### Environment Organization
- `list_environments` - List all environments
- `create_environment` - Create new environments for organization

### Resilience Validation
- `list_resilience_probes` - List all configured resilience probes
- `create_resilience_probe` - Create HTTP, CMD, K8s, or Prometheus probes

### Discovery & Analytics
- `list_chaos_hubs` - List available ChaosHubs
- `get_chaos_faults` - Browse available chaos faults
- `get_experiment_statistics` - Get comprehensive platform statistics

## Example Interactions

### Creating a Chaos Experiment

```
"Create a pod deletion experiment named 'payment-service-chaos' targeting the payment-service pods in production, with a 30-second duration and a weight of 8"
```

### Monitoring Experiments

```
"Show me the status of all running chaos experiments and their resiliency scores"
```

### Infrastructure Management

```
"List all active chaos infrastructures in the production environment"
```

### Resilience Validation

```
"Create an HTTP probe that checks if the payment API is responding with 200 status every 5 seconds"
```

## Performance & Optimization

The Go implementation provides several performance advantages:

- **Fast Startup**: Binary starts in milliseconds
- **Low Memory Usage**: Minimal runtime overhead
- **Concurrent Operations**: Efficient handling of multiple GraphQL requests
- **Static Binary**: Single executable with no dependencies
- **Cross-Platform**: Native binaries for Linux, macOS, and Windows

## Architecture

### Key Components

- **LitmusChaosServer**: Main server struct handling MCP protocol
- **GraphQL Client**: Direct communication with Chaos Center APIs
- **Tool Handlers**: Individual handlers for each chaos engineering operation
- **Error Handling**: Comprehensive error management and user feedback
- **JSON Processing**: High-performance JSON marshaling/unmarshaling

### Design Principles

- **Simplicity**: Clean, readable Go code following idiomatic patterns
- **Performance**: Optimized for speed and minimal resource usage
- **Reliability**: Comprehensive error handling and graceful failures
- **Maintainability**: Well-structured code with clear separation of concerns

## Building and Deployment

### Local Development

```bash
# Quick start
make dev

# Manual build and run
make build
./bin/litmuschaos-mcp-server
```

### Production Deployment

```bash
# Build optimized release binary
make build-release

# Build for multiple platforms
make build-all

# Build Docker image
make docker-build
```

### Cross-Compilation

```bash
# Build for Linux
make build-linux

# Build for macOS (Intel and Apple Silicon)
make build-darwin

# Build for Windows
make build-windows

# Build for all platforms
make build-all
```

## Error Handling

The server includes comprehensive error handling for:

- **Authentication Failures**: Invalid tokens or project IDs
- **Network Issues**: Connection timeouts and retries
- **GraphQL Errors**: API-specific error messages with context
- **Validation Errors**: Input parameter validation with detailed feedback
- **Infrastructure Issues**: Connectivity and status problems

## Security Considerations

- **Access Tokens**: Store securely and rotate regularly
- **Network Security**: Use HTTPS endpoints in production
- **Least Privilege**: Use project-specific tokens with minimal permissions
- **Container Security**: Runs as non-root user in minimal scratch image
- **Input Validation**: All parameters validated before processing

## Troubleshooting

### Common Issues

**Connection Failed**
```bash
# Verify Chaos Center is accessible
curl -f http://your-chaos-center:8080/health

# Check environment variables
echo $CHAOS_CENTER_ENDPOINT
echo $LITMUS_PROJECT_ID
```

**Build Issues**
```bash
# Update dependencies
make tidy

# Verify Go version
go version

# Clean and rebuild
make clean build
```

**Runtime Errors**
```bash
# Check logs for detailed error information
./bin/litmuschaos-mcp-server 2>&1 | tee debug.log

# Verify GraphQL connectivity
curl -X POST http://your-chaos-center:8080/query \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $LITMUS_ACCESS_TOKEN" \
  -d '{"query":"query { __typename }"}'
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following Go best practices
4. Add tests for new functionality
5. Run checks (`make check`)
6. Commit your changes (`git commit -am 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Add comprehensive error handling
- Include unit tests for new features
- Maintain backwards compatibility
- Update documentation for API changes
- Use meaningful commit messages

## License

MIT License - see LICENSE file for details.

## Support

- **Documentation**: [LitmusChaos Docs](https://docs.litmuschaos.io/)
- **Issues**: [GitHub Issues](https://github.com/ksatchit/litmuschaos-mcp-server-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ksatchit/litmuschaos-mcp-server-go/discussions)
- **Community**: [LitmusChaos Slack](https://slack.litmuschaos.io/)


