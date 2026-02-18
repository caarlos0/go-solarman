package solarman

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const baseURL = "https://globalapi.solarmanpv.com"

type Client struct {
	appID string
	token string
}

func New(appID, appSecret, email, password string) (*Client, error) {
	token, err := newOauthToken(
		appID, appSecret, email,
		fmt.Sprintf("%x", sha256.Sum256([]byte(password))),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		appID: appID,
		token: token,
	}, nil
}

func newOauthToken(appID, appSecret, email, password string) (string, error) {
	data := fmt.Sprintf(`{"appSecret":%q,"email":%q,"password":%q}`, appSecret, email, password)
	url := fmt.Sprintf(baseURL+"/account/v1.0/token?appId=%s&language=en&=", appID)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("could not auth: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not auth: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not auth: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("could not auth: %w: %s", err, string(bts))
	}

	var aresp authResponse
	if err := json.Unmarshal(bts, &aresp); err != nil {
		return "", fmt.Errorf("could not auth: %w", err)
	}

	if !aresp.Success {
		return "", fmt.Errorf("could not auth: solarman error: %s", aresp.Msg)
	}

	var tresp tokenResponse
	if err := json.Unmarshal(bts, &tresp); err != nil {
		return "", fmt.Errorf("could not auth: %w", err)
	}

	return tresp.AccessToken, nil
}

type authResponse struct {
	Msg     string `json:"msg"`
	Success bool   `json:"success"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
}
