package solarman

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const baseURL = "https://globalapi.solarmanpv.com"

type Client struct {
	c *http.Client

	appID string
}

func New(appID, appSecret, email, password string) (*Client, error) {
	t, err := newOauthToken(
		appID, appSecret, email,
		fmt.Sprintf("%x", sha256.Sum256([]byte(password))),
	)
	if err != nil {
		return nil, err
	}

	oauthConfg := oauth2.Config{
		ClientID:     appID,
		ClientSecret: appSecret,
	}

	c := oauthConfg.Client(context.Background(), t)
	return &Client{
		c:     c,
		appID: appID,
	}, nil
}

func newOauthToken(appID, appSecret, email, password string) (*oauth2.Token, error) {
	data := fmt.Sprintf(`{"appSecret":%q,"email":%q,"password":%q}`, appSecret, email, password)
	url := fmt.Sprintf(baseURL+"/account/v1.0/token?appId=%s&language=en&=", appID)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("could not auth: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not auth: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not auth: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not auth: %w: %s", err, string(bts))
	}

	var aresp authResponse
	if err := json.Unmarshal(bts, &aresp); err != nil {
		return nil, fmt.Errorf("could not auth: %w", err)
	}

	if !aresp.Success {
		return nil, fmt.Errorf("could not auth: solarman error: %s", aresp.Msg)
	}

	var tresp tokenResponse
	if err := json.Unmarshal(bts, &tresp); err != nil {
		return nil, fmt.Errorf("could not auth: %w", err)
	}

	expiresIn, _ := strconv.ParseInt(tresp.ExpiresIn, 10, 64)
	token := &oauth2.Token{
		AccessToken:  tresp.AccessToken,
		TokenType:    tresp.TokenType,
		RefreshToken: tresp.RefreshToken,
	}
	if expiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}
	return token, nil
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
