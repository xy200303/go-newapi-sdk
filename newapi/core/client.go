package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout   = 15 * time.Second
	defaultUserAgent = "go-newapi-sdk/1.0"
)

type Option func(*Client) error

type Client struct {
	BaseURL     string
	RootToken   string
	RootUserID  int
	HTTPClient  *http.Client
	UserAgent   string
	DefaultAuth Auth
}

type APIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func New(baseURL string, opts ...Option) (*Client, error) {
	normalizedURL, err := normalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	client := &Client{
		BaseURL:    normalizedURL,
		HTTPClient: &http.Client{Timeout: defaultTimeout},
		UserAgent:  defaultUserAgent,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	if client.HTTPClient == nil {
		client.HTTPClient = &http.Client{Timeout: defaultTimeout}
	}
	if strings.TrimSpace(client.UserAgent) == "" {
		client.UserAgent = defaultUserAgent
	}

	return client, nil
}

func NewClient(baseURL, rootToken string, rootUserID, timeoutSec int) *Client {
	timeout := defaultTimeout
	if timeoutSec > 0 {
		timeout = time.Duration(timeoutSec) * time.Second
	}

	client, err := New(
		baseURL,
		WithAdminAuth(rootToken, rootUserID),
		WithTimeout(timeout),
	)
	if err == nil {
		return client
	}

	return &Client{
		BaseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		RootToken:  rootToken,
		RootUserID: rootUserID,
		HTTPClient: &http.Client{Timeout: timeout},
		UserAgent:  defaultUserAgent,
		DefaultAuth: Auth{
			BearerToken: rootToken,
			UserID:      rootUserID,
		},
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		if httpClient == nil {
			return fmt.Errorf("http client cannot be nil")
		}
		c.HTTPClient = httpClient
		return nil
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) error {
		if timeout <= 0 {
			return fmt.Errorf("timeout must be greater than zero")
		}
		if c.HTTPClient == nil {
			c.HTTPClient = &http.Client{}
		}
		c.HTTPClient.Timeout = timeout
		return nil
	}
}

func WithAdminAuth(rootToken string, rootUserID int) Option {
	return func(c *Client) error {
		c.RootToken = strings.TrimSpace(rootToken)
		c.RootUserID = rootUserID
		c.DefaultAuth = Auth{
			BearerToken: c.RootToken,
			UserID:      c.RootUserID,
		}
		return nil
	}
}

func WithDefaultAuth(auth Auth) Option {
	return func(c *Client) error {
		c.DefaultAuth = auth
		return nil
	}
}

func WithBearerToken(token string) Option {
	return func(c *Client) error {
		c.DefaultAuth.BearerToken = strings.TrimSpace(token)
		return nil
	}
}

func WithSessionCookie(cookie string) Option {
	return func(c *Client) error {
		c.DefaultAuth.SessionCookie = strings.TrimSpace(cookie)
		return nil
	}
}

func WithDefaultUserID(userID int) Option {
	return func(c *Client) error {
		c.DefaultAuth.UserID = userID
		return nil
	}
}

func WithUserAgent(userAgent string) Option {
	return func(c *Client) error {
		c.UserAgent = strings.TrimSpace(userAgent)
		return nil
	}
}

func (c *Client) DoRequest(req *http.Request) (*APIResponse, error) {
	resp, err := c.DoRequestRaw(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	apiResp, err := ParseAPIResponse(resp.StatusCode, body)
	if err != nil {
		return apiResp, err
	}

	return apiResp, nil
}

func (c *Client) DoRequestRaw(req *http.Request) (*http.Response, error) {
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: defaultTimeout}
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	return resp, nil
}

func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	requestURL, err := c.resolveURL(path)
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, err
	}
	c.setDefaultHeaders(req)
	return req, nil
}

func (c *Client) NewJSONRequest(ctx context.Context, method, path string, payload any) (*http.Request, error) {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}
		body = bytes.NewReader(raw)
	}

	req, err := c.NewRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *Client) Decode(req *http.Request, out any) error {
	apiResp, err := c.DoRequest(req)
	if err != nil {
		return err
	}
	return decodeResponseData(apiResp.Data, out)
}

func (c *Client) ApplyAdminAuth(req *http.Request) {
	if strings.TrimSpace(c.RootToken) != "" {
		req.Header.Set("Authorization", "Bearer "+c.RootToken)
	}
	if c.RootUserID > 0 {
		req.Header.Set("New-Api-User", strconv.Itoa(c.RootUserID))
	}
}

func (c *Client) ApplyUserAuth(req *http.Request, accessToken string, userID int) {
	if strings.TrimSpace(accessToken) != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	if userID > 0 {
		req.Header.Set("New-Api-User", strconv.Itoa(userID))
	}
}

func (c *Client) resolveURL(path string) (string, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path, nil
	}

	baseURL, err := normalizeBaseURL(c.BaseURL)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return baseURL + path, nil
}

func (c *Client) setDefaultHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(c.UserAgent) != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
}

func (c *Client) resolveAuth(explicit *Auth) Auth {
	if explicit != nil {
		return *explicit
	}
	if strings.TrimSpace(c.DefaultAuth.BearerToken) != "" || c.DefaultAuth.UserID > 0 || strings.TrimSpace(c.DefaultAuth.SessionCookie) != "" {
		return c.DefaultAuth
	}
	return Auth{
		BearerToken: c.RootToken,
		UserID:      c.RootUserID,
	}
}

func normalizeBaseURL(baseURL string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmed == "" {
		return "", fmt.Errorf("base URL cannot be empty")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("base URL must include scheme and host")
	}

	return trimmed, nil
}

func ParseAPIResponse(statusCode int, body []byte) (*APIResponse, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		if statusCode >= 200 && statusCode < 300 {
			return &APIResponse{Success: true}, nil
		}
		return nil, &APIError{
			StatusCode: statusCode,
			Message:    http.StatusText(statusCode),
		}
	}

	var apiResp APIResponse
	if err := json.Unmarshal(trimmed, &apiResp); err != nil {
		if statusCode >= 200 && statusCode < 300 {
			return nil, fmt.Errorf("parsing response JSON: %w (body: %s)", err, Truncate(trimmed, 200))
		}

		return nil, &APIError{
			StatusCode: statusCode,
			Message:    http.StatusText(statusCode),
			Body:       Truncate(trimmed, 200),
		}
	}

	if statusCode < 200 || statusCode >= 300 || !apiResp.Success {
		message := strings.TrimSpace(apiResp.Message)
		if message == "" {
			message = http.StatusText(statusCode)
		}
		if message == "" {
			message = "request failed"
		}

		return &apiResp, &APIError{
			StatusCode: statusCode,
			Message:    message,
			Body:       Truncate(trimmed, 200),
		}
	}

	return &apiResp, nil
}

func decodeResponseData(data json.RawMessage, out any) error {
	if out == nil {
		return nil
	}

	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}

	if err := json.Unmarshal(trimmed, out); err != nil {
		return fmt.Errorf("parsing response data: %w", err)
	}
	return nil
}

func Truncate(b []byte, maxLen int) string {
	if len(b) <= maxLen {
		return string(b)
	}
	return string(b[:maxLen]) + "..."
}
