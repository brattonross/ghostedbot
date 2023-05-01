package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/brattonross/ghostedbot/internal/debug"
	"github.com/brattonross/ghostedbot/internal/discord"
	"github.com/brattonross/ghostedbot/internal/mdn"
)

func main() {
	port := os.Getenv("PORT")
	publicKey := os.Getenv("DISCORD_PUBLIC_KEY")

	pb, err := hex.DecodeString(publicKey)
	if err != nil {
		log.Fatalf("failed to decode public key: %s\n", err)
	}

	handler := discord.NewInteractionsHandler(pb)

	handler.RegisterApplicationCommandHandler("mdn", mdn.SearchHandler)

	handler.RegisterApplicationCommandHandler("test", func(ctx *discord.InteractionContext) (*discord.InteractionResponse, error) {
		return discord.MessageResponse("test successful <:AlienUnpleased:940285855292080149>"), nil
	})

	handler.RegisterApplicationCommandHandler("version", func(ctx *discord.InteractionContext) (*discord.InteractionResponse, error) {
		return discord.MessageResponse(fmt.Sprintf("Built %s using commit %s", debug.FormattedBuildDate(), debug.BuildHash)), nil
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
