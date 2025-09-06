package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKafkaSupervisor() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Druid Kafka ingestion supervisor",
		
		CreateContext: resourceKafkaSupervisorCreate,
		ReadContext:   resourceKafkaSupervisorRead,
		UpdateContext: resourceKafkaSupervisorUpdate,
		DeleteContext: resourceKafkaSupervisorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"datasource": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Druid datasource",
			},
			
			// Data Schema Configuration
			"timestamp_spec": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Timestamp specification",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"column": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the timestamp column",
						},
						"format": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "auto",
							Description: "Timestamp format (e.g., 'iso', 'millis', 'auto')",
						},
						"missing_value": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Default timestamp for rows with missing timestamps",
						},
					},
				},
			},
			
			"dimensions_spec": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Dimension specification",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dimensions": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of dimension specifications",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Dimension name",
									},
									"type": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "string",
										Description: "Dimension type (string, long, float, double)",
									},
									"multi_value_handling": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "How to handle multi-value dimensions",
									},
								},
							},
						},
						"dimension_exclusions": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of columns to exclude from dimensions",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"spatial_dimensions": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Spatial dimension configurations",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dim_name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Name of the spatial dimension",
									},
									"dims": {
										Type:        schema.TypeList,
										Required:    true,
										Description: "List of coordinate columns",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			
			"metrics_spec": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of aggregation metrics",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Metric name",
						},
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Aggregation type (count, longSum, doubleSum, etc.)",
						},
						"field_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Input field name for the metric",
						},
					},
				},
			},
			
			"granularity_spec": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Granularity specification",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "uniform",
							Description: "Granularity type (uniform or arbitrary)",
						},
						"segment_granularity": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "HOUR",
							Description: "Segment granularity (HOUR, DAY, WEEK, etc.)",
						},
						"query_granularity": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "NONE",
							Description: "Query granularity (NONE, SECOND, MINUTE, etc.)",
						},
						"rollup": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether to enable rollup",
						},
					},
				},
			},
			
			// IO Configuration
			"topic": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"topic", "topic_pattern"},
				Description:  "Kafka topic to consume from",
			},
			
			"topic_pattern": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"topic", "topic_pattern"},
				Description:  "Regex pattern for multiple Kafka topics",
			},
			
			"input_format": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Input format configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Input format type (json, csv, tsv, etc.)",
						},
						"flat_spec": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Flat spec for delimited formats",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"use_field_discovery": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     true,
										Description: "Whether to use field discovery",
									},
									"delimiter": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Field delimiter",
									},
									"columns": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "Column names",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			
			"consumer_properties": {
				Type:        schema.TypeMap,
				Required:    true,
				Description: "Kafka consumer properties",
				Elem:        &schema.Schema{Type: schema.TypeString},
				ValidateFunc: func(v interface{}, k string) (warnings []string, errors []error) {
					props := v.(map[string]interface{})
					if _, ok := props["bootstrap.servers"]; !ok {
						errors = append(errors, fmt.Errorf("bootstrap.servers is required in consumer_properties"))
					}
					return warnings, errors
				},
			},
			
			"task_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				Description:  "Number of reading tasks",
				ValidateFunc: validation.IntAtLeast(1),
			},
			
			"replicas": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				Description:  "Number of replica tasks",
				ValidateFunc: validation.IntAtLeast(1),
			},
			
			"task_duration": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "PT1H",
				Description: "Task duration in ISO 8601 format",
			},
			
			"use_earliest_offset": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to use earliest offset for new topics",
			},
			
			"completion_timeout": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "PT30M",
				Description: "Task completion timeout in ISO 8601 format",
			},
			
			"idle_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Idle configuration for the supervisor",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether idle detection is enabled",
						},
						"inactive_after_millis": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Milliseconds of inactivity before going idle",
						},
					},
				},
			},
			
			// Tuning Configuration
			"tuning_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Tuning configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_rows_per_segment": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     5000000,
							Description: "Maximum number of rows per segment",
						},
						"max_rows_in_memory": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     150000,
							Description: "Maximum rows in memory before persisting",
						},
						"max_bytes_in_memory": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Maximum bytes in memory before persisting",
						},
						"skip_bytes_in_memory_overhead_check": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Skip overhead check for bytes in memory",
						},
						"max_pending_persists": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Maximum pending persist operations",
						},
						"index_spec": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Index specification",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bitmap": {
										Type:        schema.TypeMap,
										Optional:    true,
										Description: "Bitmap index configuration",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
									"dimension_compression": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "lz4",
										Description: "Dimension compression algorithm",
									},
									"metric_compression": {
										Type:        schema.TypeString,
										Optional:     true,
										Default:     "lz4",
										Description: "Metric compression algorithm",
									},
								},
							},
						},
						"segment_write_out_medium_factory": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Segment write-out medium factory configuration",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"intermediate_persist_period": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "PT10M",
							Description: "Intermediate persist period",
						},
						"max_parse_exceptions": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     2147483647,
							Description: "Maximum parse exceptions allowed",
						},
						"max_saved_parse_exceptions": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Maximum saved parse exceptions",
						},
						"log_parse_exceptions": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether to log parse exceptions",
						},
						"reset_offset_automatically": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Automatically reset consumer offset on errors",
						},
						"worker_threads": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Number of worker threads",
							ValidateFunc: validation.IntAtLeast(1),
						},
						"chat_threads": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Number of chat handler threads",
							ValidateFunc: validation.IntAtLeast(1),
						},
						"chat_retries": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      8,
							Description:  "Number of chat retries",
							ValidateFunc: validation.IntAtLeast(0),
						},
						"http_timeout": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "PT10S",
							Description: "HTTP timeout for task communication",
						},
						"shutdown_timeout": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "PT80S",
							Description: "Task shutdown timeout",
						},
					},
				},
			},
			
			"context": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Additional context parameters",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			
			"suspended": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the supervisor is suspended",
			},
			
			// Computed fields
			"supervisor_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The supervisor ID returned by Druid",
			},
			
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current state of the supervisor",
			},
		},
	}
}

func resourceKafkaSupervisorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	
	spec, err := buildSupervisorSpec(d)
	if err != nil {
		return diag.FromErr(err)
	}
	
	supervisorID, err := client.CreateSupervisor(ctx, spec)
	if err != nil {
		return diag.FromErr(err)
	}
	
	d.SetId(supervisorID)
	d.Set("supervisor_id", supervisorID)
	
	return resourceKafkaSupervisorRead(ctx, d, meta)
}

func resourceKafkaSupervisorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	
	supervisor, err := client.GetSupervisor(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	
	if supervisor == nil {
		d.SetId("")
		return nil
	}
	
	d.Set("state", supervisor.State)
	d.Set("supervisor_id", supervisor.ID)
	
	return nil
}

func resourceKafkaSupervisorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	
	spec, err := buildSupervisorSpec(d)
	if err != nil {
		return diag.FromErr(err)
	}
	
	_, err = client.CreateSupervisor(ctx, spec)
	if err != nil {
		return diag.FromErr(err)
	}
	
	return resourceKafkaSupervisorRead(ctx, d, meta)
}

func resourceKafkaSupervisorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	
	err := client.DeleteSupervisor(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	
	d.SetId("")
	return nil
}

func buildSupervisorSpec(d *schema.ResourceData) (map[string]interface{}, error) {
	spec := map[string]interface{}{
		"type": "kafka",
		"spec": map[string]interface{}{
			"dataSchema": buildDataSchema(d),
			"ioConfig":   buildIOConfig(d),
		},
	}
	
	if tuningConfig := buildTuningConfig(d); tuningConfig != nil {
		spec["spec"].(map[string]interface{})["tuningConfig"] = tuningConfig
	}
	
	if context := d.Get("context").(map[string]interface{}); len(context) > 0 {
		spec["spec"].(map[string]interface{})["context"] = context
	}
	
	if suspended := d.Get("suspended").(bool); suspended {
		spec["spec"].(map[string]interface{})["suspended"] = suspended
	}
	
	return spec, nil
}

func buildDataSchema(d *schema.ResourceData) map[string]interface{} {
	dataSchema := map[string]interface{}{
		"dataSource": d.Get("datasource").(string),
	}
	
	// Timestamp spec
	if timestampSpecs := d.Get("timestamp_spec").([]interface{}); len(timestampSpecs) > 0 {
		timestampSpec := timestampSpecs[0].(map[string]interface{})
		ts := map[string]interface{}{
			"column": timestampSpec["column"].(string),
			"format": timestampSpec["format"].(string),
		}
		if missingValue := timestampSpec["missing_value"].(string); missingValue != "" {
			ts["missingValue"] = missingValue
		}
		dataSchema["timestampSpec"] = ts
	}
	
	// Dimensions spec
	if dimensionsSpecs := d.Get("dimensions_spec").([]interface{}); len(dimensionsSpecs) > 0 {
		dimensionsSpec := dimensionsSpecs[0].(map[string]interface{})
		ds := map[string]interface{}{}
		
		if dimensions := dimensionsSpec["dimensions"].([]interface{}); len(dimensions) > 0 {
			var dims []interface{}
			for _, dim := range dimensions {
				dimMap := dim.(map[string]interface{})
				d := map[string]interface{}{
					"name": dimMap["name"].(string),
					"type": dimMap["type"].(string),
				}
				if mvh := dimMap["multi_value_handling"].(string); mvh != "" {
					d["multiValueHandling"] = mvh
				}
				dims = append(dims, d)
			}
			ds["dimensions"] = dims
		}
		
		if exclusions := dimensionsSpec["dimension_exclusions"].([]interface{}); len(exclusions) > 0 {
			ds["dimensionExclusions"] = exclusions
		}
		
		if spatialDims := dimensionsSpec["spatial_dimensions"].([]interface{}); len(spatialDims) > 0 {
			var spatialDimensions []interface{}
			for _, spatial := range spatialDims {
				spatialMap := spatial.(map[string]interface{})
				spatialDimensions = append(spatialDimensions, map[string]interface{}{
					"dimName": spatialMap["dim_name"].(string),
					"dims":    spatialMap["dims"].([]interface{}),
				})
			}
			ds["spatialDimensions"] = spatialDimensions
		}
		
		if len(ds) > 0 {
			dataSchema["dimensionsSpec"] = ds
		}
	}
	
	// Metrics spec
	if metricsSpecs := d.Get("metrics_spec").([]interface{}); len(metricsSpecs) > 0 {
		var metrics []interface{}
		for _, metric := range metricsSpecs {
			metricMap := metric.(map[string]interface{})
			m := map[string]interface{}{
				"name": metricMap["name"].(string),
				"type": metricMap["type"].(string),
			}
			if fieldName := metricMap["field_name"].(string); fieldName != "" {
				m["fieldName"] = fieldName
			}
			metrics = append(metrics, m)
		}
		dataSchema["metricsSpec"] = metrics
	}
	
	// Granularity spec
	if granularitySpecs := d.Get("granularity_spec").([]interface{}); len(granularitySpecs) > 0 {
		granularitySpec := granularitySpecs[0].(map[string]interface{})
		gs := map[string]interface{}{
			"type":               granularitySpec["type"].(string),
			"segmentGranularity": granularitySpec["segment_granularity"].(string),
			"queryGranularity":   granularitySpec["query_granularity"].(string),
			"rollup":             granularitySpec["rollup"].(bool),
		}
		dataSchema["granularitySpec"] = gs
	}
	
	return dataSchema
}

func buildIOConfig(d *schema.ResourceData) map[string]interface{} {
	ioConfig := map[string]interface{}{}
	
	// Topic or topic pattern
	if topic := d.Get("topic").(string); topic != "" {
		ioConfig["topic"] = topic
	}
	if topicPattern := d.Get("topic_pattern").(string); topicPattern != "" {
		ioConfig["topicPattern"] = topicPattern
	}
	
	// Input format
	if inputFormats := d.Get("input_format").([]interface{}); len(inputFormats) > 0 {
		inputFormat := inputFormats[0].(map[string]interface{})
		inputFormatConfig := map[string]interface{}{
			"type": inputFormat["type"].(string),
		}
		
		if flatSpecs := inputFormat["flat_spec"].([]interface{}); len(flatSpecs) > 0 {
			flatSpec := flatSpecs[0].(map[string]interface{})
			fs := map[string]interface{}{
				"useFieldDiscovery": flatSpec["use_field_discovery"].(bool),
			}
			if delimiter := flatSpec["delimiter"].(string); delimiter != "" {
				fs["delimiter"] = delimiter
			}
			if columns := flatSpec["columns"].([]interface{}); len(columns) > 0 {
				fs["columns"] = columns
			}
			inputFormatConfig["flatSpec"] = fs
		}
		
		ioConfig["inputFormat"] = inputFormatConfig
	}
	
	// Consumer properties
	if consumerProps := d.Get("consumer_properties").(map[string]interface{}); len(consumerProps) > 0 {
		ioConfig["consumerProperties"] = consumerProps
	}
	
	// Task configuration
	ioConfig["taskCount"] = d.Get("task_count").(int)
	ioConfig["replicas"] = d.Get("replicas").(int)
	ioConfig["taskDuration"] = d.Get("task_duration").(string)
	ioConfig["useEarliestOffset"] = d.Get("use_earliest_offset").(bool)
	ioConfig["completionTimeout"] = d.Get("completion_timeout").(string)
	
	// Idle configuration
	if idleConfigs := d.Get("idle_config").([]interface{}); len(idleConfigs) > 0 {
		idleConfig := idleConfigs[0].(map[string]interface{})
		ic := map[string]interface{}{
			"enabled": idleConfig["enabled"].(bool),
		}
		if inactiveAfter := idleConfig["inactive_after_millis"].(int); inactiveAfter > 0 {
			ic["inactiveAfterMillis"] = inactiveAfter
		}
		ioConfig["idleConfig"] = ic
	}
	
	return ioConfig
}

func buildTuningConfig(d *schema.ResourceData) map[string]interface{} {
	if tuningConfigs := d.Get("tuning_config").([]interface{}); len(tuningConfigs) > 0 {
		tuningConfig := tuningConfigs[0].(map[string]interface{})
		tc := map[string]interface{}{
			"type": "kafka",
		}
		
		if maxRows := tuningConfig["max_rows_per_segment"].(int); maxRows > 0 {
			tc["maxRowsPerSegment"] = maxRows
		}
		if maxRowsInMemory := tuningConfig["max_rows_in_memory"].(int); maxRowsInMemory > 0 {
			tc["maxRowsInMemory"] = maxRowsInMemory
		}
		if maxBytesInMemory := tuningConfig["max_bytes_in_memory"].(int); maxBytesInMemory > 0 {
			tc["maxBytesInMemory"] = maxBytesInMemory
		}
		if skipCheck := tuningConfig["skip_bytes_in_memory_overhead_check"].(bool); skipCheck {
			tc["skipBytesInMemoryOverheadCheck"] = skipCheck
		}
		if maxPending := tuningConfig["max_pending_persists"].(int); maxPending > 0 {
			tc["maxPendingPersists"] = maxPending
		}
		if intermediatePersist := tuningConfig["intermediate_persist_period"].(string); intermediatePersist != "" {
			tc["intermediatePersistPeriod"] = intermediatePersist
		}
		if maxParseExceptions := tuningConfig["max_parse_exceptions"].(int); maxParseExceptions > 0 {
			tc["maxParseExceptions"] = maxParseExceptions
		}
		if maxSavedExceptions := tuningConfig["max_saved_parse_exceptions"].(int); maxSavedExceptions > 0 {
			tc["maxSavedParseExceptions"] = maxSavedExceptions
		}
		if logExceptions := tuningConfig["log_parse_exceptions"].(bool); logExceptions {
			tc["logParseExceptions"] = logExceptions
		}
		if resetOffset := tuningConfig["reset_offset_automatically"].(bool); resetOffset {
			tc["resetOffsetAutomatically"] = resetOffset
		}
		if workerThreads := tuningConfig["worker_threads"].(int); workerThreads > 0 {
			tc["workerThreads"] = workerThreads
		}
		if chatThreads := tuningConfig["chat_threads"].(int); chatThreads > 0 {
			tc["chatThreads"] = chatThreads
		}
		if chatRetries := tuningConfig["chat_retries"].(int); chatRetries > 0 {
			tc["chatRetries"] = chatRetries
		}
		if httpTimeout := tuningConfig["http_timeout"].(string); httpTimeout != "" {
			tc["httpTimeout"] = httpTimeout
		}
		if shutdownTimeout := tuningConfig["shutdown_timeout"].(string); shutdownTimeout != "" {
			tc["shutdownTimeout"] = shutdownTimeout
		}
		
		// Index spec
		if indexSpecs := tuningConfig["index_spec"].([]interface{}); len(indexSpecs) > 0 {
			indexSpec := indexSpecs[0].(map[string]interface{})
			is := map[string]interface{}{}
			
			if bitmap := indexSpec["bitmap"].(map[string]interface{}); len(bitmap) > 0 {
				is["bitmap"] = bitmap
			}
			if dimCompression := indexSpec["dimension_compression"].(string); dimCompression != "" {
				is["dimensionCompression"] = dimCompression
			}
			if metricCompression := indexSpec["metric_compression"].(string); metricCompression != "" {
				is["metricCompression"] = metricCompression
			}
			
			if len(is) > 0 {
				tc["indexSpec"] = is
			}
		}
		
		if segmentWriteOut := tuningConfig["segment_write_out_medium_factory"].(map[string]interface{}); len(segmentWriteOut) > 0 {
			tc["segmentWriteOutMediumFactory"] = segmentWriteOut
		}
		
		return tc
	}
	
	return nil
}