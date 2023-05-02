package checkem

import (
	"fmt"

	"github.com/brattonross/ghostedbot/internal/discord"
)

var checkemNames = map[int8]string{
	2:  "dubs",
	3:  "trips",
	4:  "quads",
	5:  "quints",
	6:  "sexts",
	7:  "septs",
	8:  "octs",
	9:  "nines",
	10: "decs",
}

const (
	over2Format  = "%s - <:EZ:1103063620209885214> <a:Clap:1103063782760124540> gratz on the %s"
	over10Format = "%s - <:Paggi:1103063622474792980> <a:Clap:1103063782760124540> you got more than 10 repeating digits?!"
)

func Checkem(id string) string {
	char := id[len(id)-1]
	var repeated int8
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] != char {
			break
		}
		repeated++
	}

	if repeated < 2 {
		return id
	}

	if repeated > 10 {
		return fmt.Sprintf(over10Format, id)
	}

	name := checkemNames[repeated]
	return fmt.Sprintf(over2Format, id, name)
}

func Handler(ctx *discord.InteractionContext) (*discord.InteractionResponse, error) {
	return discord.MessageResponse(Checkem(ctx.Interaction.Id)), nil
}
