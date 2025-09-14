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

		// each command has its own function in commands.go,
		// just to keep the source a bit more legible
		// this switch is still necessary to call those functions, albeit a little spammy
		// eta when golang gets rust match-esque thingy to make this less ugly
		switch command := opts[1]; command {

		case "help":
			header, body := command_help() // call command function
			create_send_response(header, body, s, m) // have to pass in *discordgo.Session (s) and *discordgo.MessageCreate (m)

		case "install":
			header, body := command_install()
			create_send_response(header, body, s, m)

		case "docs":
			header, body := command_docs()
			create_send_response(header, body, s, m)

		case "discord":
			header, body := command_discord()
			create_send_response(header, body, s, m)

		case "github":
			header, body := command_github()
			create_send_response(header, body, s, m)

		case "redist":
			header, body := command_redist()
			create_send_response(header, body, s, m)

		case "repair":
			header, body := command_repair()
			create_send_response(header, body, s, m)

		case "dedicated":
			header, body := command_dedicated()
			create_send_response(header, body, s, m)

		case "vcredist":
			header, body := command_vcredist()
			create_send_response(header, body, s, m)

		case "unlockstats":
			header, body := command_unlockstats()
			create_send_response(header, body, s, m)

		case "performance":
			header, body := command_performance()
			create_send_response(header, body, s, m)

		case "fps":
			header, body := command_fps()
			create_send_response(header, body, s, m)

		case "fov":
			header, body := command_fov()
			create_send_response(header, body, s, m)

		case "nickname":
			header, body := command_nickname()
			create_send_response(header, body, s, m)

		case "console":
			header, body := command_console()
			create_send_response(header, body, s, m)

		case "directx":
			header, body := command_directx()
			create_send_response(header, body, s, m)

		default:
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
