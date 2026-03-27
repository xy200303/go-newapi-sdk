package newapi

import (
	"context"
	"io"
	"net/http"

	"xy200303/go-newapi-sdk/newapi/aimodel"
	"xy200303/go-newapi-sdk/newapi/core"
	"xy200303/go-newapi-sdk/newapi/management"
)

type Option = core.Option

var (
	WithHTTPClient    = core.WithHTTPClient
	WithTimeout       = core.WithTimeout
	WithAdminAuth     = core.WithAdminAuth
	WithDefaultAuth   = core.WithDefaultAuth
	WithBearerToken   = core.WithBearerToken
	WithSessionCookie = core.WithSessionCookie
	WithDefaultUserID = core.WithDefaultUserID
	WithUserAgent     = core.WithUserAgent
)

type Client struct {
	*core.Client

	AIModel    *aimodel.Service
	Management *management.Service
}

func NewApiClient(baseURL string, opts ...Option) (*Client, error) {
	return New(baseURL, opts...)
}

func New(baseURL string, opts ...Option) (*Client, error) {
	coreClient, err := core.New(baseURL, opts...)
	if err != nil {
		return nil, err
	}

	client := &Client{}
	client.Client = coreClient
	client.initServices()
	return client, nil
}

func NewClient(baseURL, rootToken string, rootUserID, timeoutSec int) *Client {
	client := &Client{}
	client.Client = core.NewClient(baseURL, rootToken, rootUserID, timeoutSec)
	client.initServices()
	return client
}

func (c *Client) initServices() {
	c.AIModel = aimodel.NewService(c.Client)
	c.Management = management.NewService(c.Client)
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	return c.Client.NewRequest(ctx, method, path, body)
}

func (c *Client) newJSONRequest(ctx context.Context, method, path string, payload any) (*http.Request, error) {
	return c.Client.NewJSONRequest(ctx, method, path, payload)
}

func (c *Client) doRequest(req *http.Request) (*core.APIResponse, error) {
	return c.Client.DoRequest(req)
}

func (c *Client) doRequestRaw(req *http.Request) (*http.Response, error) {
	return c.Client.DoRequestRaw(req)
}

func (c *Client) decode(req *http.Request, out any) error {
	return c.Client.Decode(req, out)
}

func (c *Client) setAdminAuth(req *http.Request) {
	c.Client.ApplyAdminAuth(req)
}

func (c *Client) setUserAuth(req *http.Request, accessToken string, userID int) {
	c.Client.ApplyUserAuth(req, accessToken, userID)
}
