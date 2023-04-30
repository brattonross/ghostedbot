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

		handler(w, req)

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

		handler(w, req)

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

		handler(w, req)

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
