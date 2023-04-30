package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
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

type ed25519Validator struct {
	publicKey ed25519.PublicKey
}

func (v *ed25519Validator) Validate(r *http.Request) error {
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

	validator := &ed25519Validator{publicKey: pb}
	handler := discord.NewInteractionsHandler(validator)

	// test command
	handler.RegisterApplicationCommandHandler("test", func(ctx *discord.InteractionContext) (*discord.InteractionResponse, error) {
		return &discord.InteractionResponse{
			Type: discord.InteractionResponseTypeChannelMessageWithSource,
			Data: &discord.InteractionResponseData{
				Content: "test successful <:AlienUnpleased:940285855292080149>",
			},
		}, nil
	})

	// version command
	_, err = discordClient.ApplicationCommands.Register(applicationId, &discord.RegisterApplicationCommandOptions{
		Name:        "version",
		Description: discord.String("Print version information."),
	})
	if err != nil {
		log.Fatalf("failed to register application command: %s\n", err)
	}

	handler.RegisterApplicationCommandHandler("version", func(ctx *discord.InteractionContext) (*discord.InteractionResponse, error) {
		return &discord.InteractionResponse{
			Type: discord.InteractionResponseTypeChannelMessageWithSource,
			Data: &discord.InteractionResponseData{
				Content: fmt.Sprintf("Built at %s using commit with SHA %s", formattedBuildDate, buildHash),
			},
		}, nil
	})

	http.HandleFunc("/interactions", handler.ServeHTTP)

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
