package newapi

type User struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AccessToken string `json:"access_token"`
	Quota       int    `json:"quota"`
	UsedQuota   int    `json:"used_quota"`
	Group       string `json:"group"`
	Status      int    `json:"status"`
	Remark      string `json:"remark,omitempty"`
}

type Token struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Key            string `json:"key"`
	UnlimitedQuota bool   `json:"unlimited_quota"`
}

type TokenPageInfo struct {
	Items    []Token `json:"items"`
	Total    int64   `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
}

type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object,omitempty"`
	OwnedBy string `json:"owned_by,omitempty"`
	Name    string `json:"name,omitempty"`
	Group   string `json:"group,omitempty"`
}

type Log struct {
	ID               int    `json:"id"`
	CreatedAt        int64  `json:"created_at"`
	Type             int    `json:"type"`
	Content          string `json:"content"`
	Username         string `json:"username"`
	TokenName        string `json:"token_name"`
	ModelName        string `json:"model_name"`
	Quota            int    `json:"quota"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	UseTime          int    `json:"use_time"`
	IsStream         bool   `json:"is_stream"`
	ChannelID        int    `json:"channel"`
	TokenID          int    `json:"token_id"`
	Group            string `json:"group"`
	RequestID        string `json:"request_id,omitempty"`
}

type LogPageInfo struct {
	Items    []Log `json:"items"`
	Total    int   `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type Session struct {
	Cookie string
}

type CreateUserRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type UpdateUserRequest struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Group       string `json:"group"`
	Remark      string `json:"remark"`
	Quota       int    `json:"quota"`
}

type CreateTokenRequest struct {
	Name           string `json:"name"`
	RemainQuota    int    `json:"remain_quota"`
	UnlimitedQuota bool   `json:"unlimited_quota"`
}

type CreateRedemptionsRequest struct {
	Name        string `json:"name"`
	Quota       int    `json:"quota"`
	Count       int    `json:"count"`
	ExpiredTime int64  `json:"expired_time"`
}
