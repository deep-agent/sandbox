package sandbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/deep-agent/sandbox/model"
	"github.com/golang-jwt/jwt/v5"
)

type tokenProvider func() (string, error)

type Client struct {
	baseURL       string
	httpClient    *http.Client
	tokenProvider tokenProvider
}

type Option func(*Client)

func signToken(secret []byte) (string, error) {
	claims := jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func WithSecret(secret string) Option {
	secretBytes := []byte(secret)
	return func(c *Client) {
		c.tokenProvider = func() (string, error) {
			return signToken(secretBytes)
		}
	}
}

func WithSecretFromEnv(envKey string) Option {
	return func(c *Client) {
		c.tokenProvider = func() (string, error) {
			secret := os.Getenv(envKey)
			if secret == "" {
				return "", nil
			}
			return signToken([]byte(secret))
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type response struct {
	Code    int             `json:"code"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (c *Client) doRequest(method, path string, body interface{}) (*response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.tokenProvider != nil {
		token, err := c.tokenProvider()
		if err != nil {
			return nil, fmt.Errorf("failed to get token: %w", err)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result, nil
}

func (c *Client) GetContext() (*model.SandboxContext, error) {
	resp, err := c.doRequest("GET", "/v1/sandbox", nil)
	if err != nil {
		return nil, err
	}

	var ctx model.SandboxContext
	if err := json.Unmarshal(resp.Data, &ctx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}

	return &ctx, nil
}
