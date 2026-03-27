package newapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (c *Client) AdminCreateUser(username, password, displayName string) error {
	return c.AdminCreateUserContext(context.Background(), CreateUserRequest{
		Username:    username,
		Password:    password,
		DisplayName: displayName,
	})
}

func (c *Client) AdminCreateUserContext(ctx context.Context, input CreateUserRequest) error {
	req, err := c.newJSONRequest(ctx, http.MethodPost, "/api/user/", input)
	if err != nil {
		return err
	}
	c.setAdminAuth(req)

	_, err = c.doRequest(req)
	return err
}

func (c *Client) AdminCreateUserAndGet(username, password, displayName string) (*User, error) {
	return c.AdminCreateUserAndGetContext(context.Background(), username, password, displayName)
}

func (c *Client) AdminCreateUserAndGetContext(ctx context.Context, username, password, displayName string) (*User, error) {
	if err := c.AdminCreateUserContext(ctx, CreateUserRequest{
		Username:    username,
		Password:    password,
		DisplayName: displayName,
	}); err != nil {
		if existing, searchErr := c.AdminSearchUserContext(ctx, username); searchErr == nil && existing != nil && existing.Username == username {
			return existing, nil
		}
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		user, err := c.AdminSearchUserContext(ctx, username)
		if err == nil && user != nil && user.Username == username {
			return user, nil
		}
		lastErr = err

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(300 * time.Millisecond):
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("created user %q but failed to fetch it from search API", username)
}

func (c *Client) AdminSearchUser(keyword string) (*User, error) {
	return c.AdminSearchUserContext(context.Background(), keyword)
}

func (c *Client) AdminSearchUserContext(ctx context.Context, keyword string) (*User, error) {
	u := c.BaseURL + "/api/user/search?keyword=" + url.QueryEscape(keyword)
	req, err := c.newRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	c.setAdminAuth(req)

	var page struct {
		Items []User `json:"items"`
		Total int    `json:"total"`
	}
	if err := c.decode(req, &page); err != nil {
		return nil, err
	}
	if len(page.Items) == 0 {
		return nil, fmt.Errorf("%w: no users found for keyword %q", ErrNotFound, keyword)
	}

	trimmedKeyword := strings.TrimSpace(keyword)
	for _, item := range page.Items {
		if strings.TrimSpace(item.Username) == trimmedKeyword {
			matched := item
			return &matched, nil
		}
	}
	return &page.Items[0], nil
}

func (c *Client) AdminGetUser(userID int) (*User, error) {
	return c.AdminGetUserContext(context.Background(), userID)
}

func (c *Client) AdminGetUserContext(ctx context.Context, userID int) (*User, error) {
	u := fmt.Sprintf("%s/api/user/%d", c.BaseURL, userID)
	req, err := c.newRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	c.setAdminAuth(req)

	var user User
	if err := c.decode(req, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) AdminManageUser(userID int, action string) error {
	return c.AdminManageUserContext(context.Background(), userID, action)
}

func (c *Client) AdminManageUserContext(ctx context.Context, userID int, action string) error {
	req, err := c.newJSONRequest(ctx, http.MethodPost, "/api/user/manage", map[string]any{
		"id":     userID,
		"action": action,
	})
	if err != nil {
		return err
	}
	c.setAdminAuth(req)

	_, err = c.doRequest(req)
	return err
}

func (c *Client) AdminUpdateUser(userID int, username, displayName, group, remark string, quota int) error {
	return c.AdminUpdateUserContext(context.Background(), UpdateUserRequest{
		ID:          userID,
		Username:    username,
		DisplayName: displayName,
		Group:       group,
		Remark:      remark,
		Quota:       quota,
	})
}

func (c *Client) AdminUpdateUserContext(ctx context.Context, input UpdateUserRequest) error {
	req, err := c.newJSONRequest(ctx, http.MethodPut, "/api/user/", input)
	if err != nil {
		return err
	}
	c.setAdminAuth(req)

	_, err = c.doRequest(req)
	return err
}

func (c *Client) AdminCreateRedemptions(name string, quota int, count int, expiredTime int64) ([]string, error) {
	return c.AdminCreateRedemptionsContext(context.Background(), CreateRedemptionsRequest{
		Name:        name,
		Quota:       quota,
		Count:       count,
		ExpiredTime: expiredTime,
	})
}

func (c *Client) AdminCreateRedemptionsContext(ctx context.Context, input CreateRedemptionsRequest) ([]string, error) {
	req, err := c.newJSONRequest(ctx, http.MethodPost, "/api/redemption/", input)
	if err != nil {
		return nil, err
	}
	c.setAdminAuth(req)

	var codes []string
	if err := c.decode(req, &codes); err != nil {
		return nil, err
	}
	return codes, nil
}

func (c *Client) AdminGetLogs(params url.Values) (*LogPageInfo, error) {
	return c.AdminGetLogsContext(context.Background(), params)
}

func (c *Client) AdminGetLogsContext(ctx context.Context, params url.Values) (*LogPageInfo, error) {
	u := c.BaseURL + "/api/log/?" + params.Encode()
	req, err := c.newRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	c.setAdminAuth(req)

	var page LogPageInfo
	if err := c.decode(req, &page); err != nil {
		return nil, err
	}
	return &page, nil
}
