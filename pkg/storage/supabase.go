package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client uploads/deletes files on Supabase Storage.
type Client struct {
	baseURL    string
	serviceKey string
	bucket     string
}

func NewSupabaseClient(baseURL, serviceKey, bucket string) *Client {
	return &Client{baseURL: baseURL, serviceKey: serviceKey, bucket: bucket}
}

// Upload sends a file to Supabase Storage at the given path and returns the public URL.
func (c *Client) Upload(path string, data []byte, contentType string) (string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.baseURL, c.bucket, path)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.serviceKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("supabase upload failed (%d): %s", resp.StatusCode, string(body))
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", c.baseURL, c.bucket, path)
	return publicURL, nil
}

// Delete removes files at the given paths from Supabase Storage.
func (c *Client) Delete(paths []string) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s", c.baseURL, c.bucket)
	body, _ := json.Marshal(map[string][]string{"prefixes": paths})
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.serviceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase delete failed (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}
