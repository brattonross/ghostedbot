package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func String(v string) *string {
	return &v
}

func Int(v int) *int {
	return &v
}

func Bool(v bool) *bool {
	return &v
}

type ApplicationCommand struct {
	Id            string  `json:"id"`
	Type          int     `json:"type"`
	ApplicationId string  `json:"application_id"`
	GuidId        *string `json:"guid_id,omitempty"`
}

type ApplicationCommandOptionChoice struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ApplicationCommandOption struct {
	Name        string                           `json:"name"`
	Description string                           `json:"description"`
	Type        int                              `json:"type"`
	Required    *bool                            `json:"required,omitempty"`
	Choices     []ApplicationCommandOptionChoice `json:"choices,omitempty"`
}

type RegisterApplicationCommandOptions struct {
	Name        string                     `json:"name"`
	Type        *int                       `json:"type,omitempty"`
	Description *string                    `json:"description,omitempty"`
	Options     []ApplicationCommandOption `json:"options,omitempty"`
}

type ApplicationCommandsClient service

func (c *ApplicationCommandsClient) Register(applicationId string, options *RegisterApplicationCommandOptions) (*ApplicationCommand, error) {
	u, err := c.client.BaseURL.Parse(fmt.Sprintf("applications/%s/commands", applicationId))
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if options != nil {
		buf = &bytes.Buffer{}
		err = json.NewEncoder(buf).Encode(options)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", c.client.botToken))

	res, err := c.client.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Check for 4xx or 5xx status codes
	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	var command *ApplicationCommand
	err = json.NewDecoder(res.Body).Decode(&command)
	if err == io.EOF {
		// ignore EOF errors caused by empty response body
		err = nil
	}
	if err != nil {
		return nil, err
	}

	return command, nil
}

const defaultBaseURL = "https://discord.com/api/v10/"

type service struct {
	client *Client
}

type Client struct {
	botToken string
	client   *http.Client

	BaseURL *url.URL

	// Reuse a single common service instead of one for each section of the API.
	common service

	ApplicationCommands *ApplicationCommandsClient
}

func NewClient(botToken string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		client:   http.DefaultClient,
		BaseURL:  baseURL,
		botToken: botToken,
	}
	c.common.client = c
	c.ApplicationCommands = (*ApplicationCommandsClient)(&c.common)

	return c
}
