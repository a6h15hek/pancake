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

// ChatGPTRequestPayload defines the structure for the OpenAI API request body.
type ChatGPTRequestPayload struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float32       `json:"temperature"`
}

// ChatMessage represents a single message in the chat history.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatGPTResponsePayload defines the structure for the OpenAI API response body.
type ChatGPTResponsePayload struct {
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice represents a single response choice from the API.
type Choice struct {
	Message ChatMessage `json:"message"`
}

// ChatGPTClient holds the configuration for making direct API calls to OpenAI.
type ChatGPTClient struct {
	config     ChatGPTConfig
	httpClient *http.Client
}

// NewChatGPTClient creates a new ChatGPT client from the provided configuration.
func NewChatGPTClient(config ChatGPTConfig) (*ChatGPTClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("❌ ChatGPT API key is not set.\n" +
			"Please run 'pancake edit config' to open the configuration file and add your API key.\n\n" +
			"Example configuration in pancake.yml:\n" +
			"chatgpt:\n" +
			"  api_key: \"YOUR_API_KEY_HERE\"\n" +
			"  model: \"gpt-3.5-turbo\"\n" +
			"  url: \"https://api.openai.com/v1/chat/completions\"\n" +
			"  temperature: 0.7")
	}
	if config.URL == "" {
		return nil, fmt.Errorf("❌ ChatGPT API URL is not set")
	}
	if config.Model == "" {
		return nil, fmt.Errorf("❌ ChatGPT Model is not set")
	}

	return &ChatGPTClient{
		config:     config,
		httpClient: &http.Client{},
	}, nil
}

// GenerateContent sends a prompt to the ChatGPT model via REST API and returns the response.
func (c *ChatGPTClient) GenerateContent(prompt string) (string, error) {
	// The calling function `ai.go` combines the entire conversation history
	// into a single `prompt` string. We send this entire string as a single
	// user message. A more advanced implementation could parse this string
	// back into separate system, user, and assistant messages.
	messages := []ChatMessage{
		{Role: "system", Content: c.config.Context},
		{Role: "user", Content: prompt},
	}

	payload := ChatGPTRequestPayload{
		Model:       c.config.Model,
		Messages:    messages,
		Temperature: c.config.Temperature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("❌ failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.URL, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("❌ failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("❌ failed to send request to chatgpt api: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("❌ failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("❌ chatgpt api returned non-200 status: %s\nResponse: %s", resp.Status, string(respBody))
	}

	var responsePayload ChatGPTResponsePayload
	if err := json.Unmarshal(respBody, &responsePayload); err != nil {
		return "", fmt.Errorf("❌ failed to unmarshal response payload: %w", err)
	}

	if responsePayload.Error != nil {
		return "", responsePayload.Error
	}

	if len(responsePayload.Choices) == 0 {
		return "", fmt.Errorf("❌ received an empty response from the model")
	}

	return responsePayload.Choices[0].Message.Content, nil
}
