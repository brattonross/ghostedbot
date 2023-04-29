package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/brattonross/ghostedbot/internal/discord"
)

// variables that get substituted by the deployment script (-ldflags).
var (
	buildHash          string
	buildDate          string
	formattedBuildDate string
)

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

type discordInteractionsRequestValidator interface {
	// validate returns an error if the request is not a valid interactions request.
	validate(r *http.Request) error
}

type ed25519Validator struct {
	publicKey ed25519.PublicKey
}

func (v *ed25519Validator) validate(r *http.Request) error {
	signature := r.Header.Get("X-Signature-Ed25519")
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return err
	}

	timestamp := r.Header.Get("X-Signature-Timestamp")
	message := bytes.NewBufferString(timestamp)

	var body bytes.Buffer
	// copy the body into both the message and body buffers,
	// the latter of which will be used to re-populate the request body.
	_, err = io.Copy(message, io.TeeReader(r.Body, &body))
	if err != nil {
		return err
	}

	defer r.Body.Close()
	defer func() {
		r.Body = io.NopCloser(&body)
	}()

	if ok := ed25519.Verify(v.publicKey, message.Bytes(), sig); !ok {
		return fmt.Errorf("invalid request signature")
	}

	return nil
}

var discordClient *discord.Client

type discordInteractionHandlerOptions struct {
	requestValidator discordInteractionsRequestValidator
}

func newDiscordInteractionHandler(opts discordInteractionHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Header().Add("Allow", http.MethodPost)
			return
		}

		err := opts.requestValidator.validate(r)
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
						Content: fmt.Sprintf("roastedbot: built at %s, using commit with SHA %s", formattedBuildDate, buildHash),
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

func main() {
	port := os.Getenv("PORT")
	applicationId := os.Getenv("DISCORD_APPLICATION_ID")
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	publicKey := os.Getenv("DISCORD_PUBLIC_KEY")

	pb, err := hex.DecodeString(publicKey)
	if err != nil {
		log.Fatalf("failed to decode public key: %s\n", err)
	}

	discordClient = discord.NewClient(botToken)

	_, err = discordClient.ApplicationCommands.Register(applicationId, &discord.RegisterApplicationCommandOptions{
		Name:        "version",
		Description: discord.String("Print version information."),
	})
	if err != nil {
		log.Fatalf("failed to register application command: %s\n", err)
	}

	http.HandleFunc("/interactions", newDiscordInteractionHandler(discordInteractionHandlerOptions{
		requestValidator: &ed25519Validator{publicKey: pb},
	}))

	if buildHash == "" {
		buildHash = "dev"
	}

	if buildDate == "" {
		buildDate = fmt.Sprintf("%d", time.Now().Unix())
	}

	epoch, err := strconv.ParseInt(buildDate, 10, 64)
	if err != nil {
		log.Fatalf("failed to parse build date: %s\n", err)
	}
	formattedBuildDate = time.Unix(epoch, 0).Format(time.RFC3339)

	log.Printf("starting roastedbot: built at %s using commit with SHA %s\n", formattedBuildDate, buildHash)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
