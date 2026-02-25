package main

// this is a "utility" file for various internal functions
// if you're looking for the bot commands, check commands.go

import (
    "github.com/bwmarrin/discordgo"

    "time"
    "io"
    "os"
    "bufio"
    "net/http"
    "encoding/json"
    "io/ioutil"
    "log"
    "strconv"
    "slices"
    "compress/gzip"
    "path/filepath"
)

// builds embeds and sends output for all commands
// header and body are passed into this from the function map call below,
// map call fetches this information from each commands function in commands.go
func create_send_response(header string, body string, s *discordgo.Session, m *discordgo.MessageCreate) {
    embed := &discordgo.MessageEmbed {
        Title: header,
        Description: body,
        Color: 0x0ff00,
    }

    s.ChannelMessageSendEmbed(m.ChannelID, embed)

    return
}

// builds and sends output for player count in status
func create_send_status(s *discordgo.Session) (bool) {
    players := fetch_players()

    if players != "0" {
        err := s.UpdateStatusComplex(discordgo.UpdateStatusData { // https://pkg.go.dev/github.com/bwmarrin/discordgo#UpdateStatusData
            Status: "online", // try to prevent the bot from getting sleepy
            Activities: []*discordgo.Activity { // https://pkg.go.dev/github.com/bwmarrin/discordgo#Activity
                {
                    Type: 4, // https://pkg.go.dev/github.com/bwmarrin/discordgo#ActivityType
                    Name: "Custom Status", // i have no idea why this won't work without this but sure
                    State: "Current players: " + players,
                },
            },
        })

        if err != nil {
            log.Print(err)
            return false
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
            log.Print(err)
            return false
        }
    }

    return true
}


// gets and returns amount of active players
func fetch_players() (string) {
    type Server struct {
        Client int `json:"clients"` // each `servers` entry contains a `clients` variable, pull that
    }

    type Response struct {
        Servers []Server `json:"servers"` // we're looking through entries in `servers`
    }

    r, err := http.Get("https://master." + base_url + "v1/servers/iw4x?protocol=152")
    if err != nil {
        log.Print(err)
        return "0"
    }
    defer r.Body.Close()

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Print(err)
        return "0"
    }

    var response Response
    json.Unmarshal(body, &response)

    var result int = 0
    for _, p := range response.Servers {
        result += p.Client // for every entry, sum with current value of result
    }

    // this needs to be a string when used for status, convert from int
    result_output := strconv.Itoa(result)

    return result_output
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

func get_logfile_length(location string) (int) {
    logfile, err := os.Open(filepath.Join(location, "iw4xchat.log"))
    if err != nil {
        log.Print("iw4x-discord-bot: failed to read logfile size: ", err)
    }
    defer logfile.Close() // close the file once this function returns 

    line_count := 0
    scanner := bufio.NewScanner(logfile)

    for scanner.Scan() {
        line_count++
    }
    if err := scanner.Err(); err != nil {
        log.Print("iw4x-discord-bot: failed to read logfile size: ", err)
    }
	
    return line_count
}

func cycle_logfile(location string, log_archive_dir string) (bool) {
    logfile, err := os.Open(filepath.Join(location, "iw4xchat.log"))
    if err != nil {
        log.Print(err)
        return false
    }
    defer logfile.Close()

    now := time.Now()
    formatted_now := now.Format("06-01-02") // this will give us a date.gz backup in archive/
	
    archive_path := filepath.Join(log_archive_dir, formatted_now+".gz")
    destination, err := os.Create(archive_path)
    if err != nil {
        log.Print(err)
        return false
    }
    defer destination.Close()

    gzip_writer, err := gzip.NewWriterLevel(destination, gzip.BestCompression) // https://pkg.go.dev/compress/flate#BestCompression
    if err != nil {
        log.Print(err)
        return false
    }
    defer gzip_writer.Close()
	
    if _, err := io.Copy(gzip_writer, logfile); err != nil {
        log.Print(err)
        return false
    }

    // truncate logfile to clear it out
    if err := os.Truncate(filepath.Join(location, "iw4xchat.log"), 0); err != nil {
        log.Print(err)
        return false
    }

    return true
}
