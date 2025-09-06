terraform {
  required_providers {
    druid = {
      source  = "local/provider/druid"
      version = "~> 1.0"
    }
  }
}

provider "druid" {
  endpoint = "http://localhost:8888"
  timeout  = 60
}

# Wikipedia sample data ingestion from Kafka
resource "druid_kafka_supervisor" "wikipedia_ingestion" {
  datasource = "wikipedia-kafka"

  # Timestamp configuration
  timestamp_spec {
    column = "__time"
    format = "iso"
  }

  # Dimension configuration - matches our sample data structure
  dimensions_spec {
    dimensions {
      name = "page"
      type = "string"
    }
    
    dimensions {
      name = "language"
      type = "string"
    }
    
    dimensions {
      name = "user"
      type = "string"
    }
    
    dimensions {
      name = "unpatrolled"
      type = "string"
    }
    
    dimensions {
      name = "newPage"
      type = "string"
    }
    
    dimensions {
      name = "robot"
      type = "string"
    }
    
    dimensions {
      name = "anonymous"
      type = "string"
    }
    
    dimensions {
      name = "namespace"
      type = "string"
    }
    
    dimensions {
      name = "continent"
      type = "string"
    }
    
    dimensions {
      name = "country"
      type = "string"
    }
    
    dimensions {
      name = "region"
      type = "string"
    }
    
    dimensions {
      name = "city"
      type = "string"
    }
  }

  # Metrics configuration - aggregating the numeric fields
  metrics_spec {
    name = "count"
    type = "count"
  }
  
  metrics_spec {
    name       = "added"
    type       = "longSum"
    field_name = "added"
  }
  
  metrics_spec {
    name       = "deleted"
    type       = "longSum"
    field_name = "deleted"
  }
  
  metrics_spec {
    name       = "delta"
    type       = "longSum"
    field_name = "delta"
  }

  # Granularity configuration
  granularity_spec {
    type                = "uniform"
    segment_granularity = "HOUR"
    query_granularity   = "MINUTE"
    rollup              = false
  }

  # Kafka topic to ingest from
  topic = "wikipedia"

  # JSON input format
  input_format {
    type = "json"
  }

  # Kafka consumer configuration
  consumer_properties = {
    "bootstrap.servers"        = "kafka:9092"
    "group.id"                = "druid-wikipedia-consumer"
    "auto.offset.reset"       = "earliest"
    "enable.auto.commit"      = "false"
    "max.poll.records"        = "500"
    "fetch.min.bytes"         = "1"
    "fetch.max.wait.ms"       = "500"
  }

  # Task configuration
  task_count           = 2
  replicas            = 1
  task_duration       = "PT30M"
  use_earliest_offset = true
  completion_timeout  = "PT20M"

  # Tuning configuration optimized for sample data
  tuning_config {
    max_rows_per_segment    = 1000000
    max_rows_in_memory      = 25000
    max_parse_exceptions    = 10
    log_parse_exceptions    = true
    intermediate_persist_period = "PT5M"
    
    # Reset offset automatically if there are issues
    reset_offset_automatically = true
    
    # Optimize for development
    worker_threads = 2
    chat_retries   = 3
    http_timeout   = "PT30S"
  }

  # Context for query optimization
  context = {
    "useCache"       = "false"
    "populateCache"  = "false"
    "priority"       = "50"
  }
}

# Output the supervisor details
output "supervisor_id" {
  value = druid_kafka_supervisor.wikipedia_ingestion.supervisor_id
}

output "supervisor_state" {
  value = druid_kafka_supervisor.wikipedia_ingestion.state
}

output "datasource" {
  value = druid_kafka_supervisor.wikipedia_ingestion.datasource
}