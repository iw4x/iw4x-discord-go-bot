package main

// this is a "utility" file for various internal functions
// if you're looking for the bot commands, check commands.go

import (
    "github.com/bwmarrin/discordgo"

    "time"
    "io"
    "net/http"
    "encoding/json"
    "log"
    "strconv"
    "slices"
    "strings"
    "fmt"
)

// the information the stats portion of the master server can return, at the top of the file for easy modifications
type MasterStats struct {
    Players  int `json:"players"`
    Servers  int `json:"servers"`
    Bots     int `json:"bots"`
    Capacity int `json:"slots"`
}

// the player counts the x64 platform returns
type X64Population struct {
    Online    int            `json:"online"`
    Playing   int            `json:"playing"`
    Playlists map[string]int `json:"playlists"`
}

// builds embeds and sends output for all commands
// header and body are passed into this from the function map call below,
// map call fetches this information from each commands function in commands.go
func create_send_response(header string, body string, s *discordgo.Session, m *discordgo.MessageCreate) (error) {
    embed := &discordgo.MessageEmbed {
        Title: header,
        Description: body,
        Color: 0x0ff00,
    }

    _, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
    if err != nil {
        return err
    }

    return nil
}

// builds and sends output for player count in status
func create_send_status(s *discordgo.Session) (error) {
    // neither of these return so the bot can apply its "Currently sleeping.." status
    var players, x64_players int
    stats, err := fetch_master_stats()
    if err != nil {
        log.Print("iw4x-discord-bot: failed to fetch player count: ", err)
    } else {
        players = stats.Players
    }

    x64, err := fetch_x64_population()
    if err != nil {
        log.Print("iw4x-discord-bot: failed to fetch x64 player count: ", err)
    } else {
        x64_players = x64.Online
    }

    if players != 0 || x64_players != 0 {
        err := s.UpdateStatusComplex(discordgo.UpdateStatusData { // https://pkg.go.dev/github.com/bwmarrin/discordgo#UpdateStatusData
            Status: "online", // try to prevent the bot from getting sleepy
            Activities: []*discordgo.Activity { // https://pkg.go.dev/github.com/bwmarrin/discordgo#Activity
                {
                    Type: 4, // https://pkg.go.dev/github.com/bwmarrin/discordgo#ActivityType
                    Name: "Custom Status", // i have no idea why this won't work without this but sure
                    State: strconv.Itoa(players) + " playing, " + strconv.Itoa(x64_players) +" on MM",
                },
            },
        })

        if err != nil {
            return err
        }
    } else {
        err := s.UpdateStatusComplex(discordgo.UpdateStatusData {
            Status: "idle",
            Activities: []*discordgo.Activity {
                {
                    Type: 4,
                    Name: "Custom Status",
                    State: "Currently sleeping..",
                },
            },
        })

        if err != nil {
            return err
        }
    }

    return nil
}

func fetch_sale() (string, error) {
    type steam_sale map[string]struct {
        Data struct {
            PriceOverview struct {
                DiscountPercent int `json:"discount_percent"`
            } `json:"price_overview"`
        } `json:"data"`
    }

    r, err := http.Get("https://store.steampowered.com/api/appdetails?appids=10180&filters=price_overview")
    if err != nil {
        return "0", err
    }
    defer r.Body.Close()

    body, err := io.ReadAll(r.Body)
    if err != nil {
        return "0", err
    }

    var result steam_sale
    if err := json.Unmarshal(body, &result); err != nil {
        return "0", err
    }

    sale_percentage := result["10180"].Data.PriceOverview.DiscountPercent
    sale_output := strconv.Itoa(sale_percentage)

    return sale_output, nil
}

// this function can pull various information about iw4x from the master
func fetch_master_stats() (MasterStats, error) {
    var response MasterStats
    r, err := http.Get("https://master." + base_url + "v1/stats?protocol=153")
    if err != nil {
        return MasterStats{}, err
    }
    defer r.Body.Close()

    body, err := io.ReadAll(r.Body)
    if err != nil {
        return MasterStats{}, err
    }

    if r.StatusCode != http.StatusOK {
        return MasterStats{}, fmt.Errorf("%s", r.Status)
    }

    if err := json.Unmarshal(body, &response); err != nil {
        return MasterStats{}, err
    }

    return response, nil
}

// pulls the player counts of the x64 platform
func fetch_x64_population() (X64Population, error) {
    var response X64Population
    client := http.Client{Timeout: 10 * time.Second}
    r, err := client.Get(x64_population_url)
    if err != nil {
        return X64Population{}, err
    }
    defer r.Body.Close()

    body, err := io.ReadAll(r.Body)
    if err != nil {
        return X64Population{}, err
    }

    if r.StatusCode != http.StatusOK {
        return X64Population{}, fmt.Errorf("%s", r.Status)
    }

    if err := json.Unmarshal(body, &response); err != nil {
        return X64Population{}, err
    }

    return response, nil
}

// this is explicitly for staff only commands, and checks whether or not
// the command issuer has the 'staff' role or not- if not, it will return 1.
func check_permissions(m *discordgo.MessageCreate) (bool) {
    // https://pkg.go.dev/github.com/bwmarrin/discordgo#Member
    if slices.Contains(m.Member.Roles, staff_role_id) {
        return true
    } else {
        return false
    }
}

func send_join_message(s *discordgo.Session, joiner_id string) (error) {
    var output = []string{
        "Welcome to the IW4x server, <@"+joiner_id+">!",
        "",
        "To help you get up and running quickly, please check out our quick links and mini-FAQ below.",
        "",
        "### :warning: Important: No Piracy Supported",
        "Please note that **we do not support piracy under any circumstances**. You must own at least the base game on Steam to play IW4x and to receive support in this server.",
        "",
        "### Quick Links & Mini-FAQ",
        "* **How do I install IW4x?**",
        "    Check out our Quickstart guide to get everything installed and ready to go: [get-started/quickstart](<https://docs.iw4x.io/get-started/quickstart/>)",
        "* **How do I play a private match with friends?**",
        "    To play privately with your friends, you will need to set up a private server: [hosting/server-hosting](<https://docs.iw4x.io/hosting/server-hosting/>)",
        "* **How do I add bots to my game?**",
        "    You can play with bots using the Bot Warfare mod: [guides/bot-warfare](<https://docs.iw4x.io/guides/bot-warfare/>)",
        "* **How do I change the game language?**",
        "    Follow these steps to change your localization settings: [guides/change-language](<https://docs.iw4x.io/guides/change-language/>)",
        "* **How can I improve my game's performance?**",
        "    Experiencing stuttering or low FPS? Check out our optimization tips: [guides/performance](<https://docs.iw4x.io/guides/performance/>)",
        "",
        "### Still need help?",
        "* Type `!iw4x help` for a list of helpful bot commands.",
        "* If your question isn't answered above, feel free to ask in <#1111982470045368361>.",
        "* If you are reporting a bug or have a more complex issue, please create a thread in <#1420088697960796170>.",
        "",
        "*Please be patient, avoid cross-posting in multiple channels, and someone will help you out as soon as they can!*",
    }

    welcome_message := strings.Join(output[:], "\n")

    _, err := s.ChannelMessageSend("1114942926926127154", welcome_message)
    if err != nil {
        return err
    }

    return nil
}

func is_staff_command(opt string) (bool) {
    return slices.Contains([]string{"restart", "staffhelp", "uptime"}, opt)
}
