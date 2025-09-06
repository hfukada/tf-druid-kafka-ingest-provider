# Kafka Ingestion Example

This example demonstrates how to use the Druid Terraform provider to create a Kafka ingestion supervisor that ingests sample Wikipedia-like data from a Kafka stream.

## Prerequisites

1. Docker and Docker Compose installed
2. Go 1.24.2 or later
3. Terraform installed

## Quick Start

### 1. Start the Infrastructure

From the repository root:
```bash
# Start Druid, Kafka, and supporting services
docker-compose up -d

# Wait for services to be ready (about 60 seconds)
```

### 2. Run the Complete Demo

```bash
cd examples/kafka-ingestion
./setup-ingestion-demo.sh
```

This script will:
- Verify services are running
- Create Kafka topics
- Build the Terraform provider
- Generate sample data
- Apply the Terraform configuration
- Create the Druid ingestion supervisor

### 3. Verify Ingestion

**Check Supervisor Status:**
- Visit http://localhost:8888 (Druid Console)
- Go to "Ingestion" → "Supervisors"
- Verify "wikipedia-kafka" supervisor is running

**Monitor Kafka:**
- Visit http://localhost:8080 (Kafka UI)
- Check "wikipedia" topic for messages

**Query Data:**
- In Druid Console, go to "Query"
- Use the sample queries from `sample-queries.sql`

## Manual Steps (Alternative)

If you prefer to run steps manually:

### 1. Create Kafka Topics
```bash
../../scripts/setup-kafka-topics.sh
```

### 2. Build and Setup Provider
```bash
# Build provider
cd ../..
go build -o terraform-provider-druid

# Setup Terraform
cd examples/kafka-ingestion
mkdir -p .terraform/providers/local/provider/druid/1.0.0/linux_amd64/
cp ../../terraform-provider-druid .terraform/providers/local/provider/druid/1.0.0/linux_amd64/
terraform init
```

### 3. Generate Sample Data
```bash
../../scripts/produce-sample-data.sh wikipedia 100
```

### 4. Apply Terraform
```bash
terraform plan
terraform apply
```

## Configuration Details

The `main.tf` configuration creates a supervisor that:

- **Datasource**: `wikipedia-kafka`
- **Topic**: `wikipedia` 
- **Format**: JSON with automatic field discovery
- **Dimensions**: All string fields from sample data
- **Metrics**: Count, added, deleted, delta aggregations
- **Granularity**: Hour segments, minute query granularity
- **Tasks**: 2 parallel tasks for better throughput

## Monitoring and Troubleshooting

### Check Supervisor Status
```bash
# View Terraform state
terraform show

# Check supervisor in Druid console
curl -s "http://localhost:8888/druid/indexer/v1/supervisor/wikipedia-kafka-supervisor/status" | jq
```

### Generate More Data
```bash
# Continuous data generation
for i in {1..10}; do
  ../../scripts/produce-sample-data.sh wikipedia 50
  sleep 30
done
```

### View Logs
```bash
# Druid Middle Manager logs
docker logs middlemanager

# Kafka logs
docker logs kafka
```

## Sample Queries

See `sample-queries.sql` for example SQL queries to explore the ingested data:

1. **Basic exploration**: Recent records
2. **Language analysis**: Activity by language
3. **Time series**: Activity over time windows
4. **User patterns**: Anonymous vs registered users
5. **Real-time monitoring**: Per-minute metrics

## Cleanup

```bash
# Remove the supervisor
terraform destroy -auto-approve

# Stop all services
docker-compose down -v
```

## Customization

### Modify Sample Data

Edit `../../scripts/produce-sample-data.sh` to change:
- Field values and distributions  
- Message frequency
- Data volume

### Tune Ingestion Performance

In `main.tf`, adjust:
- `task_count`: Number of parallel tasks
- `max_rows_in_memory`: Memory buffer size
- `task_duration`: How long tasks run
- `intermediate_persist_period`: Persistence frequency

### Add More Metrics

Add aggregation metrics in the `metrics_spec` blocks:
```hcl
metrics_spec {
  name       = "unique_users"
  type       = "hyperUnique"
  field_name = "user"
}
```

## Expected Results

After successful ingestion, you should see:
- Supervisor in "RUNNING" state
- Growing row count in datasource
- Real-time data in queries
- Segments being created in Druid