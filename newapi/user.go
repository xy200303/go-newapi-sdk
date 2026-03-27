package newapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"xy200303/go-newapi-sdk/newapi/core"
)

func (c *Client) UserLogin(username, password string) (string, error) {
	return c.UserLoginContext(context.Background(), username, password)
}

func (c *Client) UserLoginContext(ctx context.Context, username, password string) (string, error) {
	req, err := c.newJSONRequest(ctx, http.MethodPost, "/api/user/login", map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return "", err
	}

	resp, err := c.doRequestRaw(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session" {
			return cookie.Name + "=" + cookie.Value, nil
		}
	}

	trimmedBody := strings.TrimSpace(string(respBody))
	if trimmedBody == "" {
		return "", fmt.Errorf("login failed: session cookie not found in login response")
	}

	apiResp, parseErr := core.ParseAPIResponse(resp.StatusCode, respBody)
	if parseErr == nil && apiResp != nil {
		return "", fmt.Errorf("login failed: session cookie not found in login response")
	}
	if parseErr != nil {
		return "", parseErr
	}

	return "", fmt.Errorf("login failed: unexpected response status=%d body=%s", resp.StatusCode, core.Truncate(respBody, 200))
}

func (c *Client) UserGenerateAccessToken(sessionCookie string, userID int) (string, error) {
	return c.UserGenerateAccessTokenContext(context.Background(), sessionCookie, userID)
}

func (c *Client) UserGenerateAccessTokenContext(ctx context.Context, sessionCookie string, userID int) (string, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/api/user/token", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Cookie", sessionCookie)
	req.Header.Set("New-Api-User", fmt.Sprintf("%d", userID))

	var token string
	if err := c.decode(req, &token); err != nil {
		return "", err
	}
	return token, nil
}

func (c *Client) UserGetSelf(accessToken string, userID int) (*User, error) {
	return c.UserGetSelfContext(context.Background(), accessToken, userID)
}

func (c *Client) UserGetSelfContext(ctx context.Context, accessToken string, userID int) (*User, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/api/user/self", nil)
	if err != nil {
		return nil, err
	}
	c.setUserAuth(req, accessToken, userID)

	var user User
	if err := c.decode(req, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) UserCreateToken(accessToken string, userID int, name string, unlimitedQuota bool) (*Token, error) {
	return c.UserCreateTokenContext(context.Background(), accessToken, userID, CreateTokenRequest{
		Name:           name,
		RemainQuota:    0,
		UnlimitedQuota: unlimitedQuota,
	})
}

func (c *Client) UserCreateTokenContext(ctx context.Context, accessToken string, userID int, input CreateTokenRequest) (*Token, error) {
	req, err := c.newJSONRequest(ctx, http.MethodPost, "/api/token/", input)
	if err != nil {
		return nil, err
	}
	c.setUserAuth(req, accessToken, userID)

	apiResp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if token, ok := parseTokenResponse(apiResp.Data); ok {
		return token, nil
	}

	if fallbackToken, fallbackErr := c.findLatestTokenByNameContext(ctx, accessToken, userID, input.Name, input.UnlimitedQuota); fallbackErr == nil && fallbackToken != nil {
		return fallbackToken, nil
	}

	return &Token{
		Name:           strings.TrimSpace(input.Name),
		UnlimitedQuota: input.UnlimitedQuota,
	}, nil
}

func (c *Client) UserListTokens(accessToken string, userID int, page, pageSize int) ([]Token, error) {
	pageInfo, err := c.UserListTokensPageContext(context.Background(), accessToken, userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return pageInfo.Items, nil
}

func (c *Client) UserListTokensPage(accessToken string, userID int, page, pageSize int) (*TokenPageInfo, error) {
	return c.UserListTokensPageContext(context.Background(), accessToken, userID, page, pageSize)
}

func (c *Client) UserListTokensPageContext(ctx context.Context, accessToken string, userID int, page, pageSize int) (*TokenPageInfo, error) {
	u := fmt.Sprintf("%s/api/token/?p=%d&page_size=%d", c.BaseURL, page, pageSize)
	req, err := c.newRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	c.setUserAuth(req, accessToken, userID)

	var pageInfo TokenPageInfo
	if err := c.decode(req, &pageInfo); err != nil {
		return nil, err
	}
	for i := range pageInfo.Items {
		pageInfo.Items[i] = normalizeToken(pageInfo.Items[i])
	}
	return &pageInfo, nil
}

func (c *Client) UserDeleteToken(accessToken string, userID int, tokenID int) error {
	return c.UserDeleteTokenContext(context.Background(), accessToken, userID, tokenID)
}

func (c *Client) UserDeleteTokenContext(ctx context.Context, accessToken string, userID int, tokenID int) error {
	u := c.BaseURL + "/api/token/" + fmt.Sprintf("%d", tokenID)
	req, err := c.newRequest(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	c.setUserAuth(req, accessToken, userID)

	_, err = c.doRequest(req)
	return err
}

func (c *Client) UserListModels(accessToken string, userID int) ([]Model, error) {
	return c.UserListModelsContext(context.Background(), accessToken, userID)
}

func (c *Client) UserListModelsContext(ctx context.Context, accessToken string, userID int) ([]Model, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/api/models", nil)
	if err != nil {
		return nil, err
	}
	c.setUserAuth(req, accessToken, userID)

	apiResp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var directModels []Model
	if err := json.Unmarshal(apiResp.Data, &directModels); err == nil {
		return directModels, nil
	}

	var wrapped struct {
		Data  []Model  `json:"data"`
		Items []Model  `json:"items"`
		Names []string `json:"models"`
	}
	if err := json.Unmarshal(apiResp.Data, &wrapped); err == nil {
		if len(wrapped.Data) > 0 {
			return wrapped.Data, nil
		}
		if len(wrapped.Items) > 0 {
			return wrapped.Items, nil
		}
		if len(wrapped.Names) > 0 {
			models := make([]Model, 0, len(wrapped.Names))
			for _, name := range wrapped.Names {
				models = append(models, Model{ID: name, Name: name})
			}
			return models, nil
		}
	}

	var directNames []string
	if err := json.Unmarshal(apiResp.Data, &directNames); err == nil {
		models := make([]Model, 0, len(directNames))
		for _, name := range directNames {
			models = append(models, Model{ID: name, Name: name})
		}
		return models, nil
	}

	var groupedNames map[string]json.RawMessage
	if err := json.Unmarshal(apiResp.Data, &groupedNames); err == nil && len(groupedNames) > 0 {
		models := make([]Model, 0)
		seen := make(map[string]struct{})
		for group, rawNames := range groupedNames {
			if string(rawNames) == "null" {
				continue
			}

			var names []string
			if err := json.Unmarshal(rawNames, &names); err != nil {
				continue
			}

			for _, name := range names {
				trimmedName := strings.TrimSpace(name)
				if trimmedName == "" {
					continue
				}
				if _, ok := seen[trimmedName]; ok {
					continue
				}
				seen[trimmedName] = struct{}{}
				models = append(models, Model{
					ID:    trimmedName,
					Name:  trimmedName,
					Group: group,
				})
			}
		}
		if len(models) > 0 {
			return models, nil
		}
	}

	return nil, fmt.Errorf("parsing models list: unsupported response format")
}

func (c *Client) UserRedeemRedemption(accessToken string, userID int, key string) (int, error) {
	return c.UserRedeemRedemptionContext(context.Background(), accessToken, userID, key)
}

func (c *Client) UserRedeemRedemptionContext(ctx context.Context, accessToken string, userID int, key string) (int, error) {
	req, err := c.newJSONRequest(ctx, http.MethodPost, "/api/user/topup", map[string]string{
		"key": key,
	})
	if err != nil {
		return 0, err
	}
	c.setUserAuth(req, accessToken, userID)

	var quota int
	if err := c.decode(req, &quota); err != nil {
		return 0, err
	}
	return quota, nil
}

func (c *Client) GetUserWithToken(rawToken string, userID int) (*User, error) {
	return c.GetUserWithTokenContext(context.Background(), rawToken, userID)
}

func (c *Client) GetUserWithTokenContext(ctx context.Context, rawToken string, userID int) (*User, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/api/user/self", nil)
	if err != nil {
		return nil, err
	}
	c.setUserAuth(req, rawToken, userID)

	var user User
	if err := c.decode(req, &user); err != nil {
		return nil, err
	}

	if user.ID > 0 && userID > 0 && user.ID != userID {
		return nil, fmt.Errorf("user token mismatch: expected user %d, got %d", userID, user.ID)
	}
	return &user, nil
}

func (c *Client) ListTokensWithToken(rawToken string, userID, page, pageSize int) ([]Token, error) {
	return c.UserListTokens(rawToken, userID, page, pageSize)
}

func (c *Client) ListTokensWithTokenContext(ctx context.Context, rawToken string, userID, page, pageSize int) ([]Token, error) {
	pageInfo, err := c.UserListTokensPageContext(ctx, rawToken, userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return pageInfo.Items, nil
}

func (c *Client) CreateTokenWithToken(rawToken string, userID int, name string, unlimitedQuota bool) error {
	_, err := c.UserCreateToken(rawToken, userID, name, unlimitedQuota)
	return err
}

func (c *Client) CreateTokenWithTokenContext(ctx context.Context, rawToken string, userID int, name string, unlimitedQuota bool) (*Token, error) {
	return c.UserCreateTokenContext(ctx, rawToken, userID, CreateTokenRequest{
		Name:           name,
		RemainQuota:    0,
		UnlimitedQuota: unlimitedQuota,
	})
}

func parseTokenResponse(raw json.RawMessage) (*Token, bool) {
	var token Token
	if err := json.Unmarshal(raw, &token); err == nil {
		token = normalizeToken(token)
		if token.ID > 0 || token.Key != "" || token.Name != "" {
			return &token, true
		}
	}

	var wrapped struct {
		Token *Token `json:"token"`
		Item  *Token `json:"item"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil {
		if wrapped.Token != nil {
			token = normalizeToken(*wrapped.Token)
			return &token, true
		}
		if wrapped.Item != nil {
			token = normalizeToken(*wrapped.Item)
			return &token, true
		}
	}

	return nil, false
}

func normalizeToken(token Token) Token {
	token.Name = strings.TrimSpace(token.Name)
	token.Key = normalizeTokenKey(token.Key)
	return token
}

func normalizeTokenKey(key string) string {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "sk-") {
		return trimmed
	}
	return "sk-" + trimmed
}

func (c *Client) findLatestTokenByName(accessToken string, userID int, name string, unlimitedQuota bool) (*Token, error) {
	return c.findLatestTokenByNameContext(context.Background(), accessToken, userID, name, unlimitedQuota)
}

func (c *Client) findLatestTokenByNameContext(ctx context.Context, accessToken string, userID int, name string, unlimitedQuota bool) (*Token, error) {
	pageInfo, err := c.UserListTokensPageContext(ctx, accessToken, userID, 1, 100)
	if err != nil {
		return nil, err
	}

	targetName := strings.TrimSpace(name)
	var best *Token
	bestScore := -1
	for i := range pageInfo.Items {
		item := normalizeToken(pageInfo.Items[i])
		if item.Name != targetName {
			continue
		}

		score := 0
		if item.UnlimitedQuota == unlimitedQuota {
			score += 2
		}
		if item.Key != "" {
			score++
		}

		if best == nil || score > bestScore || (score == bestScore && item.ID > best.ID) {
			candidate := item
			best = &candidate
			bestScore = score
		}
	}

	return best, nil
}
