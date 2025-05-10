package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// NakamaAuthResponse represents the structure returned by Nakama on authentication
type NakamaAuthResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Created      bool   `json:"created"`
}

// AuthenticateWithEmail authenticates a user with email/password
func AuthenticateWithEmail(server, email, password string, create bool) (*NakamaAuthResponse, error) {
	url := fmt.Sprintf("http://%s/v2/account/authenticate/email?create=%t", server, create)

	payload := map[string]interface{}{
		"email":    email,
		"password": password,
	}
	return sendAuthRequest(url, payload)
}

// AuthenticateWithDeviceID authenticates a user with a device ID
func AuthenticateWithDeviceID(server, deviceID string, create bool) (*NakamaAuthResponse, error) {
	url := fmt.Sprintf("http://%s/v2/account/authenticate/device?create=%t", server, create)

	payload := map[string]interface{}{
		"device_id": deviceID,
	}
	return sendAuthRequest(url, payload)
}

// AuthenticateWithServerKey tries to authenticate using device ID with server key auth
func AuthenticateWithServerKey(server, serverKey string) (*NakamaAuthResponse, error) {
	url := fmt.Sprintf("http://%s/v2/account/authenticate/device?create=true", server)

	payload := map[string]interface{}{
		"device_id": "lumaris-cli-serverkey",
	}
	return sendAuthRequestWithBasicAuth(url, payload, serverKey)
}

// sendAuthRequest sends an authentication request with default server key
func sendAuthRequest(url string, payload map[string]interface{}) (*NakamaAuthResponse, error) {
	return sendAuthRequestWithBasicAuth(url, payload, "defaultkey")
}

// sendAuthRequestWithBasicAuth sends a request with HTTP Basic Auth header
func sendAuthRequestWithBasicAuth(url string, payload map[string]interface{}, serverKey string) (*NakamaAuthResponse, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	auth := base64.StdEncoding.EncodeToString([]byte(serverKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth failed [%d]: %s", resp.StatusCode, string(body))
	}

	var authResp NakamaAuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("failed to parse auth response: %w", err)
	}

	return &authResp, nil
}
