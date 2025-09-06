package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccKafkaSupervisor_basic(t *testing.T) {
	mockServer := NewMockDruidServer()
	defer mockServer.Close()

	datasource := acctest.RandomWithPrefix("test-datasource")
	
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(mockServer.URL()),
		CheckDestroy:      testAccCheckKafkaSupervisorDestroy(mockServer),
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaSupervisorConfig_basic(mockServer.URL(), datasource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKafkaSupervisorExists("druid_kafka_supervisor.test", mockServer),
					resource.TestCheckResourceAttr("druid_kafka_supervisor.test", "datasource", datasource),
					resource.TestCheckResourceAttr("druid_kafka_supervisor.test", "topic", "test-topic"),
					resource.TestCheckResourceAttr("druid_kafka_supervisor.test", "task_count", "1"),
					resource.TestCheckResourceAttr("druid_kafka_supervisor.test", "replicas", "1"),
					resource.TestCheckResourceAttrSet("druid_kafka_supervisor.test", "supervisor_id"),
					resource.TestCheckResourceAttrSet("druid_kafka_supervisor.test", "state"),
				),
			},
		},
	})
}

func TestAccKafkaSupervisor_update(t *testing.T) {
	mockServer := NewMockDruidServer()
	defer mockServer.Close()

	datasource := acctest.RandomWithPrefix("test-datasource")
	
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(mockServer.URL()),
		CheckDestroy:      testAccCheckKafkaSupervisorDestroy(mockServer),
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaSupervisorConfig_basic(mockServer.URL(), datasource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKafkaSupervisorExists("druid_kafka_supervisor.test", mockServer),
					resource.TestCheckResourceAttr("druid_kafka_supervisor.test", "task_count", "1"),
				),
			},
			{
				Config: testAccKafkaSupervisorConfig_updated(mockServer.URL(), datasource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKafkaSupervisorExists("druid_kafka_supervisor.test", mockServer),
					resource.TestCheckResourceAttr("druid_kafka_supervisor.test", "task_count", "2"),
					resource.TestCheckResourceAttr("druid_kafka_supervisor.test", "replicas", "2"),
				),
			},
		},
	})
}

func TestAccKafkaSupervisor_advanced(t *testing.T) {
	mockServer := NewMockDruidServer()
	defer mockServer.Close()

	datasource := acctest.RandomWithPrefix("test-datasource")
	
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders(mockServer.URL()),
		CheckDestroy:      testAccCheckKafkaSupervisorDestroy(mockServer),
		Steps: []resource.TestStep{
			{
				Config: testAccKafkaSupervisorConfig_advanced(mockServer.URL(), datasource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKafkaSupervisorExists("druid_kafka_supervisor.test", mockServer),
					resource.TestCheckResourceAttr("druid_kafka_supervisor.test", "datasource", datasource),
					resource.TestCheckResourceAttr("druid_kafka_supervisor.test", "topic_pattern", "events-.*"),
					resource.TestCheckResourceAttr("druid_kafka_supervisor.test", "task_count", "3"),
					resource.TestCheckResourceAttr("druid_kafka_supervisor.test", "replicas", "2"),
					resource.TestCheckResourceAttr("druid_kafka_supervisor.test", "dimensions_spec.0.dimensions.0.name", "product_id"),
					resource.TestCheckResourceAttr("druid_kafka_supervisor.test", "metrics_spec.0.name", "count"),
					resource.TestCheckResourceAttr("druid_kafka_supervisor.test", "tuning_config.0.max_rows_per_segment", "2000000"),
				),
			},
		},
	})
}

func testAccPreCheck(t *testing.T) {
	// Add any pre-check logic here
}

func testAccProviders(endpoint string) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"druid": func() (*schema.Provider, error) {
			p := New()
			// Override the default endpoint for testing
			p.Schema["endpoint"].Default = endpoint
			p.Schema["endpoint"].DefaultFunc = nil
			return p, nil
		},
	}
}

func testAccCheckKafkaSupervisorExists(resourceName string, mockServer *MockDruidServer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set")
		}

		supervisor := mockServer.GetSupervisor(rs.Primary.ID)
		if supervisor == nil {
			return fmt.Errorf("supervisor %s not found in mock server", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckKafkaSupervisorDestroy(mockServer *MockDruidServer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "druid_kafka_supervisor" {
				continue
			}

			supervisor := mockServer.GetSupervisor(rs.Primary.ID)
			if supervisor != nil {
				return fmt.Errorf("supervisor %s still exists in mock server", rs.Primary.ID)
			}
		}
		return nil
	}
}

func testAccKafkaSupervisorConfig_basic(endpoint, datasource string) string {
	return fmt.Sprintf(`
provider "druid" {
  endpoint = "%s"
}

resource "druid_kafka_supervisor" "test" {
  datasource = "%s"

  timestamp_spec {
    column = "__time"
    format = "iso"
  }

  topic = "test-topic"

  input_format {
    type = "json"
  }

  consumer_properties = {
    "bootstrap.servers" = "localhost:9092"
  }

  task_count = 1
  replicas   = 1
}
`, endpoint, datasource)
}

func testAccKafkaSupervisorConfig_updated(endpoint, datasource string) string {
	return fmt.Sprintf(`
provider "druid" {
  endpoint = "%s"
}

resource "druid_kafka_supervisor" "test" {
  datasource = "%s"

  timestamp_spec {
    column = "__time"
    format = "iso"
  }

  topic = "test-topic"

  input_format {
    type = "json"
  }

  consumer_properties = {
    "bootstrap.servers" = "localhost:9092"
  }

  task_count = 2
  replicas   = 2
}
`, endpoint, datasource)
}

func testAccKafkaSupervisorConfig_advanced(endpoint, datasource string) string {
	return fmt.Sprintf(`
provider "druid" {
  endpoint = "%s"
}

resource "druid_kafka_supervisor" "test" {
  datasource = "%s"

  timestamp_spec {
    column        = "timestamp"
    format        = "auto"
    missing_value = "2010-01-01T00:00:00Z"
  }

  dimensions_spec {
    dimensions {
      name = "product_id"
      type = "string"
    }
    
    dimensions {
      name                  = "category"
      type                  = "string"
      multi_value_handling  = "sorted_array"
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

  topic_pattern = "events-.*"

  input_format {
    type = "json"
  }

  consumer_properties = {
    "bootstrap.servers" = "kafka1:9092,kafka2:9092"
    "security.protocol" = "SASL_SSL"
  }

  task_count = 3
  replicas   = 2

  tuning_config {
    max_rows_per_segment = 2000000
    max_rows_in_memory   = 100000
    log_parse_exceptions = true
  }
}
`, endpoint, datasource)
}