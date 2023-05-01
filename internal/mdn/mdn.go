package mdn

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/brattonross/ghostedbot/internal/discord"
)

type document struct {
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

type searchResponse struct {
	Documents []*document `json:"documents"`
}

func search(query string) (*searchResponse, error) {
	res, err := http.Get(fmt.Sprintf("https://developer.mozilla.org/api/v1/search?q=%s&locale=en-US", query))
	if err != nil {
		return nil, fmt.Errorf("failed to search MDN: %w", err)
	}

	var searchResults searchResponse
	err = json.NewDecoder(res.Body).Decode(&searchResults)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal MDN search results: %w", err)
	}

	return &searchResults, nil
}

// SearchHandler is a discord application command handler that searches MDN for a given query.
func SearchHandler(ctx *discord.InteractionContext) (*discord.InteractionResponse, error) {
	if len(ctx.Interaction.Data.Options) < 1 {
		return discord.MessageResponse("Please provide a search query"), nil
	}

	query := ctx.Interaction.Data.Options[0].Value.(string)
	resp, err := search(query)
	if err != nil {
		return nil, err
	}

	if len(resp.Documents) < 1 {
		return discord.MessageResponse("No articles found"), nil
	}

	message := fmt.Sprintf("%s: https://developer.mozilla.org/en-US/docs/%s", resp.Documents[0].Title, resp.Documents[0].Slug)
	return discord.MessageResponse(message), nil
}
