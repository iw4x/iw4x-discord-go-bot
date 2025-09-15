package main

import (
	"github.com/bwmarrin/discordgo"

	"log"
	"os"
	"os/signal"
	"syscall"
	"strings"
)

// we'll ignore any message that doesn't
// begin with this
const prefix string = "!iw4x"

// builds embeds and sends output for all commands to reduce spam in switch statement below
// header and body are passed into this from the switch statement below, and the switch statement below
// fetches this information from each commands function in commands.go
func create_send_response(header string, body string, s *discordgo.Session, m *discordgo.MessageCreate) {
	embed := &discordgo.MessageEmbed {
		Title: header,
		Description: body,
		Color: 0x0ff00,
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)

	return
}

func main() {
	token := os.Getenv("IW4X_DISCORD_BOT_TOKEN") // the environment variable IW4X_DISCORD_BOT_TOKEN should hold the bot token

	// spawn a new session
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	// session needs a handler to process incoming messages
	// and decide what to do with them
	// s == pointer to discordgo.Session, m == the event we want to handle with this
	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

		// if the author ID is the same as the session ID, do nothing
		if m.Author.ID == s.State.User.ID {
			return
		}

		// split up user message by spaces
		opts := strings.Split(m.Content, " ")

		// if the first opt here isn't the bot prefix, do nothing
		if opts[0] != prefix {
			return
		}

		// if nothing is given after the prefix, return
		if len(opts) < 2 {
			header := "Not enough arguments."
			body := "Expected `!iw4x <option>`.\nSee `!iw4x help` for more information on valid commands."
			create_send_response(header, body, s, m)
			return
		}

		// function map, maps key (user input command) to value pair (function name)
		// every function in this map will return (string, string) - header, body
		commands := map[string]func() (string, string) {
			"help": command_help,
			"install": command_install,
			"docs": command_docs,
			"discord": command_discord,
			"github": command_github,
			"redist": command_redist,
			"repair": command_repair,
			"dedicated": command_dedicated,
			"vcredist": command_vcredist,
			"unlockstats": command_unlockstats,
			"performance": command_performance,
			"fps": command_fps,
			"fov": command_fov,
			"nickname": command_nickname,
			"console": command_console,
			"directx": command_directx,
		}

		// `command` here is the keys associated value if the key exists, in this case a function name
		// this checks to see if opts[1] (user input post-prefix) has a matching key
		// in the function map, `exists` is what is being tested here
		if command, exists := commands[opts[1]]; exists {
			header, body := command() // calls `command` as a function, of which will be one of the matching key values
			create_send_response(header, body, s, m)
		} else {
			header := "Invalid option"
			body := "Invalid bot command: `" + opts[1] + "`"
			create_send_response(header, body, s, m)
		}

	})

	// tell discord our intent
	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	// open discord session
	err = session.Open()
	if err != nil {
		log.Fatal(err)
	}

	// when the bot terminates, close the session with discord
	defer session.Close()

	log.Print("iw4x-discord-bot: active")

	// this allows the bot to be ctrl+c'd to kill it
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Print("iw4x-discord-bot: shutdown")
}
