package main

// this file contains everything for the honeypot channel persistent
// counter embed, the embed lives at the bottom of the honeypot channel and is
// updated as bans happen

import (
    "github.com/bwmarrin/discordgo"

    "encoding/json"
    "log"
    "os"
    "path/filepath"
    "strconv"
    "time"
)

// the color of the embed sidebar thing
// this is just normal iw4x green like the rest of the embedded messages
const honeypot_colour int = 0x0ff00

// bans are coalesced for this long before the embed is updated, during a raid
// this turns a few dozen edits into one and keeps us well clear of the
// per-channel edit ratelimit
const honeypot_flush_interval time.Duration = 5 * time.Second

// how long a banned user ID is remembered for deduplication purposes,
// a spam bot maybe could get several messages out before the ban lands and each of
// those would otherwise count as a separate ban
const honeypot_dedupe_ttl time.Duration = 1 * time.Hour

// where the counter is persisted so it survives restarts, this sits next to
// chatlog.json in the directory the bot is run from
const honeypot_state_file string = "honeypot.json"

// on-disk representation of the counter, the message ID is stored so a
// restart edits the existing embed rather than posting a duplicate
type honeypot_state struct {
    Count int64 `json:"count"`
    MessageID string `json:"message_id"`
    LastBan time.Time `json:"last_ban"`
}

func load_honeypot_state(location string) (honeypot_state, error) {
    var state honeypot_state

    contents, err := os.ReadFile(filepath.Join(location, honeypot_state_file))
    if err != nil {
        if os.IsNotExist(err) { // first run, start from zero
            return state, nil
        }
        return state, err
    }

    if err := json.Unmarshal(contents, &state); err != nil {
        return honeypot_state{}, err
    }

    return state, nil
}

// written to a temporary file and renamed so a crash mid-write can't leave
// a truncated state file behind
func save_honeypot_state(location string, state honeypot_state) (error) {
    contents, err := json.Marshal(state)
    if err != nil {
        return err
    }

    state_path := filepath.Join(location, honeypot_state_file)
    temp_path := state_path + ".tmp"

    if err := os.WriteFile(temp_path, contents, 0644); err != nil {
        return err
    }

    return os.Rename(temp_path, state_path)
}

// builds the counter embed
func build_honeypot_embed(state honeypot_state) (*discordgo.MessageEmbed) {
    var body = []string {
        "This channel is used to catch spam bots.",
        "Any messages sent here will result in **an immediate ban**.",
    }

    embed := &discordgo.MessageEmbed {
        Title: "DO NOT SEND MESSAGES IN THIS CHANNEL",
        Description: body[0] + "\n" + body[1],
        Color: honeypot_colour,
        Fields: []*discordgo.MessageEmbedField {
            {
                Name: "Bots caught",
                Value: "```\n" + strconv.FormatInt(state.Count, 10) + "\n```",
                Inline: true,
            },
        },
    }

    // discord renders this as a relative timestamp next to the footer text,
    // omitted entirely until the first ban
    if !state.LastBan.IsZero() {
        embed.Footer = &discordgo.MessageEmbedFooter { Text: "Most recent ban" }
        embed.Timestamp = state.LastBan.Format(time.RFC3339)
    }

    return embed
}

// edits the existing embed if it is still the newest message in the channel,
// otherwise deletes it and reposts so the counter stays at the bottom (this shouldn't happen but still)
func refresh_honeypot_message(s *discordgo.Session, state *honeypot_state, location string) (error) {
    embeds := []*discordgo.MessageEmbed { build_honeypot_embed(*state) }

    if state.MessageID != "" {
        newest, err := s.ChannelMessages(honeypot_channel, 1, "", "", "")
        if err != nil {
            log.Print("iw4x-discord-bot: failed to fetch newest honeypot message: ", err)
        }

        // still at the bottom of the channel, a plain edit is enough
        if err == nil && len(newest) == 1 && newest[0].ID == state.MessageID {
            _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit {
                Channel: honeypot_channel,
                ID: state.MessageID,
                Embeds: &embeds,
            })
            if err == nil {
                return nil
            }
            log.Print("iw4x-discord-bot: failed to edit honeypot counter, reposting: ", err)
        }

        // something else is below us, or the message is gone entirely
        if err := s.ChannelMessageDelete(honeypot_channel, state.MessageID); err != nil {
            log.Print("iw4x-discord-bot: failed to delete stale honeypot counter: ", err)
        }
    }

    message, err := s.ChannelMessageSendEmbed(honeypot_channel, embeds[0])
    if err != nil {
        return err
    }

    state.MessageID = message.ID

    return save_honeypot_state(location, *state)
}

// owns the counter, this is the only thing that touches the count or the
// message so no locking is needed, the message handler only feeds it user IDs
// this never returns and is expected to run for the lifetime of the bot
func run_honeypot_counter(s *discordgo.Session, location string, bans <-chan string) {
    state, err := load_honeypot_state(location)
    if err != nil {
        log.Print("iw4x-discord-bot: failed to load honeypot state: ", err)
    }

    // recently banned IDs, used to avoid counting the same spam bot twice
    seen := make(map[string]time.Time)

    // post or adopt the counter embed immediately on startup
    dirty := false
    if err := refresh_honeypot_message(s, &state, location); err != nil {
        log.Print("iw4x-discord-bot: failed to create honeypot counter: ", err)
        dirty = true // retry on the next tick
    }

    flush := time.NewTicker(honeypot_flush_interval)
    defer flush.Stop()

    for {
        select {
        case user_id := <-bans:
            now := time.Now()

            for id, banned_at := range seen {
                if now.Sub(banned_at) > honeypot_dedupe_ttl {
                    delete(seen, id)
                }
            }

            if _, duplicate := seen[user_id]; duplicate {
                continue
            }
            seen[user_id] = now

            state.Count++
            state.LastBan = now
            dirty = true

            // persisted on every ban rather than on every flush so an
            // unclean shutdown can't lose bans
            if err := save_honeypot_state(location, state); err != nil {
                log.Print("iw4x-discord-bot: failed to save honeypot state: ", err)
            }

        case <-flush.C:
            if !dirty {
                continue
            }

            if err := refresh_honeypot_message(s, &state, location); err != nil {
                log.Print("iw4x-discord-bot: failed to update honeypot counter: ", err)
                continue // stays dirty, retried next tick
            }
            dirty = false
        }
    }
}
