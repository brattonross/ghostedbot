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

func String(v string) *string {
	return &v
}

func Int(v int) *int {
	return &v
}

func Bool(v bool) *bool {
	return &v
}

const (
	InteractionTypePing               = 1
	InteractionTypeApplicationCommand = 2
)

type Interaction struct {
	Type int `json:"type"`
	// TODO: We can only handle ping and application commands
	// Maybe this should be interface{}, and then we can cast based on Type
	Data ApplicationCommandInteractionData `json:"data,omitempty"`
}

const (
	InteractionResponseTypePong                     = 1
	InteractionResponseTypeChannelMessageWithSource = 4
)

type InteractionResponseData struct {
	TTS             *bool         `json:"tts,omitempty"`
	Content         *string       `json:"content,omitempty"`
	Embeds          []interface{} `json:"embeds,omitempty"`
	AllowedMentions *interface{}  `json:"allowed_mentions,omitempty"`
	Flags           *int          `json:"flags,omitempty"`
	Components      []interface{} `json:"components,omitempty"`
	Attachments     []interface{} `json:"attachments,omitempty"`
}

type InteractionResponse struct {
	Type int                      `json:"type,omitempty"`
	Data *InteractionResponseData `json:"data,omitempty"`
}

// InteractionsRequestValidator validates incoming requests to the interactions endpoint.
type InteractionsRequestValidator interface {
	// Validate returns an error if the request is not a valid interactions request.
	Validate(r *http.Request) error
}

type InteractionsHandler struct {
	applicationCommands map[string]ApplicationCommandHandlerFunc
	validator           InteractionsRequestValidator
}

func (h *InteractionsHandler) handleUnhandledInteraction(w http.ResponseWriter, interaction *Interaction) {
	log.Printf("unhandled interaction: %+v", interaction)
	w.WriteHeader(http.StatusBadRequest)
}

func (h *InteractionsHandler) handlePingInteraction(w http.ResponseWriter, interaction *Interaction) {
	log.Printf("handling ping interaction: %+v", interaction)
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(&InteractionResponse{
		Type: InteractionResponseTypePong,
	})
	if err != nil {
		log.Printf("failed to encode response: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *InteractionsHandler) handleApplicationCommandInteraction(w http.ResponseWriter, interaction *Interaction) {
	handler, ok := h.applicationCommands[interaction.Data.Name]
	if !ok {
		h.handleUnhandledInteraction(w, interaction)
		return
	}

	log.Printf("handling application command: %+v", interaction)
	ctx := &InteractionContext{
		Interaction: interaction,
	}
	res, err := handler(ctx)
	if err != nil {
		log.Printf("failed to handle application command: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err = enc.Encode(res)
	if err != nil {
		log.Printf("failed to encode response: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *InteractionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Add("Allow", http.MethodPost)
		return
	}

	err := h.validator.Validate(r)
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

	switch {
	case interaction.Type == InteractionTypePing:
		h.handlePingInteraction(w, &interaction)
		return
	case interaction.Type == InteractionTypeApplicationCommand:
		h.handleApplicationCommandInteraction(w, &interaction)
		return
	}

	h.handleUnhandledInteraction(w, &interaction)
}

type InteractionContext struct {
	Interaction *Interaction
}

type ApplicationCommandHandlerFunc func(ctx *InteractionContext) (*InteractionResponse, error)

func (h *InteractionsHandler) RegisterApplicationCommandHandler(name string, handler ApplicationCommandHandlerFunc) {
	h.applicationCommands[name] = handler
}

func NewInteractionsHandler(validator InteractionsRequestValidator) *InteractionsHandler {
	return &InteractionsHandler{
		applicationCommands: make(map[string]ApplicationCommandHandlerFunc),
		validator:           validator,
	}
}

const (
	ApplicationCommandTypeChatInput = 1
	ApplicationCommandTypeUser      = 2
	ApplicationCommandTypeMessage   = 3
)

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

const (
	ApplicationCommandOptionTypeSubCommand      = 1
	ApplicationCommandOptionTypeSubCommandGroup = 2
	ApplicationCommandOptionTypeString          = 3
	ApplicationCommandOptionTypeInteger         = 4
	ApplicationCommandOptionTypeBoolean         = 5
	ApplicationCommandOptionTypeUser            = 6
	ApplicationCommandOptionTypeChannel         = 7
	ApplicationCommandOptionTypeRole            = 8
	ApplicationCommandOptionTypeMentionable     = 9
	ApplicationCommandOptionTypeNumber          = 10
	ApplicationCommandOptionTypeAttachment      = 11
)

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

type ApplicationCommandInteractionDataOption struct {
	Name  string      `json:"name"`
	Type  int         `json:"type"`
	Value interface{} `json:"value,omitempty"`
}

type ApplicationCommandInteractionData struct {
	Id       string                                    `json:"id"`
	Name     string                                    `json:"name"`
	Type     int                                       `json:"type"`
	Options  []ApplicationCommandInteractionDataOption `json:"options,omitempty"`
	GuildId  string                                    `json:"guild_id,omitempty"`
	TargetId string                                    `json:"target_id,omitempty"`
}

type ApplicationCommandsClient service

func (c *ApplicationCommandsClient) Register(applicationId string, options *RegisterApplicationCommandOptions) (*ApplicationCommand, error) {
	u, err := c.client.BaseURL.Parse(fmt.Sprintf("applications/%s/commands", applicationId))
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if options != nil {
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err = enc.Encode(options)
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

	var command ApplicationCommand
	err = json.NewDecoder(res.Body).Decode(&command)
	if err == io.EOF {
		// ignore EOF errors caused by empty response body
		err = nil
	}
	if err != nil {
		return nil, err
	}

	return &command, nil
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
		client:   &http.Client{},
		BaseURL:  baseURL,
		botToken: botToken,
	}
	c.common.client = c
	c.ApplicationCommands = (*ApplicationCommandsClient)(&c.common)

	return c
}
