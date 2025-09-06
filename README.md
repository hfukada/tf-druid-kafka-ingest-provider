# Terraform Provider for Druid Kafka Ingestion

This is a Terraform provider for managing Apache Druid Kafka ingestion supervisors as infrastructure-as-code. The provider allows you to define and manage Kafka ingestion specifications using Terraform, similar to the druid-operator Kubernetes CRDs but for Terraform environments.

## Features

- Implementation of the Druid Kafka ingestion supervisor specification
- Support for most data schema configurations (timestamp, dimensions, metrics, granularity, flattenspec)
- IO configuration including Kafka consumer properties and task management
- Authentication support (basic auth)
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
├── examples/                        # Usage examples
│   ├── basic/main.tf                # Basic configuration example
│   ├── advanced/main.tf             # Advanced configuration example
│   └── kafka-ingestion/             # Complete Kafka ingestion demo
│       ├── main.tf                  # Ingestion supervisor configuration
│       ├── docker-compose.yaml      # Docker environment setup
│       ├── environment              # Druid configuration
│       ├── setup-ingestion-demo.sh  # Automated setup script
│       ├── setup-kafka-topics.sh    # Kafka topic creation
│       ├── produce-sample-data.sh   # Sample data generation
│       ├── sample-queries.sql       # Sample Druid queries
│       └── README.md                # Demo documentation
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
- **Acceptance Tests**: Test complete resource lifecycle with mock Druid server
- **Test Utilities**: Mock Druid server for testing API interactions

## Development Environment

### Docker Setup

The repository includes a complete Docker Compose setup with Druid and Kafka for local development and testing. The Docker environment is located in `examples/kafka-ingestion/`:

```bash
# Navigate to the kafka-ingestion example
cd examples/kafka-ingestion

# Start the entire stack (Druid + Kafka + PostgreSQL + Zookeeper)
docker-compose up -d

# Wait for services to be ready, then create Kafka topics
./setup-kafka-topics.sh

# Produce sample data to test ingestion
./produce-sample-data.sh wikipedia 100
```

**Services:**
- **Druid Router**: http://localhost:8888
- **Kafka UI**: http://localhost:8080  
- **Kafka Broker**: localhost:9092
- **PostgreSQL**: localhost:5432
- **Zookeeper**: localhost:2181

**Druid Services:**
- Coordinator: localhost:8081
- Broker: localhost:8082  
- Historical: localhost:8083
- MiddleManager: localhost:8091

### Testing the Provider

#### Complete Kafka Ingestion Demo

For a complete end-to-end demonstration:

```bash
# Build and test everything with real Kafka ingestion
cd examples/kafka-ingestion
./setup-ingestion-demo.sh
```

This will create a supervisor that ingests sample Wikipedia data from Kafka into Druid.

#### Manual Testing

With the Docker environment running, you can also test manually:

```bash
# Build the provider
go build -o terraform-provider-druid

# Run Terraform with the basic example
cd examples/basic
terraform init
terraform plan
terraform apply
```

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

