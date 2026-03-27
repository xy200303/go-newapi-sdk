package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var pathParamPattern = regexp.MustCompile(`\{([^\}]+)\}`)

type Auth struct {
	BearerToken   string
	UserID        int
	SessionCookie string
}

type CallConfig struct {
	PathParams  map[string]any
	Query       any
	JSONBody    any
	Body        io.Reader
	ContentType string
	Headers     http.Header
	Auth        *Auth
}

type RawResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

type Operation struct {
	client  *Client
	Name    string
	Title   string
	Method  string
	Path    string
	DocPath string
}

func NewOperation(client *Client, name, title, method, path, docPath string) *Operation {
	return &Operation{
		client:  client,
		Name:    name,
		Title:   title,
		Method:  method,
		Path:    path,
		DocPath: docPath,
	}
}

func (o *Operation) DocURL() string {
	if strings.HasPrefix(o.DocPath, "http://") || strings.HasPrefix(o.DocPath, "https://") {
		return o.DocPath
	}
	return "https://docs.newapi.pro" + o.DocPath
}

func (o *Operation) Do(ctx context.Context, cfg *CallConfig, out any) error {
	resp, err := o.DoRaw(ctx, cfg)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return newHTTPError(resp.StatusCode, body)
	}

	if out == nil {
		return nil
	}

	switch target := out.(type) {
	case *RawResponse:
		target.StatusCode = resp.StatusCode
		target.Header = resp.Header.Clone()
		target.Body = body
		return nil
	case *[]byte:
		*target = append((*target)[:0], body...)
		return nil
	case *string:
		*target = string(body)
		return nil
	case io.Writer:
		_, err := target.Write(body)
		return err
	}

	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return nil
	}

	if err := json.Unmarshal(trimmed, out); err == nil {
		return nil
	}

	if apiResp, err := ParseAPIResponse(resp.StatusCode, trimmed); err == nil && apiResp != nil && len(bytes.TrimSpace(apiResp.Data)) > 0 {
		if err := decodeResponseData(apiResp.Data, out); err == nil {
			return nil
		}
	}

	return fmt.Errorf("parsing response body: unsupported response format")
}

func (o *Operation) DoRaw(ctx context.Context, cfg *CallConfig) (*http.Response, error) {
	if cfg == nil {
		cfg = &CallConfig{}
	}

	body, contentType, err := prepareRequestBody(cfg)
	if err != nil {
		return nil, err
	}

	path, err := resolveOperationPath(o.Path, cfg.PathParams)
	if err != nil {
		return nil, err
	}

	requestURL, err := o.client.resolveURL(path)
	if err != nil {
		return nil, err
	}

	queryValues, err := encodeQueryValues(cfg.Query)
	if err != nil {
		return nil, err
	}
	if len(queryValues) > 0 {
		parsedURL, parseErr := url.Parse(requestURL)
		if parseErr != nil {
			return nil, parseErr
		}
		encoded := parsedURL.Query()
		for key, values := range queryValues {
			for _, value := range values {
				encoded.Add(key, value)
			}
		}
		parsedURL.RawQuery = encoded.Encode()
		requestURL = parsedURL.String()
	}

	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, o.Method, requestURL, body)
	if err != nil {
		return nil, err
	}

	o.client.setDefaultHeaders(req)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for key, values := range cfg.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	auth := o.client.resolveAuth(cfg.Auth)
	if strings.TrimSpace(auth.BearerToken) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(auth.BearerToken))
	}
	if auth.UserID > 0 {
		req.Header.Set("New-Api-User", strconv.Itoa(auth.UserID))
	}
	if strings.TrimSpace(auth.SessionCookie) != "" {
		req.Header.Set("Cookie", strings.TrimSpace(auth.SessionCookie))
	}

	return o.client.DoRequestRaw(req)
}

func prepareRequestBody(cfg *CallConfig) (io.Reader, string, error) {
	if cfg.JSONBody != nil && cfg.Body != nil {
		return nil, "", fmt.Errorf("json body and raw body cannot be set at the same time")
	}

	if cfg.JSONBody != nil {
		raw, err := json.Marshal(cfg.JSONBody)
		if err != nil {
			return nil, "", fmt.Errorf("marshaling request body: %w", err)
		}
		contentType := cfg.ContentType
		if strings.TrimSpace(contentType) == "" {
			contentType = "application/json"
		}
		return bytes.NewReader(raw), contentType, nil
	}

	return cfg.Body, strings.TrimSpace(cfg.ContentType), nil
}

func resolveOperationPath(template string, pathParams map[string]any) (string, error) {
	if len(pathParams) == 0 {
		if pathParamPattern.MatchString(template) {
			return "", fmt.Errorf("missing path params for route %q", template)
		}
		return template, nil
	}

	missing := make([]string, 0)
	path := pathParamPattern.ReplaceAllStringFunc(template, func(match string) string {
		key := strings.TrimSuffix(strings.TrimPrefix(match, "{"), "}")
		value, ok := pathParams[key]
		if !ok {
			missing = append(missing, key)
			return match
		}
		return url.PathEscape(fmt.Sprint(value))
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("missing path params: %s", strings.Join(missing, ", "))
	}
	return path, nil
}

func encodeQueryValues(input any) (url.Values, error) {
	values := url.Values{}
	if input == nil {
		return values, nil
	}

	switch typed := input.(type) {
	case url.Values:
		for key, items := range typed {
			for _, item := range items {
				values.Add(key, item)
			}
		}
		return values, nil
	case map[string]string:
		for key, value := range typed {
			values.Add(key, value)
		}
		return values, nil
	case map[string][]string:
		for key, items := range typed {
			for _, item := range items {
				values.Add(key, item)
			}
		}
		return values, nil
	}

	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("encoding query params: %w", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, fmt.Errorf("decoding query params: %w", err)
	}

	for key, value := range decoded {
		addQueryValue(values, key, value)
	}
	return values, nil
}

func addQueryValue(values url.Values, key string, value any) {
	if value == nil {
		return
	}

	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			addQueryValue(values, key, item)
		}
	case map[string]any:
		raw, err := json.Marshal(typed)
		if err == nil {
			values.Add(key, string(raw))
		}
	case bool:
		values.Add(key, strconv.FormatBool(typed))
	case float64:
		if typed == float64(int64(typed)) {
			values.Add(key, strconv.FormatInt(int64(typed), 10))
			return
		}
		values.Add(key, strconv.FormatFloat(typed, 'f', -1, 64))
	case string:
		values.Add(key, typed)
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			for i := 0; i < rv.Len(); i++ {
				addQueryValue(values, key, rv.Index(i).Interface())
			}
			return
		}
		values.Add(key, fmt.Sprint(value))
	}
}

func newHTTPError(statusCode int, body []byte) error {
	trimmed := bytes.TrimSpace(body)
	message := http.StatusText(statusCode)
	if len(trimmed) > 0 {
		var envelope struct {
			Message string `json:"message"`
			Error   struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(trimmed, &envelope); err == nil {
			switch {
			case strings.TrimSpace(envelope.Error.Message) != "":
				message = envelope.Error.Message
			case strings.TrimSpace(envelope.Message) != "":
				message = envelope.Message
			}
		}
	}

	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Body:       Truncate(trimmed, 200),
	}
}
