package main

import (
	"bytes"
	"crypto/ed25519"
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

type requestValidator interface {
	validate(r *http.Request) (bool, error)
}

type ed25519Validator struct {
	publicKey ed25519.PublicKey
}

func (v *ed25519Validator) validate(r *http.Request) (bool, error) {
	signature := r.Header.Get("X-Signature-Ed25519")
	timestamp := r.Header.Get("X-Signature-Timestamp")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false, err
	}

	buf := bytes.NewBufferString(timestamp)
	_, err = buf.Write(body)
	if err != nil {
		return false, err
	}

	return ed25519.Verify(v.publicKey, buf.Bytes(), []byte(signature)), nil
}

type discordInteractionHandlerOptions struct {
	requestValidator requestValidator
}

func newDiscordInteractionHandler(opts discordInteractionHandlerOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Header().Add("Allow", http.MethodPost)
			return
		}

		valid, err := opts.requestValidator.validate(r)
		if err != nil {
			log.Printf("failed to validate request: %s\n", err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if !valid {
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

	http.HandleFunc("/", newDiscordInteractionHandler(discordInteractionHandlerOptions{
		requestValidator: &ed25519Validator{publicKey: []byte(publicKey)},
	}))

	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		log.Fatalf("failed to start server: %s\n", err)
	}
}
