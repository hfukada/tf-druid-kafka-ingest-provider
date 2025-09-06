package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type SupervisorStatus struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

type SupervisorResponse struct {
	ID string `json:"id"`
}

func (c *Client) CreateSupervisor(ctx context.Context, spec map[string]interface{}) (string, error) {
	endpoint, err := url.JoinPath(c.Endpoint, "/druid/indexer/v1/supervisor")
	if err != nil {
		return "", fmt.Errorf("failed to construct endpoint URL: %w", err)
	}

	specBytes, err := json.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal supervisor spec: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(specBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("Druid API error (status %d): %s", resp.StatusCode, string(body))
	}

	var supervisorResp SupervisorResponse
	if err := json.Unmarshal(body, &supervisorResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal supervisor response: %w", err)
	}

	return supervisorResp.ID, nil
}

func (c *Client) GetSupervisor(ctx context.Context, supervisorID string) (*SupervisorStatus, error) {
	endpoint, err := url.JoinPath(c.Endpoint, "/druid/indexer/v1/supervisor", supervisorID, "status")
	if err != nil {
		return nil, fmt.Errorf("failed to construct endpoint URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Druid API error (status %d): %s", resp.StatusCode, string(body))
	}

	var status SupervisorStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal supervisor status: %w", err)
	}

	return &status, nil
}

func (c *Client) DeleteSupervisor(ctx context.Context, supervisorID string) error {
	endpoint, err := url.JoinPath(c.Endpoint, "/druid/indexer/v1/supervisor", supervisorID, "terminate")
	if err != nil {
		return fmt.Errorf("failed to construct endpoint URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Supervisor already doesn't exist, consider it successfully deleted
		return nil
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Druid API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) SuspendSupervisor(ctx context.Context, supervisorID string) error {
	endpoint, err := url.JoinPath(c.Endpoint, "/druid/indexer/v1/supervisor", supervisorID, "suspend")
	if err != nil {
		return fmt.Errorf("failed to construct endpoint URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Druid API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) ResumeSupervisor(ctx context.Context, supervisorID string) error {
	endpoint, err := url.JoinPath(c.Endpoint, "/druid/indexer/v1/supervisor", supervisorID, "resume")
	if err != nil {
		return fmt.Errorf("failed to construct endpoint URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Druid API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}