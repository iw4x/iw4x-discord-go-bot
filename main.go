package main

import (
    "github.com/bwmarrin/discordgo"

    "log"
    "os"
    "os/signal"
    "syscall"
    "strings"
    "time"
)

// the iw4x domain is contained in a variable here to make it easier
// to change in the future, if there are any more "events"
const base_url string = "iw4x.io/" // this variable is global and applies to both util.go and commands.go

// we'll ignore any message that doesn't
// begin with this
const prefix string = "!iw4x"

func main() {
    log.Print("iw4x-discord-bot: startup")

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
            header := "Not enough arguments!"
            body := "Expected `!iw4x <option>`.\nSee `!iw4x help` for more information on valid commands."
            create_send_response(header, body, s, m)
            log.Print("iw4x-discord-bot: invalid command issued by user: <" + m.Author.ID + ":" + m.Author.Username + ">")
            return
        } else if len(opts) > 2 { // if too many opts are given, return
            header := "Too many arguments!"
            body := "Expected `!iw4x <option>`.\nSee `!iw4x help` for more information on valid commands."
            create_send_response(header, body, s, m)
            log.Print("iw4x-discord-bot: invalid command issued by user: <" + m.Author.ID + ":" + m.Author.Username + ">")
            return
        }

        // staff-only commands
        if check_permissions(m) {
            switch staff_command := opts[1]; staff_command {
            case "restart":
                log.Print("iw4x-discord-bot: staff member: <" + m.Author.ID + "> triggered restart")
                s.ChannelMessageSend(m.ChannelID, "gn")
                session.Close()
                os.Exit(0)
            }
        }

        // function map, maps key (user input command) to value pair (function name)
        // every function in this map will return (string, string) - header, body
        // this could probably be a map of structs with predefined header/body values instead
        // but this allows for additional complexity with the command output if needed and doesn't slow down too much
        commands := map[string]func() (string, string) {
            "help": command_help,
            "install": command_install,
            "docs": command_docs,
            "discord": command_discord,
            "github": command_github,
            "repair": command_repair,
            "dedicated": command_dedicated,
            "vcredist": command_vcredist,
            "unlockstats": command_unlockstats,
            "performance": command_performance,
            "fps": command_fps,
            "fov": command_fov,
            "nickname": command_nickname,
            "console": command_console,
            "dxr": command_dxr,
            "rawfiles": command_rawfiles,
            "game": command_game,
            "dxvk": command_dxvk,
            "dlc": command_dlc,
        }

        // `command` here is the keys associated value if the key exists, in this case a function name
        // this checks to see if opts[1] (user input post-prefix) has a matching key
        // in the function map, `exists` is what is being tested here
        if command, exists := commands[opts[1]]; exists {
            command_timer := time.Now() // starts a timer
            log.Print("iw4x-discord-bot: command: '" + opts[1] + "' requested by user: <" + m.Author.ID + ":" + m.Author.Username + ">")
            header, body := command() // calls `command` as a function, of which will be one of the matching key values
            create_send_response(header, body, s, m)
            command_duration := time.Since(command_timer)
            log.Print("iw4x-discord-bot: response to command: '" + opts[1] + "' from user: <" + m.Author.ID +  ":" + m.Author.Username + "> sent in: <",  command_duration,  ">")
            return
        } else {
            log.Print("iw4x-discord-bot: invalid command issued by user: <" + m.Author.ID + ":" + m.Author.Username + ">")
            header := "Invalid option!"
            body := "Invalid bot command: `" + opts[1] + "`\nSee `!iw4x help` for more information on valid commands."
            create_send_response(header, body, s, m)
            return
        }

    })

    // since the above is set to trigger on message send,
    // this is set to trigger on Ready to set bot status
    var stale chan bool // this is nil on first run
    // when discord sends a new ready, the new thread will signal the old one to stop
    session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
        if stale != nil { close(stale) } // send signal if non nil
        stale = make(chan bool) // this sets it non nil

        // we'll fetch players immediately on ready, and then once every 1.5 minutes
        // after this seems to fail occasionally, just rerun until it doesn't, but
        // wait before retrying to avoid tight loop
        for {
            if create_send_status(s) {
                break
            }
            time.Sleep(5 * time.Second)
        }

        // every 1.5 minutes
        status_ticker := time.NewTicker(90 * time.Second)
        for {
            // this select will perpetually poll for events on both of these channels
            // when this returns, and a new handler is spawned, it will kill the old
            // handler thread
            select {
            case <-status_ticker.C: // listens for signal on timer
                create_send_status(s)
            case _, _ = <-stale: // this allows the thread to be killed by the new thread
                return
            }
        }
    })

    // tell discord our intent
    session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

    // open discord session
    err = session.Open()
    if err != nil {
        log.Fatal(err)
    }

    log.Print("iw4x-discord-bot: active")

    // when this function returns, close the session with discord
    defer session.Close()

    // this allows the bot to be ctrl+c'd
    // this isn't a graceful shutdown
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    <-sc

    log.Print("iw4x-discord-bot: shutdown")

    return
}
