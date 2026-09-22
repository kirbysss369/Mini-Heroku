package controlplane

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Node struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PublicHost string `json:"public_host"`
	Status     string `json:"status"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, requestTimeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

func (c *Client) Register(ctx context.Context, name, publicHost string) (Node, error) {
	body, err := json.Marshal(struct {
		Name       string `json:"name"`
		PublicHost string `json:"public_host"`
	}{Name: name, PublicHost: publicHost})
	if err != nil {
		return Node{}, fmt.Errorf("encode registration request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/internal/nodes/register",
		bytes.NewReader(body),
	)
	if err != nil {
		return Node{}, fmt.Errorf("create registration request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Node{}, fmt.Errorf("send registration request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Node{}, fmt.Errorf("registration returned HTTP %d", resp.StatusCode)
	}

	var node Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return Node{}, fmt.Errorf("decode registration response: %w", err)
	}
	if node.ID == "" {
		return Node{}, fmt.Errorf("registration response missing node id")
	}

	return node, nil
}

func (c *Client) Heartbeat(ctx context.Context, nodeID string) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/internal/nodes/"+nodeID+"/heartbeat",
		nil,
	)
	if err != nil {
		return fmt.Errorf("create heartbeat request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send heartbeat request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("heartbeat returned HTTP %d", resp.StatusCode)
	}

	return nil
}
