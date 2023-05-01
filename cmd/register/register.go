package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/brattonross/ghostedbot/internal/discord"
	"github.com/davecgh/go-spew/spew"
)

type commands struct {
	Global []*discord.RegisterApplicationCommandOptions `json:"global,omitempty"`
}

func main() {
	applicationId := os.Getenv("DISCORD_APPLICATION_ID")
	botToken := os.Getenv("DISCORD_BOT_TOKEN")

	if applicationId == "" {
		log.Fatal("Missing required environment variable DISCORD_APPLICATION_ID")
	}

	if botToken == "" {
		log.Fatal("Missing required environment variable DISCORD_BOT_TOKEN")
	}

	applicationCommandsSpecPath := "./config/application_commands.json"
	args := os.Args[1:]
	if len(args) > 0 {
		applicationCommandsSpecPath = args[0]
	}

	jsonFile, err := os.Open(applicationCommandsSpecPath)
	if err != nil {
		log.Fatalf("failed to open commands.json: %s\n", err)
	}

	defer jsonFile.Close()

	var commands commands
	err = json.NewDecoder(jsonFile).Decode(&commands)
	if err != nil {
		log.Fatalf("failed to decode commands.json: %s\n", err)
	}

	client := discord.NewClient(botToken)

	registeredCommands, err := client.ApplicationCommands.BulkOverwrite(applicationId, commands.Global)
	if err != nil {
		log.Fatalf("failed to bulk overwrite application commands: %s\n", err)
	}

	spew.Printf("registered commands: %+v\n", registeredCommands)
}
