package words

import (
	"math/rand"
	"strings"

	"github.com/brattonross/ghostedbot/internal/discord"
)

// Shuffle shuffles the words in a given string, using space as a delimiter.
func Shuffle(s string) string {
	words := strings.Split(s, " ")
	rand.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})
	return strings.Join(words, " ")
}

func ShuffleHandler(ctx *discord.InteractionContext) (*discord.InteractionResponse, error) {
	if len(ctx.Interaction.Data.Options) < 1 {
		return discord.MessageResponse("Please provide a string to shuffle."), nil
	}

	return discord.MessageResponse(Shuffle(ctx.Interaction.Data.Options[0].Value.(string))), nil
}
