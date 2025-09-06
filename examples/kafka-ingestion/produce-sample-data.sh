#!/bin/bash

# Script to produce sample Wikipedia data to Kafka
# Usage: ./produce-sample-data.sh [topic] [number_of_messages]

TOPIC=${1:-wikipedia}
NUM_MESSAGES=${2:-100}

echo "Producing $NUM_MESSAGES sample Wikipedia messages to topic '$TOPIC'"

# Create topic first if it doesn't exist
docker exec kafka /opt/kafka/bin/kafka-topics.sh \
    --create \
    --bootstrap-server localhost:9092 \
    --topic $TOPIC \
    --partitions 3 \
    --replication-factor 1 \
    --if-not-exists

# Generate and produce sample data
for i in $(seq 1 $NUM_MESSAGES); do
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")
    
    # Generate random sample data
    PAGES=("Main_Page" "Wikipedia" "Barack_Obama" "United_States" "World_War_II" "Python_(programming_language)" "Linux" "Google" "Facebook" "COVID-19")
    LANGUAGES=("en" "es" "fr" "de" "it" "pt" "ru" "ja" "zh" "ar")
    USERS=("User1" "User2" "User3" "AdminUser" "BotUser" "GuestUser")
    COUNTRIES=("United States" "United Kingdom" "Germany" "France" "Japan" "China" "Brazil" "Canada" "Australia" "India")
    CITIES=("New York" "London" "Berlin" "Paris" "Tokyo" "Beijing" "São Paulo" "Toronto" "Sydney" "Mumbai")
    
    PAGE=${PAGES[$((RANDOM % ${#PAGES[@]}))]}
    LANGUAGE=${LANGUAGES[$((RANDOM % ${#LANGUAGES[@]}))]}
    USER=${USERS[$((RANDOM % ${#USERS[@]}))]}
    COUNTRY=${COUNTRIES[$((RANDOM % ${#COUNTRIES[@]}))]}
    CITY=${CITIES[$((RANDOM % ${#CITIES[@]}))]}
    
    ADDED=$((RANDOM % 1000 + 1))
    DELETED=$((RANDOM % 500))
    DELTA=$((ADDED - DELETED))
    
    JSON_MESSAGE=$(cat << EOF
{
  "__time": "$TIMESTAMP",
  "page": "$PAGE",
  "language": "$LANGUAGE",
  "user": "$USER",
  "unpatrolled": "$(if [ $((RANDOM % 10)) -lt 2 ]; then echo true; else echo false; fi)",
  "newPage": "$(if [ $((RANDOM % 10)) -lt 1 ]; then echo true; else echo false; fi)",
  "robot": "$(if [ $((RANDOM % 10)) -lt 3 ]; then echo true; else echo false; fi)",
  "anonymous": "$(if [ $((RANDOM % 10)) -lt 4 ]; then echo true; else echo false; fi)",
  "namespace": "Main",
  "continent": "$(if [[ "$COUNTRY" == "United States" || "$COUNTRY" == "Canada" || "$COUNTRY" == "Brazil" ]]; then echo "North America"; elif [[ "$COUNTRY" == "United Kingdom" || "$COUNTRY" == "Germany" || "$COUNTRY" == "France" ]]; then echo "Europe"; else echo "Asia"; fi)",
  "country": "$COUNTRY",
  "region": "$(echo $COUNTRY | cut -d' ' -f1)",
  "city": "$CITY",
  "added": $ADDED,
  "deleted": $DELETED,
  "delta": $DELTA
}
EOF
    )
    
    echo "$JSON_MESSAGE" | tr -d '\n' | docker exec -i kafka /opt/kafka/bin/kafka-console-producer.sh \
        --topic $TOPIC \
        --bootstrap-server localhost:9092
    
    if [ $((i % 10)) -eq 0 ]; then
        echo "Produced $i messages..."
    fi
    
    # Small delay to avoid overwhelming
    sleep 0.1
done

echo "Successfully produced $NUM_MESSAGES messages to topic '$TOPIC'"
echo "You can now run your Terraform configuration to start ingesting this data into Druid"
