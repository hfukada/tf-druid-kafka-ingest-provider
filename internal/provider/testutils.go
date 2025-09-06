package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// MockDruidServer provides a mock Druid server for testing
type MockDruidServer struct {
	server      *httptest.Server
	supervisors map[string]*SupervisorStatus
	mutex       sync.RWMutex
}

// NewMockDruidServer creates a new mock Druid server
func NewMockDruidServer() *MockDruidServer {
	mock := &MockDruidServer{
		supervisors: make(map[string]*SupervisorStatus),
	}
	
	mux := http.NewServeMux()
	
	// Create or update supervisor
	mux.HandleFunc("/druid/indexer/v1/supervisor", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		var spec map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		// Extract datasource from spec for ID generation
		supervisorID := "test-supervisor-id"
		if specData, ok := spec["spec"].(map[string]interface{}); ok {
			if dataSchema, ok := specData["dataSchema"].(map[string]interface{}); ok {
				if datasource, ok := dataSchema["dataSource"].(string); ok {
					supervisorID = datasource + "-supervisor"
				}
			}
		}
		
		mock.mutex.Lock()
		mock.supervisors[supervisorID] = &SupervisorStatus{
			ID:    supervisorID,
			State: "RUNNING",
		}
		mock.mutex.Unlock()
		
		response := SupervisorResponse{ID: supervisorID}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	// Get supervisor status
	mux.HandleFunc("/druid/indexer/v1/supervisor/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		// Extract supervisor ID from URL
		path := strings.TrimPrefix(r.URL.Path, "/druid/indexer/v1/supervisor/")
		supervisorID := strings.TrimSuffix(path, "/status")
		
		if !strings.HasSuffix(path, "/status") {
			// Handle other endpoints like suspend, resume, terminate
			if strings.HasSuffix(path, "/suspend") || strings.HasSuffix(path, "/resume") {
				if r.Method != http.MethodPost {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}
				supervisorID = strings.TrimSuffix(supervisorID, "/suspend")
				supervisorID = strings.TrimSuffix(supervisorID, "/resume")
				
				mock.mutex.Lock()
				if supervisor, exists := mock.supervisors[supervisorID]; exists {
					if strings.HasSuffix(path, "/suspend") {
						supervisor.State = "SUSPENDED"
					} else {
						supervisor.State = "RUNNING"
					}
				}
				mock.mutex.Unlock()
				
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "success"}`))
				return
			}
			
			if strings.HasSuffix(path, "/terminate") {
				if r.Method != http.MethodPost {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}
				supervisorID = strings.TrimSuffix(supervisorID, "/terminate")
				
				mock.mutex.Lock()
				delete(mock.supervisors, supervisorID)
				mock.mutex.Unlock()
				
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "success"}`))
				return
			}
		}
		
		mock.mutex.RLock()
		supervisor, exists := mock.supervisors[supervisorID]
		mock.mutex.RUnlock()
		
		if !exists {
			http.Error(w, "Supervisor not found", http.StatusNotFound)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(supervisor)
	})
	
	mock.server = httptest.NewServer(mux)
	return mock
}

// URL returns the mock server URL
func (m *MockDruidServer) URL() string {
	return m.server.URL
}

// Close closes the mock server
func (m *MockDruidServer) Close() {
	m.server.Close()
}

// GetSupervisor returns a supervisor by ID
func (m *MockDruidServer) GetSupervisor(id string) *SupervisorStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.supervisors[id]
}

// AddSupervisor adds a supervisor to the mock server
func (m *MockDruidServer) AddSupervisor(id, state string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.supervisors[id] = &SupervisorStatus{
		ID:    id,
		State: state,
	}
}

// ClearSupervisors removes all supervisors
func (m *MockDruidServer) ClearSupervisors() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.supervisors = make(map[string]*SupervisorStatus)
}