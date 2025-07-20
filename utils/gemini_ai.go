/*
Copyright © 2024 Abhishek M. Yadav <abhishekyadav@duck.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GeminiRequestPayload defines the structure for the API request body.
type GeminiRequestPayload struct {
	Contents          []Content          `json:"contents"`
	SystemInstruction *SystemInstruction `json:"systemInstruction,omitempty"`
	GenerationConfig  *GenerationConfig  `json:"generationConfig,omitempty"`
}

// Content represents the content parts of the request.
type Content struct {
	Parts []Part `json:"parts"`
}

// Part represents a single part of the content.
type Part struct {
	Text string `json:"text"`
}

// SystemInstruction defines the system-level instructions for the model.
type SystemInstruction struct {
	Parts []Part `json:"parts"`
}

// GenerationConfig configures the model's generation behavior.
type GenerationConfig struct {
	Temperature float32 `json:"temperature"`
}

// GeminiResponsePayload defines the structure for the API response body.
type GeminiResponsePayload struct {
	Candidates []Candidate `json:"candidates"`
	Error      *APIError   `json:"error,omitempty"`
}

// Candidate represents a single response candidate.
type Candidate struct {
	Content Content `json:"content"`
}

// APIError represents an error response from the Gemini API.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("gemini api error: %s (code: %d, status: %s)", e.Message, e.Code, e.Status)
}

// Client holds the configuration for making direct API calls to Gemini.
type Client struct {
	config     GeminiConfig
	httpClient *http.Client
}

// NewAIClient creates a new Gemini AI client from the provided configuration.
func NewAIClient(config GeminiConfig) (*Client, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("❌ Gemini API key is not set.\n" +
			"Please run 'pancake edit config' to open the configuration file and add your API key.\n\n" +
			"Example configuration in pancake.yml:\n" +
			"gemini:\n" +
			"  api_key: \"YOUR_API_KEY_HERE\"\n" +
			"  url: \"https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent\"\n" +
			"  temperature: 0.7\n" +
			"  context: \"You are a helpful assistant that translates natural language into executable shell commands...\"")
	}
	if config.URL == "" {
		return nil, fmt.Errorf("❌ Gemini API URL is not set.\n" +
			"Please run 'pancake edit config' to open the configuration file and add the API URL.\n\n" +
			"Example:\n" +
			"gemini:\n" +
			"  url: \"https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent\"")
	}

	return &Client{
		config:     config,
		httpClient: &http.Client{},
	}, nil
}

// GenerateContent sends a prompt to the Gemini model via REST API and returns the response.
func (c *Client) GenerateContent(prompt string) (string, error) {
	payload := GeminiRequestPayload{
		Contents:          []Content{{Parts: []Part{{Text: prompt}}}},
		SystemInstruction: &SystemInstruction{Parts: []Part{{Text: c.config.Context}}},
		GenerationConfig:  &GenerationConfig{Temperature: c.config.Temperature},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("❌ failed to marshal request payload: %w", err)
	}

	fullURL := fmt.Sprintf("%s?key=%s", c.config.URL, c.config.APIKey)
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("❌ failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("❌ failed to send request to gemini api: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("❌ failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("❌ gemini api returned non-200 status: %s\nResponse: %s", resp.Status, string(respBody))
	}

	var responsePayload GeminiResponsePayload
	if err := json.Unmarshal(respBody, &responsePayload); err != nil {
		return "", fmt.Errorf("❌ failed to unmarshal response payload: %w", err)
	}

	if responsePayload.Error != nil {
		return "", responsePayload.Error
	}

	if len(responsePayload.Candidates) == 0 || len(responsePayload.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("❌ received an empty or invalid response from the model")
	}

	return responsePayload.Candidates[0].Content.Parts[0].Text, nil
}
