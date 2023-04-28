package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"flag"
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

type config struct {
	Discord struct {
		PublicKey []byte `json:"public_key"`
	} `json:"discord"`
}

func main() {
	configPath := flag.String("config", "./config.json", "path to config file")
	flag.Parse()

	configFile, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("failed to load config file at \"%s\": %s\n", *configPath, err)
	}

	var cfg config
	err = json.Unmarshal(configFile, &cfg)
	if err != nil {
		log.Fatalf("failed to parse config file: %s\n", err)
	}

	http.HandleFunc("/", newDiscordInteractionHandler(discordInteractionHandlerOptions{
		requestValidator: &ed25519Validator{publicKey: cfg.Discord.PublicKey},
	}))

	if err := http.ListenAndServe(":8090", nil); err != nil {
		log.Fatalf("failed to start server: %s\n", err)
	}
}
