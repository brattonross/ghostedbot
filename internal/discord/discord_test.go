package discord_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/brattonross/ghostedbot/internal/discord"
)

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
		Type:        discord.Int(1),
		Description: discord.String("Send a random adorable animal photo"),
		Options: []discord.ApplicationCommandOption{
			{
				Name:        "animal",
				Description: "The type of animal",
				Type:        3,
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
				Type:        5,
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
