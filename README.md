# Terraform Provider for Druid Kafka Ingestion

This is a Terraform provider for managing Apache Druid Kafka ingestion supervisors as infrastructure-as-code. The provider allows you to define and manage Kafka ingestion specifications using Terraform, similar to the druid-operator Kubernetes CRDs but for Terraform environments.

## Features

- Complete implementation of the Druid Kafka ingestion supervisor specification
- Support for all data schema configurations (timestamp, dimensions, metrics, granularity)
- Full IO configuration including Kafka consumer properties and task management
- Comprehensive tuning and performance configuration options
- Authentication support (basic auth)
- Proper error handling and validation
- Terraform state management with import/export capabilities

## Repository Structure

```
.
├── main.go                           # Provider entry point
├── go.mod                           # Go module dependencies
├── go.sum                           # Go module checksums
├── README.md                        # This file
├── CLAUDE.md                        # Development guidance for Claude Code
├── internal/provider/               # Core provider implementation
│   ├── provider.go                  # Main provider configuration
│   ├── config.go                    # Client configuration
│   ├── client.go                    # Druid API client
│   ├── resource_kafka_supervisor.go # Kafka supervisor resource
│   ├── testutils.go                 # Test utilities and mock server
│   ├── provider_test.go             # Provider unit tests
│   ├── client_test.go               # Client unit tests
│   ├── resource_kafka_supervisor_test.go          # Resource unit tests
│   └── resource_kafka_supervisor_acceptance_test.go # Acceptance tests
└── examples/                        # Usage examples
    ├── basic/main.tf                # Basic configuration example
    └── advanced/main.tf             # Advanced configuration example
```

## Development Setup

### Prerequisites

- Go 1.24.2 or later
- Terraform 1.0 or later (for testing)

### Building the Provider

```bash
# Build the provider binary
go build -o terraform-provider-druid

# Clean up dependencies
go mod tidy

# Format code
go fmt ./...
```

### Running Tests

#### Unit Tests

Run all unit tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test -v ./...
```

Run tests for a specific package:
```bash
go test -v ./internal/provider
```

Run a specific test:
```bash
go test -v ./internal/provider -run TestBuildSupervisorSpec
```

#### Test Coverage

Generate test coverage report:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### Acceptance Tests

Acceptance tests use a mock Druid server and test the full resource lifecycle:
```bash
go test -v ./internal/provider -run TestAcc
```

Run a specific acceptance test:
```bash
go test -v ./internal/provider -run TestAccKafkaSupervisor_basic
```

#### Test Structure

- **Unit Tests**: Test individual functions and components in isolation
  - `provider_test.go`: Tests provider configuration and validation
  - `client_test.go`: Tests HTTP client functionality with mock servers
  - `resource_kafka_supervisor_test.go`: Tests resource schema and spec building logic

- **Acceptance Tests**: Test complete resource lifecycle with mock Druid server
  - `resource_kafka_supervisor_acceptance_test.go`: Full CRUD operations testing

- **Test Utilities**: 
  - `testutils.go`: Mock Druid server for testing API interactions

### Testing Philosophy

The test suite includes:

1. **Schema Validation Tests**: Ensure all required fields, defaults, and validation rules work correctly
2. **Spec Building Tests**: Verify that Terraform configuration is correctly transformed into Druid API specifications
3. **API Client Tests**: Test HTTP client interactions with various response scenarios
4. **End-to-End Tests**: Full resource lifecycle testing with mock servers
5. **Error Handling Tests**: Comprehensive error scenario coverage

## Usage Examples

### Basic Configuration

```hcl
provider "druid" {
  endpoint = "http://localhost:8080"
}

resource "druid_kafka_supervisor" "example" {
  datasource = "wikipedia"

  timestamp_spec {
    column = "__time"
    format = "iso"
  }

  topic = "wikipedia"

  input_format {
    type = "json"
  }

  consumer_properties = {
    "bootstrap.servers" = "localhost:9092"
  }

  task_count = 1
  replicas   = 1
}
```

### Advanced Configuration

See `examples/advanced/main.tf` for a comprehensive configuration with:
- SSL/SASL authentication
- Multiple dimensions and metrics
- Performance tuning parameters
- Idle configuration
- Context parameters

## Provider Configuration

| Argument | Description | Type | Default | Required |
|----------|-------------|------|---------|----------|
| `endpoint` | Druid router endpoint URL | `string` | - | Yes |
| `username` | Username for authentication | `string` | `""` | No |
| `password` | Password for authentication | `string` | `""` | No |
| `timeout` | HTTP client timeout in seconds | `number` | `30` | No |

Environment variables:
- `DRUID_ENDPOINT`
- `DRUID_USERNAME`
- `DRUID_PASSWORD`

## Resource Schema

The `druid_kafka_supervisor` resource supports all Druid Kafka ingestion supervisor specification fields. Key configuration blocks include:

- `timestamp_spec`: Timestamp parsing configuration
- `dimensions_spec`: Dimension column definitions
- `metrics_spec`: Aggregation metrics
- `granularity_spec`: Segment and query granularity
- `input_format`: Data format specification (JSON, CSV, etc.)
- `consumer_properties`: Kafka consumer configuration
- `tuning_config`: Performance and resource tuning
- `idle_config`: Supervisor idle state management

For complete field documentation, see the resource schema in `resource_kafka_supervisor.go`.

## Contributing

1. Make changes to the codebase
2. Add or update tests as needed
3. Run the test suite: `go test ./...`
4. Ensure all tests pass
5. Format code: `go fmt ./...`
6. Build provider: `go build`

## Testing Guidelines

- Write unit tests for all new functionality
- Include both positive and negative test cases
- Test error conditions and edge cases
- Use table-driven tests for multiple scenarios
- Mock external dependencies (HTTP servers, etc.)
- Ensure acceptance tests cover full resource lifecycle

