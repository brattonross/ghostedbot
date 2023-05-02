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

var checkemNames = map[int8]string{
	2:  "dubs",
	3:  "trips",
	4:  "quads",
	5:  "quints",
	6:  "sexts",
	7:  "septs",
	8:  "octs",
	9:  "nines",
	10: "decs",
}

func main() {
	port := os.Getenv("PORT")
	publicKey := os.Getenv("DISCORD_PUBLIC_KEY")

	pb, err := hex.DecodeString(publicKey)
	if err != nil {
		log.Fatalf("failed to decode public key: %s\n", err)
	}

	handler := discord.NewInteractionsHandler(pb)

	handler.RegisterApplicationCommandHandler("checkem", func(ctx *discord.InteractionContext) (*discord.InteractionResponse, error) {
		id := ctx.Interaction.Id
		char := id[len(id)-1]
		var repeated int8
		for i := len(id) - 1; i >= 0; i-- {
			if id[i] != char {
				break
			}
			repeated++
		}

		if repeated < 2 {
			return discord.MessageResponse(id), nil
		}

		if repeated > 10 {
			message := fmt.Sprintf("%s - <:Paggi:1103063622474792980> <a:Clap:1103063782760124540> you got more than 10 repeating digits?!", id)
			return discord.MessageResponse(message), nil
		}

		name := checkemNames[repeated]
		message := fmt.Sprintf("%s - <:EZ:1103063620209885214> <a:Clap:1103063782760124540> gratz on the %s", id, name)
		return discord.MessageResponse(message), nil
	})

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
