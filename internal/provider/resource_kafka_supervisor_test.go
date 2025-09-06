package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSupervisorSpec(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "basic spec",
			input: map[string]interface{}{
				"datasource": "test-datasource",
				"timestamp_spec": []interface{}{
					map[string]interface{}{
						"column": "__time",
						"format": "iso",
					},
				},
				"topic": "test-topic",
				"input_format": []interface{}{
					map[string]interface{}{
						"type": "json",
					},
				},
				"consumer_properties": map[string]interface{}{
					"bootstrap.servers": "localhost:9092",
				},
				"task_count":           1,
				"replicas":            1,
				"task_duration":       "PT1H",
				"use_earliest_offset": false,
				"completion_timeout":  "PT30M",
				"context":             map[string]interface{}{},
				"suspended":           false,
			},
			expected: map[string]interface{}{
				"type": "kafka",
				"spec": map[string]interface{}{
					"dataSchema": map[string]interface{}{
						"dataSource": "test-datasource",
						"timestampSpec": map[string]interface{}{
							"column": "__time",
							"format": "iso",
						},
					},
					"ioConfig": map[string]interface{}{
						"topic": "test-topic",
						"inputFormat": map[string]interface{}{
							"type": "json",
						},
						"consumerProperties": map[string]interface{}{
							"bootstrap.servers": "localhost:9092",
						},
						"taskCount":          1,
						"replicas":           1,
						"taskDuration":       "PT1H",
						"useEarliestOffset":  false,
						"completionTimeout":  "PT30M",
					},
				},
			},
		},
		{
			name: "advanced spec with dimensions and metrics",
			input: map[string]interface{}{
				"datasource": "advanced-datasource",
				"timestamp_spec": []interface{}{
					map[string]interface{}{
						"column":        "timestamp",
						"format":        "auto",
						"missing_value": "2010-01-01T00:00:00Z",
					},
				},
				"dimensions_spec": []interface{}{
					map[string]interface{}{
						"dimensions": []interface{}{
							map[string]interface{}{
								"name": "user_id",
								"type": "string",
							},
							map[string]interface{}{
								"name":                  "categories",
								"type":                  "string",
								"multi_value_handling":  "sorted_array",
							},
						},
						"dimension_exclusions": []interface{}{"__time", "kafka.timestamp"},
					},
				},
				"metrics_spec": []interface{}{
					map[string]interface{}{
						"name": "count",
						"type": "count",
					},
					map[string]interface{}{
						"name":       "revenue",
						"type":       "doubleSum",
						"field_name": "price",
					},
				},
				"topic_pattern": "events-.*",
				"input_format": []interface{}{
					map[string]interface{}{
						"type": "json",
					},
				},
				"consumer_properties": map[string]interface{}{
					"bootstrap.servers": "kafka1:9092,kafka2:9092",
					"security.protocol": "SASL_SSL",
				},
				"task_count":           2,
				"replicas":            1,
				"task_duration":       "PT2H",
				"use_earliest_offset": true,
				"completion_timeout":  "PT1H",
				"context":             map[string]interface{}{"priority": "75"},
				"suspended":           false,
			},
			expected: map[string]interface{}{
				"type": "kafka",
				"spec": map[string]interface{}{
					"dataSchema": map[string]interface{}{
						"dataSource": "advanced-datasource",
						"timestampSpec": map[string]interface{}{
							"column":       "timestamp",
							"format":       "auto",
							"missingValue": "2010-01-01T00:00:00Z",
						},
						"dimensionsSpec": map[string]interface{}{
							"dimensions": []interface{}{
								map[string]interface{}{
									"name": "user_id",
									"type": "string",
								},
								map[string]interface{}{
									"name":                 "categories",
									"type":                 "string",
									"multiValueHandling":   "sorted_array",
								},
							},
							"dimensionExclusions": []interface{}{"__time", "kafka.timestamp"},
						},
						"metricsSpec": []interface{}{
							map[string]interface{}{
								"name": "count",
								"type": "count",
							},
							map[string]interface{}{
								"name":      "revenue",
								"type":      "doubleSum",
								"fieldName": "price",
							},
						},
					},
					"ioConfig": map[string]interface{}{
						"topicPattern": "events-.*",
						"inputFormat": map[string]interface{}{
							"type": "json",
						},
						"consumerProperties": map[string]interface{}{
							"bootstrap.servers": "kafka1:9092,kafka2:9092",
							"security.protocol": "SASL_SSL",
						},
						"taskCount":          2,
						"replicas":           1,
						"taskDuration":       "PT2H",
						"useEarliestOffset":  true,
						"completionTimeout":  "PT1H",
					},
					"context": map[string]interface{}{
						"priority": "75",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock resource data
			d := schema.TestResourceDataRaw(t, resourceKafkaSupervisor().Schema, tt.input)
			
			result, err := buildSupervisorSpec(d)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildDataSchema(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "basic data schema",
			input: map[string]interface{}{
				"datasource": "test-datasource",
				"timestamp_spec": []interface{}{
					map[string]interface{}{
						"column": "__time",
						"format": "iso",
					},
				},
			},
			expected: map[string]interface{}{
				"dataSource": "test-datasource",
				"timestampSpec": map[string]interface{}{
					"column": "__time",
					"format": "iso",
				},
			},
		},
		{
			name: "data schema with granularity spec",
			input: map[string]interface{}{
				"datasource": "test-datasource",
				"timestamp_spec": []interface{}{
					map[string]interface{}{
						"column": "__time",
						"format": "auto",
					},
				},
				"granularity_spec": []interface{}{
					map[string]interface{}{
						"type":                "uniform",
						"segment_granularity": "HOUR",
						"query_granularity":   "MINUTE",
						"rollup":              true,
					},
				},
			},
			expected: map[string]interface{}{
				"dataSource": "test-datasource",
				"timestampSpec": map[string]interface{}{
					"column": "__time",
					"format": "auto",
				},
				"granularitySpec": map[string]interface{}{
					"type":               "uniform",
					"segmentGranularity": "HOUR",
					"queryGranularity":   "MINUTE",
					"rollup":             true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceKafkaSupervisor().Schema, tt.input)
			result := buildDataSchema(d)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildIOConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "basic IO config with topic",
			input: map[string]interface{}{
				"topic": "test-topic",
				"input_format": []interface{}{
					map[string]interface{}{
						"type": "json",
					},
				},
				"consumer_properties": map[string]interface{}{
					"bootstrap.servers": "localhost:9092",
				},
				"task_count":           1,
				"replicas":            1,
				"task_duration":       "PT1H",
				"use_earliest_offset": false,
				"completion_timeout":  "PT30M",
			},
			expected: map[string]interface{}{
				"topic": "test-topic",
				"inputFormat": map[string]interface{}{
					"type": "json",
				},
				"consumerProperties": map[string]interface{}{
					"bootstrap.servers": "localhost:9092",
				},
				"taskCount":          1,
				"replicas":           1,
				"taskDuration":       "PT1H",
				"useEarliestOffset":  false,
				"completionTimeout":  "PT30M",
			},
		},
		{
			name: "IO config with topic pattern and idle config",
			input: map[string]interface{}{
				"topic_pattern": "events-.*",
				"input_format": []interface{}{
					map[string]interface{}{
						"type": "json",
					},
				},
				"consumer_properties": map[string]interface{}{
					"bootstrap.servers": "kafka1:9092,kafka2:9092",
				},
				"task_count":           2,
				"replicas":            1,
				"task_duration":       "PT2H",
				"use_earliest_offset": true,
				"completion_timeout":  "PT1H",
				"idle_config": []interface{}{
					map[string]interface{}{
						"enabled":               true,
						"inactive_after_millis": 600000,
					},
				},
			},
			expected: map[string]interface{}{
				"topicPattern": "events-.*",
				"inputFormat": map[string]interface{}{
					"type": "json",
				},
				"consumerProperties": map[string]interface{}{
					"bootstrap.servers": "kafka1:9092,kafka2:9092",
				},
				"taskCount":          2,
				"replicas":           1,
				"taskDuration":       "PT2H",
				"useEarliestOffset":  true,
				"completionTimeout":  "PT1H",
				"idleConfig": map[string]interface{}{
					"enabled":             true,
					"inactiveAfterMillis": 600000,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceKafkaSupervisor().Schema, tt.input)
			result := buildIOConfig(d)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildTuningConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "no tuning config",
			input: map[string]interface{}{
				"datasource": "test",
			},
			expected: nil,
		},
		{
			name: "basic tuning config",
			input: map[string]interface{}{
				"tuning_config": []interface{}{
					map[string]interface{}{
						"max_rows_per_segment": 5000000,
						"max_rows_in_memory":   150000,
					},
				},
			},
			expected: map[string]interface{}{
				"type":              "kafka",
				"maxRowsPerSegment": 5000000,
				"maxRowsInMemory":   150000,
			},
		},
		{
			name: "advanced tuning config with index spec",
			input: map[string]interface{}{
				"tuning_config": []interface{}{
					map[string]interface{}{
						"max_rows_per_segment":         2000000,
						"max_rows_in_memory":          100000,
						"max_bytes_in_memory":         134217728,
						"max_parse_exceptions":        100,
						"log_parse_exceptions":        true,
						"reset_offset_automatically":  true,
						"worker_threads":              4,
						"chat_retries":                5,
						"index_spec": []interface{}{
							map[string]interface{}{
								"bitmap": map[string]interface{}{
									"type": "roaring",
								},
								"dimension_compression": "lz4",
								"metric_compression":    "lz4",
							},
						},
					},
				},
			},
			expected: map[string]interface{}{
				"type":                     "kafka",
				"maxRowsPerSegment":        2000000,
				"maxRowsInMemory":          100000,
				"maxBytesInMemory":         134217728,
				"maxParseExceptions":       100,
				"logParseExceptions":       true,
				"resetOffsetAutomatically": true,
				"workerThreads":            4,
				"chatRetries":              5,
				"indexSpec": map[string]interface{}{
					"bitmap": map[string]interface{}{
						"type": "roaring",
					},
					"dimensionCompression": "lz4",
					"metricCompression":    "lz4",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceKafkaSupervisor().Schema, tt.input)
			result := buildTuningConfig(d)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceKafkaSupervisorSchema(t *testing.T) {
	resource := resourceKafkaSupervisor()
	
	// Test required fields
	assert.True(t, resource.Schema["datasource"].Required)
	assert.True(t, resource.Schema["timestamp_spec"].Required)
	assert.True(t, resource.Schema["input_format"].Required)
	assert.True(t, resource.Schema["consumer_properties"].Required)
	
	// Test computed fields
	assert.True(t, resource.Schema["supervisor_id"].Computed)
	assert.True(t, resource.Schema["state"].Computed)
	
	// Test optional fields with defaults
	assert.Equal(t, 1, resource.Schema["task_count"].Default)
	assert.Equal(t, 1, resource.Schema["replicas"].Default)
	assert.Equal(t, "PT1H", resource.Schema["task_duration"].Default)
	assert.Equal(t, false, resource.Schema["use_earliest_offset"].Default)
	assert.Equal(t, "PT30M", resource.Schema["completion_timeout"].Default)
	assert.Equal(t, false, resource.Schema["suspended"].Default)
	
	// Test mutually exclusive fields
	topicSchema := resource.Schema["topic"]
	topicPatternSchema := resource.Schema["topic_pattern"]
	
	assert.Contains(t, topicSchema.ExactlyOneOf, "topic")
	assert.Contains(t, topicSchema.ExactlyOneOf, "topic_pattern")
	assert.Contains(t, topicPatternSchema.ExactlyOneOf, "topic")
	assert.Contains(t, topicPatternSchema.ExactlyOneOf, "topic_pattern")
}

func TestConsumerPropertiesValidation(t *testing.T) {
	resource := resourceKafkaSupervisor()
	validateFunc := resource.Schema["consumer_properties"].ValidateFunc
	
	// Test valid consumer properties
	warnings, errors := validateFunc(map[string]interface{}{
		"bootstrap.servers": "localhost:9092",
		"group.id":         "test-group",
	}, "consumer_properties")
	
	assert.Empty(t, warnings)
	assert.Empty(t, errors)
	
	// Test missing bootstrap.servers
	warnings, errors = validateFunc(map[string]interface{}{
		"group.id": "test-group",
	}, "consumer_properties")
	
	assert.Empty(t, warnings)
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0].Error(), "bootstrap.servers is required")
}