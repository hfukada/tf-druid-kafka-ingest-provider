package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateSupervisor(t *testing.T) {
	tests := []struct {
		name           string
		spec           map[string]interface{}
		responseStatus int
		responseBody   string
		expectedID     string
		expectError    bool
	}{
		{
			name: "successful creation",
			spec: map[string]interface{}{
				"type": "kafka",
				"spec": map[string]interface{}{
					"dataSchema": map[string]interface{}{
						"dataSource": "test",
					},
				},
			},
			responseStatus: http.StatusOK,
			responseBody:   `{"id": "test-supervisor-id"}`,
			expectedID:     "test-supervisor-id",
			expectError:    false,
		},
		{
			name: "server error",
			spec: map[string]interface{}{
				"type": "kafka",
			},
			responseStatus: http.StatusInternalServerError,
			responseBody:   `{"error": "Internal server error"}`,
			expectedID:     "",
			expectError:    true,
		},
		{
			name: "bad request",
			spec: map[string]interface{}{
				"type": "invalid",
			},
			responseStatus: http.StatusBadRequest,
			responseBody:   `{"error": "Invalid supervisor spec"}`,
			expectedID:     "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/druid/indexer/v1/supervisor", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				
				// Verify request body
				var requestBody map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				require.NoError(t, err)
				assert.Equal(t, tt.spec, requestBody)
				
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &Client{
				HTTPClient: server.Client(),
				Endpoint:   server.URL,
				Username:   "",
				Password:   "",
			}

			supervisorID, err := client.CreateSupervisor(context.Background(), tt.spec)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, supervisorID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, supervisorID)
			}
		})
	}
}

func TestClient_CreateSupervisorWithAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check basic auth
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "testuser", username)
		assert.Equal(t, "testpass", password)
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "test-supervisor-id"}`))
	}))
	defer server.Close()

	client := &Client{
		HTTPClient: server.Client(),
		Endpoint:   server.URL,
		Username:   "testuser",
		Password:   "testpass",
	}

	spec := map[string]interface{}{
		"type": "kafka",
		"spec": map[string]interface{}{
			"dataSchema": map[string]interface{}{
				"dataSource": "test",
			},
		},
	}

	supervisorID, err := client.CreateSupervisor(context.Background(), spec)
	assert.NoError(t, err)
	assert.Equal(t, "test-supervisor-id", supervisorID)
}

func TestClient_GetSupervisor(t *testing.T) {
	tests := []struct {
		name             string
		supervisorID     string
		responseStatus   int
		responseBody     string
		expectedStatus   *SupervisorStatus
		expectError      bool
	}{
		{
			name:           "successful get",
			supervisorID:   "test-supervisor",
			responseStatus: http.StatusOK,
			responseBody:   `{"id": "test-supervisor", "state": "RUNNING"}`,
			expectedStatus: &SupervisorStatus{
				ID:    "test-supervisor",
				State: "RUNNING",
			},
			expectError: false,
		},
		{
			name:           "supervisor not found",
			supervisorID:   "nonexistent-supervisor",
			responseStatus: http.StatusNotFound,
			responseBody:   `{"error": "Not found"}`,
			expectedStatus: nil,
			expectError:    false,
		},
		{
			name:           "server error",
			supervisorID:   "test-supervisor",
			responseStatus: http.StatusInternalServerError,
			responseBody:   `{"error": "Internal server error"}`,
			expectedStatus: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/druid/indexer/v1/supervisor/" + tt.supervisorID + "/status"
				assert.Equal(t, expectedPath, r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &Client{
				HTTPClient: server.Client(),
				Endpoint:   server.URL,
				Username:   "",
				Password:   "",
			}

			status, err := client.GetSupervisor(context.Background(), tt.supervisorID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, status)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, status)
			}
		})
	}
}

func TestClient_DeleteSupervisor(t *testing.T) {
	tests := []struct {
		name           string
		supervisorID   string
		responseStatus int
		responseBody   string
		expectError    bool
	}{
		{
			name:           "successful deletion",
			supervisorID:   "test-supervisor",
			responseStatus: http.StatusOK,
			responseBody:   `{"status": "success"}`,
			expectError:    false,
		},
		{
			name:           "supervisor not found - should not error",
			supervisorID:   "nonexistent-supervisor",
			responseStatus: http.StatusNotFound,
			responseBody:   `{"error": "Not found"}`,
			expectError:    false,
		},
		{
			name:           "server error",
			supervisorID:   "test-supervisor",
			responseStatus: http.StatusInternalServerError,
			responseBody:   `{"error": "Internal server error"}`,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/druid/indexer/v1/supervisor/" + tt.supervisorID + "/terminate"
				assert.Equal(t, expectedPath, r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)
				
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &Client{
				HTTPClient: server.Client(),
				Endpoint:   server.URL,
				Username:   "",
				Password:   "",
			}

			err := client.DeleteSupervisor(context.Background(), tt.supervisorID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_SuspendSupervisor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/druid/indexer/v1/supervisor/test-supervisor/suspend", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success"}`))
	}))
	defer server.Close()

	client := &Client{
		HTTPClient: server.Client(),
		Endpoint:   server.URL,
		Username:   "",
		Password:   "",
	}

	err := client.SuspendSupervisor(context.Background(), "test-supervisor")
	assert.NoError(t, err)
}

func TestClient_ResumeSupervisor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/druid/indexer/v1/supervisor/test-supervisor/resume", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success"}`))
	}))
	defer server.Close()

	client := &Client{
		HTTPClient: server.Client(),
		Endpoint:   server.URL,
		Username:   "",
		Password:   "",
	}

	err := client.ResumeSupervisor(context.Background(), "test-supervisor")
	assert.NoError(t, err)
}