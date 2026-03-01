package main

import (
    "github.com/bwmarrin/discordgo"

    "path/filepath"
    "log/slog"
    "log"
    "os"
    "os/signal"
    "syscall"
    "strings"
    "time"
    "encoding/json"
)

// the iw4x domain is contained in a variable here to make it easier
// to change in the future, if there are any more "events"
const base_url string = "iw4x.io/" // this variable is global and applies to both util.go and commands.go

// we'll ignore any message that doesn't
// begin with this
const prefix string = "!iw4x"

// staff role ID for privileged command authentication
const staff_role_id string = "1111982635955277854"

// what message count to target to trigger a logfile cycle
const cycle_logcount int = 15000

func main() {
    log.Print("iw4x-discord-bot: startup")
	
    token := os.Getenv("IW4X_DISCORD_BOT_TOKEN") // the environment variable IW4X_DISCORD_BOT_TOKEN should hold the bot token

    // message logging stuff, this can be kept open so not inside of the handler
    location, err := os.Getwd() // get the directory the bot is being run from, we can just log to a file right next to the bin
    if err != nil {
        log.Print("iw4x-discord-bot: failed to get current working directory: ", err)
    }
	
    f, err := os.OpenFile(filepath.Join(location, "chatlog.json"), os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644) // the file to log to
    if err != nil {
        log.Fatal("iw4x-discord-bot: could not open logfile: ", err)
    }
    defer f.Close()

	// create our message logger, we want normal logs for everything but chat messages
    message_logger := slog.New(slog.NewJSONHandler(f, nil)) // f here is the file we opened above for logging

    // create directory for log archive if it doesn't exist already
    log_archive_dir := filepath.Join(location, "archive")
    if err := os.MkdirAll(log_archive_dir, 0755); err != nil {
        log.Fatal("iw4x-discord-bot: failed to create log archive directory: ", err)
    }

    // on first startup / restart we need to check how large the message database is
    message_count, err := get_logfile_length(location)
    if err != nil {
        log.Print("iw4x-discord-bot: failed to get logfile length: ", err)
        return
    }

    // this is the channel that will be used to listen for triggers to cycle logs
    logfile_channel := make(chan bool)

    // this will hand back the new value to main
    logfile_reset_channel := make(chan int)

    // create a thread for logfile cycling
    // this will perpetually listen and when it is sent a signal
    // it will cycle the logfile, this should never need to exit
    go func() {
        for {
            select {
            case <-logfile_channel: // listens for signal on logfile_channel chan
                logfile_cycle_timer := time.Now()
                if err := cycle_logfile(location, log_archive_dir); err != nil {
                    log.Print("iw4x-discord-bot: failed to cycle logfile", err)
                    continue
                }
                logfile_reset_channel <- 0
                logfile_cycle_duration := time.Since(logfile_cycle_timer)
                log.Print("iw4x-discord-bot: logfile cycle took: <", logfile_cycle_duration, ">")
            }
        }
    }()

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

        // log user message before doing anything else with it
        message_logger.Info(
            "message-logger",
            "type", "message",
            "content", m.Content,
            "message_ID", m.ID,
            "channel_ID", m.ChannelID,	
            "author_ID", m.Author.ID,
            "author_username", m.Author.Username,
            "author_nickname", m.Author.GlobalName,
            "attachments", m.Attachments,
        )

        // add to the message count
        message_count++
        if (message_count >= cycle_logcount) {
            log.Print("iw4x-discord-bot: logfile has exceeded <", cycle_logcount, ">: triggering cycle")

            // if the message count exceeds cycle_logcount, signal another thread to
            // cycle the logfile- we do this in another thread to prevent this process
            // from hanging up the bots message handling
            logfile_channel <- true
            message_count = <-logfile_reset_channel
        }

        // split up user message by spaces
        opts := strings.Split(m.Content, " ")

        // if the first opt here isn't the bot prefix, do nothing
        if opts[0] != prefix {
            return
        }

        if len(opts) < 2 {
            header := "Not enough arguments!"
            body := "Expected `!iw4x <option>`.\nSee `!iw4x help` for more information on valid commands."

            if err := create_send_response(header, body, s, m); err != nil {
                log.Print("iw4x-discord-bot: failed to send command response: ", err)
                return
            }

            log.Print("iw4x-discord-bot: invalid command issued by user: <" + m.Author.ID + ":" + m.Author.Username + ">")
            return
        }
        
        // staff-only commands
        if check_permissions(m) {
            switch staff_command := opts[1]; staff_command {
            case "restart":
                command_timer := time.Now()
                log.Print("iw4x-discord-bot: staff member: <" + m.Author.ID + ":" + m.Author.Username + "> triggered restart")
                
                s.ChannelMessageSend(m.ChannelID, "gn")
                log.Print("iw4x-discord-bot: closing session")
                session.Close()
                
                command_duration := time.Since(command_timer)
                log.Print("iw4x-discord-bot: session closed after: ", command_duration, ", goodnight!")
                
                os.Exit(0)
            case "staffhelp":
                command_timer := time.Now()
                log.Print("iw4x-discord-bot: staff member: <" + m.Author.ID + ":" + m.Author.Username + "> requested staffhelp")

                header, body := command_staffhelp()

                if err := create_send_response(header, body, s, m); err != nil {
                    log.Print("iw4x-discord-bot: failed to send command response: ", err)
                    return
                }
                
                command_duration := time.Since(command_timer)
                log.Print("iw4x-discord-bot: response to command: 'staffhelp' from staff member: <" + m.Author.ID + ":" + "m.Author.Username" + "> sent in: <", command_duration, ">")
                return

            case "querydb":
                if len(opts) < 3 {
                    header := "Not enough arguments!"
                    body := "Expected `!iw4x querydb <opts>`.\nSee `!iw4x staffhelp` for more information on valid commands."

                    if err := create_send_response(header, body, s, m); err != nil {
                        log.Print("iw4x-discord-bot: failed to send command response: ", err)
                    }
                    return
                }
                
                command_timer := time.Now()
                log.Print("iw4x-discord-bot: staff member: <" + m.Author.ID + ":" + m.Author.Username + "> requested querydb")

                query_results, err := query_db(location, opts[2:]) // pass in only opts *after* '!iw4x querydb' as those are useless here.
                if err != nil {
                    s.ChannelMessageSend(m.ChannelID, err.Error())
                    log.Print("iw4x-discord-bot: failed to query database: ", err)
                    return
                }

                // make results pretty for staff readability
                // convert []string to []json.RawMessage
                raw_objects := make([]json.RawMessage, len(query_results))
                for i, str := range query_results {
                    raw_objects[i] = json.RawMessage(str)
                }

                // we can now use MarshalIndent to "blow out" the structure of the message
                pretty_query_results, err := json.MarshalIndent(raw_objects, "", "  ")
                if err != nil {
                    log.Print("iw4x-discord-bot: failed to make query results pretty: ", err)
                }
                
                // write query results to file
                if err := os.WriteFile("/tmp/queryresults.json", pretty_query_results, 0644); err != nil { 
                    log.Print("iw4x-discord-bot: failed to write query results to temporary file: ", err)
                    return
                }
                
                // upload file to discord
                if err := create_send_query(s, m); err != nil {
                    log.Print("iw4x-discord-bot: failed to upload query results to discord: ", err)
                    return
                }

                command_duration := time.Since(command_timer)
                log.Print("iw4x-discord-bot: response to command: 'querydb' from staff member: <" + m.Author.ID + ":" + m.Author.Username + "> sent in: <", command_duration, ">")

                return
                
            case "logstat":
                command_timer := time.Now()
                log.Print("iw4x-discord-bot: staff member: <" + m.Author.ID + ":" + m.Author.Username + "> requested logstat")

                header, body := command_logstat(message_count, location)

                if err := create_send_response(header, body, s, m); err != nil {
                    log.Print("iw4x-discord-bot: failed to send command response: ", err)
                    return
                }
                
                command_duration := time.Since(command_timer)
                log.Print("iw4x-discord-bot: response to command: 'logstat' from staff member: <" + m.Author.ID + ":" + "m.Author.Username" + "> sent in: <", command_duration, ">")

                return
            }
        }

        // this is done after the staff commands section because certain staff-only commands
        // such as querydb expect a ton of opts as opposed to just requesting a help output
        if len(opts) > 2 { // if too many opts are given, return
            header := "Too many arguments!"
            body := "Expected `!iw4x <option>`.\nSee `!iw4x help` for more information on valid commands."
            create_send_response(header, body, s, m)
            log.Print("iw4x-discord-bot: invalid command issued by user: <" + m.Author.ID + ":" + m.Author.Username + ">")
            return
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
            if err := create_send_status(s); err != nil {
                log.Print("iw4x-discord-bot: failed to send bot status: ", err)
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
                if err := create_send_status(s); err != nil {
                    log.Print("iw4x-discord-bot: failed to send bot status: ", err)
                }
            case _, _ = <-stale: // this allows the thread to be killed by the new thread
                return
            }
        }
    })

	// log message deletion
	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) {
        message_logger.Info(
            "message-logger",
            "type", "deletion",
            "message_ID", m.ID,
            "channel_ID", m.ChannelID,
        )

        return
    })

	// log message edits
	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
        message_logger.Info(
            "message-logger",
            "type", "edit",
            "content", m.Content,
            "message_ID", m.ID,
            "channel_ID", m.ChannelID,
            "author_ID", m.Author.ID,
            "author_username", m.Author.Username,
            "author_nickname", m.Author.GlobalName,
        )

        return
    })
	
    // tell discord our intent
    session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

    // open discord session
    if err = session.Open(); err != nil {
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
