terraform {
  required_providers {
    druid = {
      source  = "local/provider/druid"
      version = "~> 1.0"
    }
  }
}

provider "druid" {
  endpoint = "http://localhost:8080"
  username = var.druid_username
  password = var.druid_password
}

variable "druid_username" {
  description = "Druid username"
  type        = string
  default     = ""
}

variable "druid_password" {
  description = "Druid password"
  type        = string
  default     = ""
  sensitive   = true
}

resource "druid_kafka_supervisor" "advanced_example" {
  datasource = "events"

  timestamp_spec {
    column        = "timestamp"
    format        = "auto"
    missing_value = "2010-01-01T00:00:00Z"
  }

  dimensions_spec {
    dimensions {
      name                  = "product_id"
      type                  = "string"
    }
    
    dimensions {
      name                  = "category"
      type                  = "string"
      multi_value_handling  = "sorted_array"
    }
    
    dimensions {
      name = "user_id"
      type = "string"
    }
    
    dimensions {
      name = "session_id"
      type = "string"
    }

    dimension_exclusions = ["__time", "kafka.timestamp"]
  }

  metrics_spec {
    name = "count"
    type = "count"
  }
  
  metrics_spec {
    name       = "revenue"
    type       = "doubleSum"
    field_name = "price"
  }
  
  metrics_spec {
    name       = "quantity"
    type       = "longSum"
    field_name = "quantity"
  }

  granularity_spec {
    type                = "uniform"
    segment_granularity = "DAY"
    query_granularity   = "HOUR"
    rollup              = true
  }

  topic_pattern = "events-.*"

  input_format {
    type = "json"
  }

  consumer_properties = {
    "bootstrap.servers"                = "kafka:9092"
    "security.protocol"                = "SASL_SSL"
    "sasl.mechanism"                   = "PLAIN"
    "sasl.jaas.config"                = "org.apache.kafka.common.security.plain.PlainLoginModule required username=\"user\" password=\"pass\";"
    "ssl.truststore.location"          = "/path/to/truststore.jks"
    "ssl.truststore.password"          = "truststore-password"
    "group.id"                         = "druid-consumer-group"
    "auto.offset.reset"                = "earliest"
    "enable.auto.commit"               = "false"
    "max.poll.records"                 = "500"
    "fetch.min.bytes"                  = "1024"
    "fetch.max.wait.ms"                = "5000"
  }

  task_count            = 3
  replicas             = 2
  task_duration        = "PT2H"
  use_earliest_offset  = false
  completion_timeout   = "PT1H"

  idle_config {
    enabled              = true
    inactive_after_millis = 600000
  }

  tuning_config {
    max_rows_per_segment                    = 2000000
    max_rows_in_memory                      = 100000
    max_bytes_in_memory                     = 134217728
    skip_bytes_in_memory_overhead_check     = false
    max_pending_persists                    = 2
    intermediate_persist_period             = "PT5M"
    max_parse_exceptions                    = 100
    max_saved_parse_exceptions              = 50
    log_parse_exceptions                    = true
    reset_offset_automatically              = true
    worker_threads                          = 4
    chat_threads                            = 2
    chat_retries                            = 5
    http_timeout                            = "PT30S"
    shutdown_timeout                        = "PT2M"
    
    index_spec {
      bitmap = {
        type = "roaring"
      }
      dimension_compression = "lz4"
      metric_compression    = "lz4"
    }
  }

  context = {
    "useCache"              = "false"
    "populateCache"         = "false"
    "priority"              = "75"
    "lanes"                 = "1"
  }

  suspended = false
}
