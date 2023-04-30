package discord_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/brattonross/ghostedbot/internal/discord"
)

type passingValidator struct{}

func (v *passingValidator) Validate(r *http.Request) error {
	return nil
}

type failingValidator struct{}

func (v *failingValidator) Validate(r *http.Request) error {
	return fmt.Errorf("test error")
}

func TestNewInteractionsHandler(t *testing.T) {
	t.Run("Returns 401 if sent invalid headers", func(t *testing.T) {
		b, err := json.Marshal(&discord.Interaction{Type: discord.InteractionTypePing})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, "/interactions", bytes.NewReader(b))
		w := httptest.NewRecorder()

		handler := discord.NewInteractionsHandler(&failingValidator{})

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected response status code %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("Returns 405 if sent invalid method", func(t *testing.T) {
		b, err := json.Marshal(&discord.Interaction{Type: discord.InteractionTypePing})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodGet, "/interactions", bytes.NewReader(b))
		w := httptest.NewRecorder()

		handler := discord.NewInteractionsHandler(&passingValidator{})

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected response status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}

		if w.Header().Get("Allow") != http.MethodPost {
			t.Errorf("expected Allow header to be %s, got %s", http.MethodPost, w.Header().Get("Allow"))
		}
	})

	t.Run("Correctly handles PING interaction", func(t *testing.T) {
		b, err := json.Marshal(&discord.Interaction{Type: discord.InteractionTypePing})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, "/interactions", bytes.NewReader(b))
		w := httptest.NewRecorder()

		handler := discord.NewInteractionsHandler(&passingValidator{})

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected response status code %d, got %d", http.StatusOK, w.Code)
		}

		var response discord.InteractionResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}

		if response.Type != discord.InteractionResponseTypePong {
			t.Errorf("expected response type %d, got %d", discord.InteractionResponseTypePong, response.Type)
		}
	})

	t.Run("Correctly handles invalid interaction", func(t *testing.T) {
		b, err := json.Marshal(&discord.Interaction{Type: 999})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, "/interactions", bytes.NewReader(b))
		w := httptest.NewRecorder()

		handler := discord.NewInteractionsHandler(&passingValidator{})

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected response status code %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Correctly handles valid application command interaction", func(t *testing.T) {
		b, err := json.Marshal(&discord.Interaction{
			Type: discord.InteractionTypeApplicationCommand,
			Data: discord.ApplicationCommandInteractionData{
				Id:   "1234567890",
				Name: "blep",
				Options: []discord.ApplicationCommandInteractionDataOption{
					{
						Name:  "animal",
						Value: "animal_dog",
					},
					{
						Name:  "only_smol",
						Value: "true",
					},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, "/interactions", bytes.NewReader(b))
		w := httptest.NewRecorder()

		handler := discord.NewInteractionsHandler(&passingValidator{})
		handler.RegisterApplicationCommandHandler("blep", func(ctx *discord.InteractionContext) (*discord.InteractionResponse, error) {
			if ctx.Interaction.Data.Options[0].Value != "animal_dog" {
				t.Errorf("expected option value %s, got %s", "animal_dog", ctx.Interaction.Data.Options[0].Value)
			}

			if ctx.Interaction.Data.Options[1].Value != "true" {
				t.Errorf("expected option value %s, got %s", "true", ctx.Interaction.Data.Options[1].Value)
			}

			return &discord.InteractionResponse{
				Type: discord.InteractionResponseTypeChannelMessageWithSource,
				Data: &discord.InteractionResponseData{
					Content: discord.String("You requested a dog"),
				},
			}, nil
		})

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected response status code %d, got %d", http.StatusOK, w.Code)
		}

		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("expected response content type %s, got %s", "application/json", w.Header().Get("Content-Type"))
		}

		var response discord.InteractionResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}

		if response.Type != discord.InteractionResponseTypeChannelMessageWithSource {
			t.Errorf("expected response type %d, got %d", discord.InteractionResponseTypeChannelMessageWithSource, response.Type)
		}

		if *response.Data.Content != "You requested a dog" {
			t.Errorf("expected response content %s, got %s", "You requested a dog", *response.Data.Content)
		}
	})
}

func TestClientRegisterGlobalApplicationCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected request method %s, got %s", http.MethodPost, r.Method)
		}

		if r.URL.Path != "/applications/1234567890/commands" {
			t.Errorf("expected request path %s, got %s", "/applications/1234567890/commands", r.URL.Path)
		}

		w.Write([]byte(`{"id": "1234567890", "application_id": "1234567890", "name": "blep", "description": "Send a random adorable animal photo", "version": "1234567890", "default_permission": true}`))
	}))
	defer server.Close()

	client := discord.NewClient("1234567890")
	serverURL, _ := url.Parse(server.URL)
	client.BaseURL = serverURL

	command, err := client.ApplicationCommands.Register("1234567890", &discord.RegisterApplicationCommandOptions{
		Name:        "blep",
		Type:        discord.Int(discord.ApplicationCommandTypeChatInput),
		Description: discord.String("Send a random adorable animal photo"),
		Options: []discord.ApplicationCommandOption{
			{
				Name:        "animal",
				Description: "The type of animal",
				Type:        discord.ApplicationCommandOptionTypeString,
				Required:    discord.Bool(true),
				Choices: []discord.ApplicationCommandOptionChoice{
					{
						Name:  "Dog",
						Value: "animal_dog",
					},
					{
						Name:  "Cat",
						Value: "animal_cat",
					},
					{
						Name:  "Penguin",
						Value: "animal_penguin",
					},
				},
			},
			{
				Name:        "only_smol",
				Description: "Whether to show only baby animals",
				Type:        discord.ApplicationCommandOptionTypeBoolean,
				Required:    discord.Bool(false),
			},
		},
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if command.Id != "1234567890" {
		t.Errorf("expected command ID %s, got %s", "1234567890", command.Id)
	}
}
