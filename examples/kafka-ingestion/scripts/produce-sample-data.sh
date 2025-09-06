#!/bin/bash

# Script to produce sample JSON data to Kafka topics for testing Druid ingestion
# Run this after Kafka topics are created

TOPIC=${1:-wikipedia}
COUNT=${2:-100}

echo "Producing $COUNT sample messages to topic: $TOPIC"

for i in $(seq 1 $COUNT); do
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")
    
    # Generate sample Wikipedia-like data
    SAMPLE_DATA=$(cat <<EOF
{
  "__time": "$TIMESTAMP",
  "page": "Page_$((RANDOM % 1000))",
  "language": "$(shuf -n1 -e en es fr de it)",
  "user": "User_$((RANDOM % 100))",
  "unpatrolled": "$(shuf -n1 -e true false)",
  "newPage": "$(shuf -n1 -e true false)",
  "robot": "$(shuf -n1 -e true false)",
  "anonymous": "$(shuf -n1 -e true false)",
  "namespace": "$(shuf -n1 -e Main User Talk Help)",
  "continent": "$(shuf -n1 -e 'North America' 'Europe' 'Asia' 'South America')",
  "country": "$(shuf -n1 -e 'United States' 'United Kingdom' 'Germany' 'France' 'Spain')",
  "region": "Region_$((RANDOM % 50))",
  "city": "City_$((RANDOM % 200))",
  "added": $((RANDOM % 1000)),
  "deleted": $((RANDOM % 100)),
  "delta": $((RANDOM % 500 - 250))
}
EOF
)
    
    # Send to Kafka
    echo "$SAMPLE_DATA" | docker exec -i kafka kafka-console-producer --bootstrap-server localhost:9092 --topic $TOPIC
    
    if [ $((i % 10)) -eq 0 ]; then
        echo "Produced $i messages..."
    fi
    
    # Small delay to avoid overwhelming
    sleep 0.1
done

echo "Finished producing $COUNT messages to topic: $TOPIC"
echo "You can verify the messages in Kafka UI at http://localhost:8080"