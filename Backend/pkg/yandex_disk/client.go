package yandex_disk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type DiskResource struct {
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	Type       string    `json:"type"` // "dir" или "file"
	MimeType   string    `json:"mime_type"`
	Size       int64     `json:"size"`
	Modified   time.Time `json:"modified"`
	Created    time.Time `json:"created"`
	ResourceID string    `json:"resource_id"`
}

type DiskResponse struct {
	Embedded struct {
		Items []DiskResource `json:"items"`
	} `json:"_embedded"`
	Path string `json:"path"`
}

func NewClient(clientID, clientSecret, redirectURI string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) GetAuthURL() string {
	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", c.clientID)
	params.Add("redirect_uri", c.redirectURI)
	params.Add("scope", "cloud_api:disk.read cloud_api:disk.write")
	
	return "https://oauth.yandex.ru/authorize?" + params.Encode()
}

func (c *Client) ExchangeCodeForToken(ctx context.Context, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	
	resp, err := c.httpClient.PostForm("https://oauth.yandex.ru/token", data)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("oauth server returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}
	
	return &tokenResp, nil
}

func (c *Client) GetFilesList(ctx context.Context, accessToken, path string) (*DiskResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"https://cloud-api.yandex.net/v1/disk/resources",
		nil,
	)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "OAuth "+accessToken)
	
	params := req.URL.Query()
	params.Add("path", path)
	params.Add("limit", "100")
	req.URL.RawQuery = params.Encode()
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("yandex disk API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var diskResp DiskResponse
	if err := json.NewDecoder(resp.Body).Decode(&diskResp); err != nil {
		return nil, err
	}
	
	return &diskResp, nil
}