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
