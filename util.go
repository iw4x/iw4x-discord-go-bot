package main

// this is a "utility" file for various internal functions
// if you're looking for the bot commands, check commands.go

import (
    "github.com/bwmarrin/discordgo"

    "sync"
    "time"
    "io"
    "os"
    "bufio"
    "net/http"
    "encoding/json"
    "log"
    "strconv"
    "slices"
    "compress/gzip"
    "path/filepath"
    "flag"
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

// this gives us an io.writer that allows us to replace the underlying destination at runtime
// this is needed so the slog handler can keep writing while cycle_logfile redirects
// log entries to a new file without losing messages mid-cycle
type swappableWriter struct {
    mu sync.Mutex
    w io.Writer
    count int64
}

// this satisfies io.Writer, lock is only held for the duration of a single write to the log
func (s *swappableWriter) Write(p []byte) (int, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    n, err := s.w.Write(p)
    if err == nil {
        s.count++
    }
    return n, err
}

// replaces the underlying writer and returns the previous one
// so it can be closed cleanly after archiving
func (s *swappableWriter) Swap(new io.Writer) (io.Writer, int64) {
    s.mu.Lock()
    defer s.mu.Unlock()
    old, oldCount := s.w, s.count
    s.w, s.count = new, 0
    return old, oldCount
}

// count returns the number of successfully writes since the last swap
func (s *swappableWriter) Count() int64 {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.count
}

// seed the count based on the existing logfiles line count
func (s *swappableWriter) SetCount(c int64) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.count = c
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
    stats, err := fetch_master_stats()
    var players string
    if err != nil {
        // this doesn't return so the bot can apply its "Currently sleeping.." status
        log.Print("iw4x-discord-bot: failed to fetch player count: ", err)
        players = "0"
    } else {
        players = strconv.Itoa(stats.Players)
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
    r, err := http.Get("https://master." + base_url + "v1/stats?protocol=152")
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

func cycle_logfile(location string, log_archive_dir string, swappable *swappableWriter) (error) {
    active_path := filepath.Join(location, "chatlog.json")
    cycling_path := filepath.Join(location, "chatlog.json.cycling")

    // rename the active log, existing writers keep writing against the same inode
    if err := os.Rename(active_path, cycling_path); err != nil {
        return err
    }

    // open a fresh active log
    new_file, err := os.OpenFile(active_path, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
    if err != nil {
        return err
    }

    // atomically redirect the logger to the new file
    old_writer, _ := swappable.Swap(new_file)

    // close the old file
    if old_file, ok := old_writer.(*os.File); ok {
        if err := old_file.Close(); err != nil {
            log.Print("iw4x-discord-bot: failed to close old logfile during cycle: ", err)
        }
    }

    // archive the cycled out logfile, open for reading separately
    cycling_file, err := os.Open(cycling_path)
    if err != nil {
        return err
    }
    defer cycling_file.Close()

    formatted_now := time.Now().Format("06-01-02") // date.gz in archive/
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

    if _, err := io.Copy(gzip_writer, cycling_file); err != nil {
        return err
    }

    // remove cycle file
    cycling_file.Close()
    if err := os.Remove(cycling_path); err != nil {
        log.Print("iw4x-discord-bot: failed to remove cycle logfile: ", err)
    }

    return err
}

func query_db(location string, opts []string, invoking_message_id string) ([]string, error) {
    type Attachment struct {}

    type Database struct {
        MType string `json:"type"`
        Content string `json:"content"`
        MID string `json:"message_id"`
        CID string `json:"channel_id"`
        AID string `json:"author_id"`
        AUsername string `json:"author_username"`
        ANickname string `json:"author_nickname"`
        Attachments []Attachment `json:"attachments"`
    }

    // given these are discord messages, opt handling is a little bit odd
    // we need to set up a few variables for potential opts to assign their values to
    // any values given that aren't assigned here will simply be discarded
    var m_value string
    var c_value string
    var a_value string
    var u_value string
    var n_value string
    var s_value string
    var d_value bool
    var e_value bool
    var t_value bool

    flags := flag.NewFlagSet("querydb", flag.ContinueOnError)
    flags.SetOutput(io.Discard) // don't print usage information to the serverside log

    flags.StringVar(&m_value, "m", "", "Message ID")
    flags.StringVar(&c_value, "c", "", "Channel ID")
    flags.StringVar(&a_value, "a", "", "Author ID")
    flags.StringVar(&u_value, "u", "", "Author Username")
    flags.StringVar(&n_value, "n", "", "Author Nickname")
    flags.StringVar(&s_value, "s", "", "Message Content")

    flags.BoolVar(&d_value, "d", false, "Deleted messages") // these just need to be toggled and do not take a value
    flags.BoolVar(&e_value, "e", false, "Edited messages")
    flags.BoolVar(&t_value, "t", false, "Attachment messages")
    
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

        // !iw4x querydb shouldnt match its own filters, but it's still logged for *later* queries
        if db.MID == invoking_message_id {
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

        // make the query case-insensitive
        if s_value != "" && !strings.Contains(strings.ToLower(db.Content), strings.ToLower(s_value)) {
            keep=false
        }

        if d_value && db.MType != "deletion" {
            keep = false
        }

        if e_value && db.MType != "edit" {
            keep = false
        }

        if t_value && len(db.Attachments) == 0 {
            keep = false
        }
        
        if keep && !seen[line] {
            matching_db_entries = append(matching_db_entries, line)
            seen[line] = true
        }
    }

    return matching_db_entries, nil
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
    return slices.Contains([]string{"restart", "staffhelp", "querydb", "logstat", "uptime"}, opt)
}

// this lets us combine user commands into tokens, anything wrapped in double quores is kept intact
func tokenize(s string) ([]string, error) {
    var tokens []string // list of tokens to be passed back
    var current strings.Builder // temporary storage for building tokens
    in_quotes := false
    has_token := false

    for i := 0; i < len(s); i++ {
        c := s[i]

        if in_quotes {
            if c == '"' {
                in_quotes = false
                continue
            }
            current.WriteByte(c)
            continue
        }

        // outside of a quoted region
        if c == '"' {
            in_quotes = true
            has_token = true // opening a quote starts a token even if empty inside
            continue
        } else if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
            if has_token {
                tokens = append(tokens, current.String())
                current.Reset()
                has_token = false
            }
            continue
        }
        current.WriteByte(c)
        has_token = true
    }

    if in_quotes {
        return nil, fmt.Errorf("unclosed double quote")
    }

    if has_token {
        tokens = append(tokens, current.String())
    }

    return tokens, nil
}
