package words

import (
	"math/rand"
	"strings"

	"github.com/brattonross/ghostedbot/internal/discord"
)

func LeftPad(s string, length int, char string) string {
	length = length - len(s)
	if length <= 0 {
		return s
	}
	if len(char) == 0 {
		char = " "
	}
	pad := ""
	for true {
		if length&1 == 1 {
			pad += char
		}
		length >>= 1
		if length == 0 {
			break
		}
		char += char
	}
	return pad + s
}

func LeftPadHandler(ctx *discord.InteractionContext) (*discord.InteractionResponse, error) {
	str := ctx.Interaction.Data.Options[0].Value.(string)
	length := int(ctx.Interaction.Data.Options[1].Value.(float64))
	char := ""
	if len(ctx.Interaction.Data.Options) > 2 {
		char = ctx.Interaction.Data.Options[2].Value.(string)
	}
	return discord.MessageResponse(LeftPad(str, length, char)), nil
}

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
