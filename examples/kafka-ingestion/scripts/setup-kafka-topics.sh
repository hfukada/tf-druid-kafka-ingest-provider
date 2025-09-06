#!/bin/bash

# Script to create Kafka topics for testing Druid ingestion
# Run this after docker-compose is up and running

echo "Creating Kafka topics for Druid ingestion testing..."

# Wait for Kafka to be ready
echo "Waiting for Kafka to be ready..."
sleep 10

# Create topics
docker exec kafka kafka-topics --create --topic wikipedia --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
docker exec kafka kafka-topics --create --topic events --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
docker exec kafka kafka-topics --create --topic test-topic --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1

# List topics
echo "Created topics:"
docker exec kafka kafka-topics --list --bootstrap-server localhost:9092

echo "Kafka topics setup complete!"
echo ""
echo "You can now:"
echo "  - Access Kafka UI at http://localhost:8080"
echo "  - Access Druid router at http://localhost:8888"
echo "  - Use bootstrap.servers: localhost:9092 for Kafka consumers"