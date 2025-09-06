#!/bin/bash

# Complete setup script for Kafka ingestion demonstration
set -e

echo "=== Druid Kafka Ingestion Demo Setup ==="
echo ""

# Check if docker-compose is running
if ! docker ps | grep -q "druid\|kafka"; then
    echo "ERROR: Docker services not running. Please start with:"
    echo "   docker compose up -d"
    exit 1
fi

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 15

# Build the Terraform provider
echo "Building Terraform provider..."
cd ../..
go build -o terraform-provider-druid
cd examples/kafka-ingestion

# Create provider plugin directory
echo "Setting up Terraform provider..."
mkdir -p .terraform/providers/local/provider/druid/1.0.0/linux_amd64/
cp ../../terraform-provider-druid .terraform/providers/local/provider/druid/1.0.0/linux_amd64/

# Initialize Terraform
echo "Initializing Terraform..."
terraform init

# Generate sample data
echo "Creating Kafka topics/sample data..."
./produce-sample-data.sh

echo "Waiting a moment for data to be available..."
sleep 5

# Apply Terraform configuration
echo "Creating Druid ingestion supervisor..."
terraform plan
terraform apply -auto-approve

echo ""
echo "Demo setup complete!"
echo ""
echo "Access points:"
echo "   - Druid Console: http://localhost:8888"
echo "   - Kafka UI: http://localhost:8080"
echo ""
echo "You can now:"
echo "   1. Check supervisor status in Druid console"
echo "   2. Query the 'wikipedia-kafka' datasource"
echo "   3. Generate more data: ../../scripts/produce-sample-data.sh wikipedia 100"
echo "   4. View ingestion progress in real-time"
echo ""
echo "Sample Druid query:"
echo "   SELECT * FROM \"wikipedia-kafka\" WHERE __time >= CURRENT_TIMESTAMP - INTERVAL '1' HOUR LIMIT 10"
echo ""
echo "To cleanup:"
echo "   terraform destroy -auto-approve"
