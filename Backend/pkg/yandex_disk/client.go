package yandex_disk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	MediaType  string    `json:"media_type"` // Дополнительное поле
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
	params.Add("scope", "cloud_api:disk.read cloud_api:disk.write cloud_api:disk.app_folder")
	
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
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "OAuth "+accessToken)
	
	params := req.URL.Query()
	
	// Яндекс.Диск ожидает "disk:/" для корня
	yandexPath := "disk:/"
	if path != "" && path != "/" {
		// Очищаем путь
		cleanPath := strings.TrimPrefix(path, "/")
		cleanPath = strings.TrimPrefix(cleanPath, "disk:/")
		yandexPath = "disk:/" + cleanPath
	}
	
	params.Add("path", yandexPath)
	params.Add("limit", "1000") // Увеличим лимит
	params.Add("sort", "name")  // Сортировка по имени
	params.Add("preview_size", "S") // Для картинок
	params.Add("preview_crop", "false")
	
	req.URL.RawQuery = params.Encode()
	
	fmt.Printf("DEBUG: Requesting Yandex.Disk path: %s\n", yandexPath)
	fmt.Printf("DEBUG: Full URL: %s\n", req.URL.String())
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	// Читаем тело ответа
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("DEBUG: Yandex.Disk API error (%d): %s\n", resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("yandex disk API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	// Выводим сырой ответ для отладки
	fmt.Printf("DEBUG: Raw Yandex.Disk response (first 500 chars): %.500s\n", string(bodyBytes))
	
	var diskResp DiskResponse
	if err := json.Unmarshal(bodyBytes, &diskResp); err != nil {
		fmt.Printf("DEBUG: JSON decode error: %v\n", err)
		fmt.Printf("DEBUG: Raw JSON: %s\n", string(bodyBytes))
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}
	
	fmt.Printf("DEBUG: Successfully parsed %d items from Yandex.Disk\n", len(diskResp.Embedded.Items))
	for i, item := range diskResp.Embedded.Items {
		fmt.Printf("DEBUG: Item %d: %s (type: %s, size: %d)\n", i+1, item.Name, item.Type, item.Size)
	}
	
	return &diskResp, nil
}
// UploadFile - загружает файл в Яндекс.Диск
func (c *Client) UploadFile(ctx context.Context, accessToken, path string, content io.Reader) error {
	// 1. Получаем URL для загрузки
	uploadURL, err := c.getUploadURL(ctx, accessToken, path)
	if err != nil {
		return fmt.Errorf("failed to get upload URL: %w", err)
	}
	
	// 2. Загружаем файл
	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, content)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/octet-stream")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()
	
	// Проверяем статус ответа
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// DownloadFile - скачивает файл из Яндекс.Диска
func (c *Client) DownloadFile(ctx context.Context, accessToken, path string) (io.ReadCloser, error) {
	// 1. Получаем URL для скачивания
	downloadURL, err := c.getDownloadURL(ctx, accessToken, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get download URL: %w", err)
	}
	
	// 2. Скачиваем файл
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}
	
	req.Header.Set("Authorization", "OAuth "+accessToken)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	
	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	return resp.Body, nil
}

// DeleteFile - удаляет файл из Яндекс.Диска
func (c *Client) DeleteFile(ctx context.Context, accessToken, path string) error {
	req, err := http.NewRequestWithContext(
		ctx,
		"DELETE",
		"https://cloud-api.yandex.net/v1/disk/resources",
		nil,
	)
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "OAuth "+accessToken)
	
	params := req.URL.Query()
	params.Add("path", path)
	params.Add("permanently", "true")
	req.URL.RawQuery = params.Encode()
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// getUploadURL - получает URL для загрузки файла
func (c *Client) getUploadURL(ctx context.Context, accessToken, path string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"https://cloud-api.yandex.net/v1/disk/resources/upload",
		nil,
	)
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Authorization", "OAuth "+accessToken)
	
	params := req.URL.Query()
	params.Add("path", path)
	params.Add("overwrite", "true")
	req.URL.RawQuery = params.Encode()
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get upload URL: status %d: %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Href string `json:"href"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	return result.Href, nil
}

// getDownloadURL - получает URL для скачивания файла
func (c *Client) getDownloadURL(ctx context.Context, accessToken, path string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"https://cloud-api.yandex.net/v1/disk/resources/download",
		nil,
	)
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Authorization", "OAuth "+accessToken)
	
	params := req.URL.Query()
	params.Add("path", path)
	req.URL.RawQuery = params.Encode()
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get download URL: status %d: %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Href string `json:"href"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	return result.Href, nil
}