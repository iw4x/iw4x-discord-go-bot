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
    "flag"
)

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
    players, err := fetch_players()
    if err != nil {
        return err
    }

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

func create_send_query(s *discordgo.Session, m *discordgo.MessageCreate) (error) {
    file, err := os.Open("/tmp/queryresults.json")
    if err != nil {
        return err
    }
    defer file.Close()

    message := &discordgo.MessageSend{
        Content: "Query results:",
        Files: []*discordgo.File{
            {
                Name: "query_results.json",
                Reader: file,
            },
        },
    }

    _, err = s.ChannelMessageSendComplex(m.ChannelID, message)
    if err != nil {
        return err
    }

    return nil
}

// gets and returns amount of active players
func fetch_players() (string, error) {
    type Server struct {
        Client int `json:"clients"` // each `servers` entry contains a `clients` variable, pull that
    }

    type Response struct {
        Servers []Server `json:"servers"` // we're looking through entries in `servers`
    }

    r, err := http.Get("https://master." + base_url + "v1/servers/iw4x?protocol=152")
    if err != nil {
        return "0", err
    }
    defer r.Body.Close()

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        return "0", err
    }

    var response Response
    json.Unmarshal(body, &response)

    var result int = 0
    for _, p := range response.Servers {
        result += p.Client // for every entry, sum with current value of result
    }

    // this needs to be a string when used for status, convert from int
    result_output := strconv.Itoa(result)

    return result_output, nil
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

func get_logfile_length(location string) (int, error) {
    logfile, err := os.Open(filepath.Join(location, "chatlog.json"))
    if err != nil {
        return 0, err
    }
    defer logfile.Close() // close the file once this function returns 

    line_count := 0
    scanner := bufio.NewScanner(logfile)

    for scanner.Scan() {
        line_count++
    }
    if err := scanner.Err(); err != nil {
        return 0, err
    }
	
    return line_count, nil
}

func cycle_logfile(location string, log_archive_dir string) (error) {
    logfile, err := os.Open(filepath.Join(location, "chatlog.json"))
    if err != nil {
        return err
    }
    defer logfile.Close()

    now := time.Now()
    formatted_now := now.Format("06-01-02") // this will give us a date.gz backup in archive/
	
    archive_path := filepath.Join(log_archive_dir, formatted_now+".gz")
    destination, err := os.Create(archive_path)
    if err != nil {
        return err
    }
    defer destination.Close()

    gzip_writer, err := gzip.NewWriterLevel(destination, gzip.BestCompression) // https://pkg.go.dev/compress/flate#BestCompression
    if err != nil {
        return err
    }
    defer gzip_writer.Close()
	
    if _, err := io.Copy(gzip_writer, logfile); err != nil {
        return err
    }

    // truncate logfile to clear it out
    if err := os.Truncate(filepath.Join(location, "chatlog.json"), 0); err != nil {
        log.Print(err)
        return err
    }

    return nil
}

func query_db(location string, opts []string) ([]string, error) {
    type Database struct {
        MType string `json:"type"`
        Content string `json:"content"`
        MID string `json:"message_id"`
        CID string `json:"channel_id"`
        AID string `json:"author_id"`
        AUsername string `json:"author_username"`
        ANickname string `json:"author_nickname"`
    }

    // given these are discord messages, opt handling is a little bit odd
    // we need to set up a few variables for potential opts to assign their values to
    // any values given that aren't assigned here will simply be discarded
    var m_value string
    var c_value string
    var a_value string
    var u_value string
    var n_value string
    var d_value bool
    var e_value bool
    
    flags := flag.NewFlagSet("querydb", flag.ContinueOnError)

    flags.StringVar(&m_value, "m", "", "Message ID")
    flags.StringVar(&c_value, "c", "", "Channel ID")
    flags.StringVar(&a_value, "a", "", "Author ID")
    flags.StringVar(&u_value, "u", "", "Author Username")
    flags.StringVar(&n_value, "n", "", "Author Nickname")
    
    flags.BoolVar(&d_value, "d", false, "Deleted messages") // these just need to be toggled and do not take a value
    flags.BoolVar(&e_value, "e", false, "Edited messages")
    
    if err := flags.Parse(opts[:]); err != nil {
        return nil, err
    }
    
    file, err := os.Open(filepath.Join(location, "chatlog.json"))
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // this is the slice we will populate matching lines into
    var matching_db_entries []string

    // this is used to track uniqueness so we don't populate duplicates
    seen := make(map[string]bool)
    
    scanner := bufio.NewScanner(file)
    
    for scanner.Scan() {
        line := scanner.Text()
        keep := true
        
        var db Database
        if err := json.Unmarshal([]byte(line), &db); err != nil {
            log.Print("iw4x-discord-bot: failed to parse database entry: ", err)
            continue
        }

        if m_value != "" && m_value != db.MID { // if m_value is empty, this opt probably wasnt specified
            keep = false
        }

        if c_value != "" && c_value != db.CID {
            keep = false
        }

        if a_value != "" && a_value != db.AID {
            keep = false
        }

        if u_value != "" && u_value != db.AUsername {
            keep = false
        }

        if n_value != "" && n_value != db.ANickname {
            keep = false
        }
 
        if d_value && db.MType != "deletion" {
            keep = false
        }

        if e_value && db.MType != "edit" {
            keep = false
        }
        
        if keep && !seen[line] {
            matching_db_entries = append(matching_db_entries, line)
            seen[line] = true
        }
    }

    return matching_db_entries, nil
}
