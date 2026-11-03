package main

import (
    "github.com/bwmarrin/discordgo"

    "path/filepath"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

// the iw4x domain is contained in a variable here to make it easier
// to change in the future, if there are any more "events"
const base_url string = "iw4x.io/" // this variable is global and applies to both util.go and commands.go

// where the x64 platform serves its player counts
const x64_population_url string = "http://147.135.10.99:8443/v1/population"

// staff role ID for privileged command authentication
const staff_role_id string = "1111982635955277854"

// the honeypot channel ID
// the bot will watch this channel and autoban anyone who sends a message there
const honeypot_channel string = "1524421140225720430"

func main() {
    log.Print("iw4x-discord-bot: startup")

    uptime := time.Now()

    token := os.Getenv("IW4X_DISCORD_BOT_TOKEN") // the environment variable IW4X_DISCORD_BOT_TOKEN should hold the bot token
    if token == "" {
        log.Print("iw4x-discord-bot: failed to acquire bot token, cannot continue")
        return
    }

    location, err := os.Getwd()
    if err != nil {
        log.Print("iw4x-discord-bot: failed to get current working directory: ", err)
    }

    // buffered so the message handler doesn't block on a honeypot ban
    // this is a pretty large buffer
    honeypot_bans := make(chan string, 256)

    // spawn a new session
    session, err := discordgo.New("Bot " + token)
    if err != nil {
        log.Fatal(err)
    }

    // all bot commands are slash commands, see slash.go
    // s == pointer to discordgo.Session, i == the event we want to handle with this
    session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        handle_interaction(s, i, uptime)
    })

    // the message handler is only used for the honeypot channel now
    // message content isnt needed for that to work so the priv'd
    // message content intent isnt required anymore
    session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

        // if the author ID is the same as the session ID, do nothing
        if m.Author.ID == s.State.User.ID {
            return
        }

        // if the message is in the honeypot channel just autoban
        // and return quickly, further processing not needed
        // the 1 in the arguments here is to delete the past 1 day of history from this user
        if m.ChannelID == honeypot_channel {
            if check_permissions(m.Member) { // skip staff
                return
            }

            if err := s.GuildBanCreateWithReason(m.GuildID, m.Author.ID, "User sent message in the honeypot channel.", 1); err != nil {
                log.Print("iw4x-discord-bot: failed to ban user in honeypot channel: ", err)
                return
            }
            log.Print("iw4x-discord-bot: honeypot autobanned user: <" + m.Author.ID + ":" + m.Author.Username + ">")

            select {
            case honeypot_bans <- m.Author.ID:
            default:
                // just a theoretical, if the counter is backed up just skip the increment,
                // the ban already happened
                log.Print("iw4x-discord-bot: honeypot counter queue full, dropping increment")
            }
            return
        }
    })

    // this is set to trigger on Ready to set bot status
    var stale_status chan bool // this is nil on first run
    // when discord sends a new ready, the new thread will signal the old one to stop
    session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
        if stale_status != nil { close(stale_status) } // send signal if non nil
        stale_status = make(chan bool) // this sets it non nil

        // we'll fetch players immediately on ready, and then once every 1.5 minutes
        // after this seems to fail occasionally, just rerun until it doesn't, but
        // wait before retrying to avoid tight loop
        for {
            if err := create_send_status(s); err != nil {
                log.Print("iw4x-discord-bot: failed to send bot status: ", err)
            } else {
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
            case _, _ = <-stale_status: // this allows the thread to be killed by the new thread
                return
            }
        }
    })

    // this is pretty much the same thing as above but for steam sales instead
    // this file exists while an announced sale is on going so a restart doesn't reannounce
    sale_marker := filepath.Join(location, "sale_announced")
    var stale_sale chan bool
    session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
        if stale_sale != nil { close(stale_sale) }
        stale_sale = make(chan bool)

        sale_ticker := time.NewTicker(1 * time.Hour)
        for {
            select {
            case <-sale_ticker.C:
                sale_percentage, err := fetch_sale()
                if err != nil {
                    log.Print("iw4x-discord-bot: failed to check for steam sale: ", err)
                    continue // a failed check doesn't mean the sale ended
                }

                _, err = os.Stat(sale_marker)
                is_currently_on_sale := err == nil

                if sale_percentage != "0" {
                    if !is_currently_on_sale {
                        // convert percentage int to string for output
                        output := "Call Of Duty: Modern Warfare 2 is on sale on [Steam](https://store.steampowered.com/app/10180/Call_of_Duty_Modern_Warfare_2_2009/) for " + sale_percentage + "% off!"

                        _, err := s.ChannelMessageSend("1479998471812546662", output)
                        if err != nil {
                            log.Print("iw4x-discord-bot: failed to send steam sale notification: ", err)
                        }
                        if err := os.WriteFile(sale_marker, nil, 0644); err != nil {
                            log.Print("iw4x-discord-bot: failed to create sale marker")
                        }
                    }
                } else {
                    if is_currently_on_sale {
                        _, err := s.ChannelMessageSend("1479998471812546662", "The previously announced sale has ended.")
                        if err != nil {
                            log.Print("iw4x-discord-bot: failed to send steam sale notification: ", err)
                        }
                    }
                    if err := os.Remove(sale_marker); err != nil {
                        log.Print("iw4x-discord-bot: failed to remove sale marker")
                    }
                }

            case _, _ = <-stale_sale:
                return
            }
        }
    })

    session.AddHandler(func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
        // ping new users on join to tell them about bot commands
        joiner_id := m.User.ID
        if err := send_join_message(s, joiner_id); err != nil {
            log.Print("iw4x-discord-bot: failed to send join message: ", err)
        }

        return
    })

    // tell discord our intent
    session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMembers

    // open discord session
    if err = session.Open(); err != nil {
        log.Fatal(err)
    }

    log.Print("iw4x-discord-bot: active")

    // register slash commands with discord
    // this happens every startup so itll overwrite whatever got registered previously
    if err := register_commands(session); err != nil {
        log.Print("iw4x-discord-bot: failed to register slash commands: ", err)
    } else {
        log.Print("iw4x-discord-bot: registered slash commands")
    }

    // we can just start the counter here, its fine as soon as
    // the session is ready
    go run_honeypot_counter(session, location, honeypot_bans)

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
