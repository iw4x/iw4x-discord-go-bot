package main

// this file contains all of the slash command shit
// their definitions, registration with discord, and the interaction handling
// if you're looking for the actual command output, look at commands.go

import (
    "github.com/bwmarrin/discordgo"

    "log"
    "os"
    "slices"
    "time"
)

const command_name string = "iw4x"
const staff_command_name string = "iw4x-staff"

// deferred is for commands that make like network requests and stuff and may take longer
// than the 3 seconds discord gives us to reply (/iw4x stats for example)
type subcommand struct {
    name, description string
    run func() (string, string) // returns header, body
    deferred bool
}

// the order these are in here is the order discord will display them in
// also, discord for whatever reason has a limit on 25 subcommands per parent command
// and descriptions can be at most 100 chars
var subcommands = []subcommand {
    { "help", "Displays the help dialog", command_help, false },
    { "install", "Information about installing IW4x", command_install, false },
    { "docs", "IW4x documentation", command_docs, false },
    { "discord", "Displays the invite link for this server", command_discord, false },
    { "github", "Displays the link to IW4x source code", command_github, false },
    { "repair", "Information on repairing your game files", command_repair, false },
    { "dedicated", "Information on setting up a dedicated server", command_dedicated, false },
    { "vcredist", "Information on installing VC++ Redistributables", command_vcredist, false },
    { "unlockstats", "Information on unlocking all items in the game", command_unlockstats, false },
    { "performance", "Information on improving the games performance", command_performance, false },
    { "fps", "Information on changing the FPS limit", command_fps, false },
    { "fov", "Information on changing your FOV", command_fov, false },
    { "nickname", "Information on changing your in-game name", command_nickname, false },
    { "console", "Information on the in-game command console", command_console, false },
    { "dxr", "Information on installing DirectX and VC++ Redistributables (Error 0xc000007b)", command_dxr, false },
    { "rawfiles", "Information on installing/repairing iw4x-rawfiles", command_rawfiles, false },
    { "game", "Information on supported copies of MW2", command_game, false },
    { "dlc", "Information on MW2 and IW4x DLC", command_dlc, false },
    { "stats", "Information on IW4x server statistics", command_stats, true },
}

// restart and uptime need the session/uptime so they get a special case in handle_interaction
var staff_subcommands = []subcommand {
    { "staffhelp", "Displays the staff help dialog", command_staffhelp, false },
    { "restart", "Sends the bot a signal to restart itself", nil, false },
    { "uptime", "Displays bot uptime", nil, false },
}

// registers all of our commands with discord globally
// bulk overwrite replaces whatever got registered before so any command removed
// also gets removed from discord (on like bot update or whatever)
func register_commands(s *discordgo.Session) (error) {
    _, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", build_commands())
    return err
}

// bleh
func build_commands() ([]*discordgo.ApplicationCommand) {
    guild_only := []discordgo.InteractionContextType { discordgo.InteractionContextGuild }
    top_level := []struct { name, description string; list []subcommand } {
        { command_name, "IW4x help and information", subcommands },
        { staff_command_name, "IW4x staff commands", staff_subcommands },
    }

    var commands []*discordgo.ApplicationCommand
    for _, t := range top_level {
        command := &discordgo.ApplicationCommand { Name: t.name, Description: t.description, Contexts: &guild_only }

        for _, c := range t.list {
            command.Options = append(command.Options, &discordgo.ApplicationCommandOption {
                Type: discordgo.ApplicationCommandOptionSubCommand, Name: c.name, Description: c.description,
            })
        }

        commands = append(commands, command)
    }

    return commands
}

// this is the entry for every interaction discord sends the bot
func handle_interaction(s *discordgo.Session, i *discordgo.InteractionCreate, uptime time.Time) {
    // commands are guild only so Member should always be set
    if i.Type != discordgo.InteractionApplicationCommand || i.Member == nil {
        return
    }

    data := i.ApplicationCommandData()
    if len(data.Options) != 1 {
        return
    }

    option := data.Options[0].Name
    user := "<" + i.Member.User.ID + ":" + i.Member.User.Username + ">"

    list := subcommands
    if data.Name == staff_command_name {
        if !check_permissions(i.Member) {
            log.Print("iw4x-discord-bot: non-staff user: " + user + " attempted staff command: " + option + "'")
            respond(s, i, &discordgo.InteractionResponseData {
                Content: "You do not have permission to use this command.",
                Flags: discordgo.MessageFlagsEphemeral, // only visible to the user who tried to run this command
            })
            return
        }
        list = staff_subcommands
    }

    // this should only fail with a stale command that discord has registered on their end
    index := slices.IndexFunc(list, func(c subcommand) bool { return c.name == option })
    if index < 0 {
        log.Print("iw4x-discord-bot: unknown command: '" + data.Name + " " + option + "' issued by user: " + user)
        return
    }
    command := list[index]

    log.Print("iw4x-discord-bot: command: '" + data.Name + " " + option + "' requested by user: " + user)
    command_timer := time.Now()
    var err error

    switch {
    case data.Name == staff_command_name && option == "restart":
        log.Print("iw4x-discord-bot: triggering restart...")
        respond (s, i, &discordgo.InteractionResponseData { Content: "gn" })
        s.Close()
        os.Exit(0)

    case data.Name == staff_command_name && option == "uptime":
        err = respond(s, i, &discordgo.InteractionResponseData { Content: "awake for " + time.Since(uptime).String() })

    case command.deferred:
        // "bot is thinking..."
        err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse { Type: discordgo.InteractionResponseDeferredChannelMessageWithSource })
        if err != nil {
            break
        }

        header, body := command.run()
        if body == "" {
            header, body = "Something went wrong!", "Failed to fetch information, please try again later."
        }

        embeds := build_embeds(header, body)
        _, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit { Embeds: &embeds })

    default:
        err = respond(s, i, &discordgo.InteractionResponseData { Embeds: build_embeds(command.run()) })
    }

    if err != nil {
        log.Print("iw4x-discord-bot: failed to send response to command: '" + data.Name + " " + option + "': ", err)
        return
    }

    log.Print("iw4x-discord-bot: response to command: '" + data.Name + " " + option + "' from user: " + user + " sent in: <", time.Since(command_timer), ">")
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.InteractionResponseData) (error) {
    return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse {
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: data,
    })
}

// builds the embed
func build_embeds(header string, body string) ([]*discordgo.MessageEmbed) {
    return []*discordgo.MessageEmbed { { Title: header, Description: body, Color: 0x0ff00 } }
}
