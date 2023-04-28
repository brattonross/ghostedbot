package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type passingValidator struct{}

func (v *passingValidator) validate(r *http.Request) (bool, error) {
	return true, nil
}

type failingValidator struct{}

func (v *failingValidator) validate(r *http.Request) (bool, error) {
	return false, fmt.Errorf("failed to validate request")
}

func TestDiscordInteractionHandler(t *testing.T) {
	t.Run("Returns 401 if sent invalid headers", func(t *testing.T) {
		b, err := json.Marshal(&interaction{Type: 1})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
		w := httptest.NewRecorder()

		handler := newDiscordInteractionHandler(discordInteractionHandlerOptions{
			requestValidator: &failingValidator{},
		})

		handler(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected response status code %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("Returns 405 if sent invalid method", func(t *testing.T) {
		b, err := json.Marshal(&interaction{Type: 1})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(b))
		w := httptest.NewRecorder()

		handler := newDiscordInteractionHandler(discordInteractionHandlerOptions{
			requestValidator: &passingValidator{},
		})

		handler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected response status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}

		if w.Header().Get("Allow") != http.MethodPost {
			t.Errorf("expected Allow header to be %s, got %s", http.MethodPost, w.Header().Get("Allow"))
		}
	})
}
