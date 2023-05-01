package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/brattonross/ghostedbot/internal/debug"
	"github.com/brattonross/ghostedbot/internal/discord"
)

func main() {
	port := os.Getenv("PORT")
	publicKey := os.Getenv("DISCORD_PUBLIC_KEY")

	pb, err := hex.DecodeString(publicKey)
	if err != nil {
		log.Fatalf("failed to decode public key: %s\n", err)
	}

	handler := discord.NewInteractionsHandler(pb)

	handler.RegisterApplicationCommandHandler("mdn", func(ctx *discord.InteractionContext) (*discord.InteractionResponse, error) {
		log.Printf("received command: %+v\n", ctx.Interaction)

		if len(ctx.Interaction.Data.Options) < 1 {
			return &discord.InteractionResponse{
				Type: discord.InteractionResponseTypeChannelMessageWithSource,
				Data: &discord.InteractionResponseData{
					Content: discord.String("Please provide a search query"),
				},
			}, nil
		}

		query := ctx.Interaction.Data.Options[0].Value.(string)
		res, err := http.Get(fmt.Sprintf("https://developer.mozilla.org/api/v1/search?q=%s&locale=en-US", query))
		if err != nil {
			return nil, fmt.Errorf("failed to search MDN: %w", err)
		}

		var searchResults struct {
			Documents []struct {
				Title string `json:"title"`
				Slug  string `json:"slug"`
			}
		}

		err = json.NewDecoder(res.Body).Decode(&searchResults)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal MDN search results: %w", err)
		}

		if len(searchResults.Documents) < 1 {
			return &discord.InteractionResponse{
				Type: discord.InteractionResponseTypeChannelMessageWithSource,
				Data: &discord.InteractionResponseData{
					Content: discord.String("No articles found"),
				},
			}, nil
		}

		return &discord.InteractionResponse{
			Type: discord.InteractionResponseTypeChannelMessageWithSource,
			Data: &discord.InteractionResponseData{
				Content: discord.String(searchResults.Documents[0].Title + ": https://developer.mozilla.org/en-US/docs/" + searchResults.Documents[0].Slug),
			},
		}, nil
	})

	handler.RegisterApplicationCommandHandler("test", func(ctx *discord.InteractionContext) (*discord.InteractionResponse, error) {
		return &discord.InteractionResponse{
			Type: discord.InteractionResponseTypeChannelMessageWithSource,
			Data: &discord.InteractionResponseData{
				Content: discord.String("test successful <:AlienUnpleased:940285855292080149>"),
			},
		}, nil
	})

	handler.RegisterApplicationCommandHandler("version", func(ctx *discord.InteractionContext) (*discord.InteractionResponse, error) {
		return &discord.InteractionResponse{
			Type: discord.InteractionResponseTypeChannelMessageWithSource,
			Data: &discord.InteractionResponseData{
				Content: discord.String(fmt.Sprintf("Built %s using commit %s", debug.FormattedBuildDate(), debug.BuildHash)),
			},
		}, nil
	})

	http.HandleFunc("/interactions", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Why do the interaction responses only become valid after
		// we use a recorder to write the response?
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)

		for k, v := range rec.Header() {
			w.Header()[k] = v
		}

		w.WriteHeader(rec.Code)

		log.Printf("response body: %s\n", rec.Body.String())
		_, err := rec.Body.WriteTo(w)
		if err != nil {
			log.Printf("failed to write response body: %s\n", err)
		}
	})

	log.Printf("starting roastedbot: built %s using commit with SHA %s\n", debug.FormattedBuildDate(), debug.BuildHash)

	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
