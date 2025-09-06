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
  timeout  = 60
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

resource "druid_kafka_supervisor" "example" {
  datasource = "wikipedia"

  timestamp_spec {
    column = "__time"
    format = "iso"
  }

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

  metrics_spec {
    name = "count"
    type = "count"
  }
  
  metrics_spec {
    name = "added"
    type = "longSum"
    field_name = "added"
  }
  
  metrics_spec {
    name = "deleted"
    type = "longSum"
    field_name = "deleted"
  }
  
  metrics_spec {
    name = "delta"
    type = "longSum"
    field_name = "delta"
  }

  granularity_spec {
    type               = "uniform"
    segment_granularity = "HOUR"
    query_granularity   = "NONE"
    rollup             = false
  }

  topic = "wikipedia"

  input_format {
    type = "json"
  }

  consumer_properties = {
    "bootstrap.servers" = "localhost:9092"
  }

  task_count           = 1
  replicas            = 1
  task_duration       = "PT1H"
  use_earliest_offset = true

  tuning_config {
    max_rows_per_segment = 5000000
    max_rows_in_memory   = 25000
  }
}