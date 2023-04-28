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
)

type interaction struct {
	Type int `json:"type"`
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
	timestamp := r.Header.Get("X-Signature-Timestamp")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	buf := bytes.NewBufferString(timestamp)
	_, err = buf.Write(body)
	if err != nil {
		return err
	}

	sb, err := hex.DecodeString(signature)
	if err != nil {
		return err
	}

	if ok := ed25519.Verify(v.publicKey, buf.Bytes(), sb); !ok {
		return fmt.Errorf("invalid request signature")
	}

	return nil
}

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
			log.Printf("failed to validate request: %s\n", err)
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}

		var interaction interaction
		err = json.NewDecoder(r.Body).Decode(&interaction)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if interaction.Type == 1 {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "{\"type\": 1}")
			return
		}

		http.Error(w, "invalid interaction type", http.StatusBadRequest)
	}
}

func main() {
	port := os.Getenv("PORT")
	publicKey := os.Getenv("DISCORD_PUBLIC_KEY")

	pb, err := hex.DecodeString(publicKey)
	if err != nil {
		log.Fatalf("failed to decode public key: %s\n", err)
	}

	http.HandleFunc("/interactions", newDiscordInteractionHandler(discordInteractionHandlerOptions{
		requestValidator: &ed25519Validator{publicKey: pb},
	}))

	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		log.Fatalf("failed to start server: %s\n", err)
	}
}
