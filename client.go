package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

const baseURL = "https://globalapi.solarmanpv.com/device/v1.0"

type Client struct {
	c *http.Client

	appID string
}

func New(appID, appSecret, username, password string) (*Client, error) {
	auth, err := newAccessToken(appID, appSecret, username, password)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(auth, &token); err != nil {
		return nil, fmt.Errorf("could not auth: %w", err)
	}

	oauthConfg := oauth2.Config{
		ClientID:     appID,
		ClientSecret: appSecret,
	}

	c := oauthConfg.Client(context.Background(), &token)
	return &Client{
		c:     c,
		appID: appID,
	}, nil
}

func newAccessToken(appID, appSecret, username, password string) ([]byte, error) {
	data := fmt.Sprintf(
		`{"appSecret":%q,"email":%q,"password":%q}`,
		appSecret,
		username,
		password,
	)

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf(
			baseURL+"/token?appId=%s&language=en&=",
			appID,
		),
		strings.NewReader(data),
	)
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
		return nil, fmt.Errorf("could not auth: %w", err)
	}

	return bts, nil
}
