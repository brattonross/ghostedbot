package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// InteractionsRequestValidator validates incoming requests to the interactions endpoint.
type InteractionsRequestValidator interface {
	// Validate returns an error if the request is not a valid interactions request.
	Validate(r *http.Request) error
}

func NewInteractionsHandler(validator InteractionsRequestValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Header().Add("Allow", http.MethodPost)
			return
		}

		err := validator.Validate(r)
		if err != nil {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}

		var interaction Interaction
		err = json.NewDecoder(r.Body).Decode(&interaction)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("interaction received: %+v", interaction)
		w.WriteHeader(http.StatusOK)

		if interaction.Type == 1 {
			fmt.Fprint(w, "{\"type\": 1}")
			return
		}

		if interaction.Type == 2 {
			if interaction.Data.Name == "version" {
				log.Printf("handling version command")
				err = json.NewEncoder(w).Encode(InteractionResponse{
					Type: 4,
					Data: InteractionResponseData{
						// Content: fmt.Sprintf("roastedbot: built at %s, using commit with SHA %s", formattedBuildDate, buildHash),
						Content: "TODO",
					},
				})
				if err != nil {
					log.Printf("failed to encode interaction response: %s\n", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
		}

		log.Printf("unhandled interaction: %+v", interaction)
		err = json.NewEncoder(w).Encode(InteractionResponse{
			Type: 4,
			Data: InteractionResponseData{
				Content: "Sorry, I don't know how to handle that command.",
			},
		})
		if err != nil {
			log.Printf("failed to encode interaction response: %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func String(v string) *string {
	return &v
}

func Int(v int) *int {
	return &v
}

func Bool(v bool) *bool {
	return &v
}

type Interaction struct {
	Type int `json:"type"`
	Data struct {
		Name string `json:"name"`
	} `json:"data"`
}

type InteractionResponseData struct {
	Content string `json:"content"`
}

type InteractionResponse struct {
	Type int                     `json:"type"`
	Data InteractionResponseData `json:"data"`
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
