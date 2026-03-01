package main

import (
    "strings"
    "strconv"
    "os"
    "log"
    "path/filepath"
)

// all of the functions here need to return a title and body of type string
// and main.go will call a function to construct a message and spit it out

func command_help() (string, string) {
    header := "Available Commands"

    // create a list of strings for each line of the body
    var output = []string{"- `!iw4x help` - Displays this help dialog",
    "- `!iw4x install` - Information about installing IW4x.",
    "- `!iw4x docs` - IW4x documentation",
    "- `!iw4x discord` - Displays the invite link for this server",
    "- `!iw4x github` - Displays the link to IW4x source code.",
    "- `!iw4x repair` - Information on repairing your game files",
    "- `!iw4x dedicated` - Information on setting up a dedicated server",
    "- `!iw4x vcredist` - Information on installing VC++ Redistributables",
    "- `!iw4x unlockstats` - Information on unlocking all items in the game",
    "- `!iw4x performance` - Information on improving the games performance",
    "- `!iw4x fps` - Information on changing the FPS limit",
    "- `!iw4x fov` - Information on changing your FOV",
    "- `!iw4x nickname` - Information on changing your in-game name",
    "- `!iw4x console` - Information on the in-game command console",
    "- `!iw4x dxr` - Information on installing DirectX and VC++ Redistributables (Error 0xc000007b)",
    "- `!iw4x rawfiles` - Information on installing/repairing iw4x-rawfiles",
    "- `!iw4x game` - Information on supported copies of MW2",
    "- `!iw4x dlc` - Information on MW2 and IW4x DLC",
    "",
    "If you would like more information about IW4x, visit the official documentation at https://docs." + base_url}

    // join `output` into one string to be passed back
    // \n is our field separator here
    body := strings.Join(output[:], "\n")

    return header, body
}

func command_install() (string, string) {
    header := "Installing/updating IW4x"

    var output = []string{"There is high quality documentation including images and GIFs available on the IW4x website.",
    "",
    "- Installing/updating via launcher **(Recommended):**",
    "  - https://docs." + base_url + "get-started/quickstart",
    "",
    "- Manual:",
    "  - Windows: https://docs." + base_url + "get-started/manual-install/windows-guide/",
    "  - Linux: https://docs." + base_url + "get-started/manual-install/linux-guide/",
    "  - MacOS: https://docs." + base_url + "get-started/manual-install/macos-guide/"}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_docs() (string, string) {
    header := "Documentation"

    var output = []string{"IW4x documentation, tutorials, and various other information can be found at:",
    "- https://docs." + base_url}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_discord() (string, string) {
    header := "Discord server invite"

    var output = []string{"- https://discord.com/invite/pV2qJscTXf",
    "- https://" + base_url + "discord"}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_github() (string, string) {
    header := "Source code"

    var output = []string{"The source code for IW4x can be found at:",
    "- https://github.com/iw4x/"}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_repair() (string, string) {
    header := "Repairing your game files"

    var output = []string{"If you are having issues running IW4x, you may try repairing your game files.",
    "",
    "1. Steam:",
    "    - Right click on **Call of Duty: Modern Warfare 2** in your Steam library.",
    "    - Left click on **Properties**.",
    "    - Navigate to the **Installed Files** tab.",
    "    - Left click on **Verify integrity of game files**.",
    "    - Wait for this process to finish.",
    "",
    "2. IW4x:",
    "    - Right click on **Call of Duty: Modern Warfare 2** in your Steam library.",
    "    - Hover your cursor over **Manage**.",
    "    - Left click on **Browse local files.**",
    "    - Replace IW4x rawfiles by re-extracting release.zip (available at https://github.com/iw4x/iw4x-rawfiles/releases) in the same directory and replacing all `.iwd` files in the `iw4x` folder with fresh ones from the same website.",
    "    - Replace iw4x.dll with a new copy, available at: https://github.com/iw4x/iw4x-client/releases",
    "",
    "If this did not solve your issue, please describe the problem in the <#1420088697960796170> channel."}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_dedicated() (string, string) {
    header := "Setting up a dedicated server"

    var output = []string{"There are detailed instructions at:",
    "- https://docs." + base_url + "hosting/server-hosting/"}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_vcredist() (string, string) {
    header := "Installing VC++ Redistributable 2005"

    var output = []string{"The VC++ Redistributable 2005 installer is in the `Redist` folder in your game directory already, called `vcredist_x86.exe`.",
    "",
    "You may also download **the x86 version** from Microsoft at: https://www.microsoft.com/en-us/download/details.aspx?id=26347",
    "",
    "Regardless of which method you choose, run the installer and follow the instructions.",
    "",
    "If you are still having problems after installing it, please restart your computer.",
    "",
    "If this did not solve your issue, please describe the problem in the <#1420088697960796170> channel.",
    "",
    "If you get errors related to d3d9.dll or xinput1_3.dll, see the `!iw4x directx` command for information."}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_unlockstats() (string, string) {
    header := "Unlock all in-game items"

    var output = []string{"1. On the main menu, left click on **Barracks** and then again on **Unlock stats**.",
    "    - You should now be max prestige and rank 70 with everything unlocked.",
    "",
    "If you wish, you may use the console command `unlockstats` to perform the same action.",
    "",
    "Do check out `!iw4x console` or the [console guide](https://" + base_url + "guides/console/) for more information on using the in-game console.",
    "",
    "https://" + base_url + "guides/unlockstats/"}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_performance() (string, string) {
    header := "Improving game performance"

    var output = []string{"There is a performance improvement guide available at:",
    "- https://docs." + base_url + "guides/performance"}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_fps() (string, string) {
    header := "Changing your FPS limit"

    var output = []string{"You may change the maximum FPS your game runs at in the **Settings** menu or by using the in-game console.",
    "",
    "If you would like to use the console, the command to adjust the limit is `com_maxfps <value>`, where `<value>` is your chosen FPS limit.",
    "    - If you would like to entirely remove the FPS limit, set `<value>` to `0`.",
    "    - It isn't recommended to exceed a limit of 333fps, as the game engine will behave erratically the higher your FPS gets.",
    "",
    "Do check out `!iw4x console` or the [console guide](https://" + base_url + "guides/console/) for more information on using the in-game console."}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_fov() (string, string) {
    header := "Changing your in-game Field of View (FOV)"

    var output = []string{"You can change your FOV using the console command `cg_fov <value>`, where `<value>` is your chosen FOV. The default setting is `65`.",
    "",
    "Do check out `!iw4x console` or the [console guide](https://" + base_url + "guides/console/) for more information on using the in-game console."}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_nickname() (string, string) {
    header := "Changing your in-game name"

    var output = []string{"1. On the main menu, left click **Barracks** and again on your current name in the top right, which may be Unknown Soldier.",
    "    - You can move your cursor with your arrow keys, or just start typing to overwrite the existing text.",
    "",
    "2. Press Enter to confirm your new name.",
    "",
    "If you wish, you may use the console command `name <mynewname>`, where `<mynewname>` is your chosen name.",
    "",
    "Do check out `!iw4x console` or the [console guide](https://" + base_url + "guides/console/) for more information on using the in-game console.",
    "",
    "https://" + base_url + "guides/namechange/"}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_console() (string, string) {
    header := "Using the in-game console"

    var output = []string{"Depending on what your keyboard layout is, the console can be opened using either:",
    "- The **tilde:** ~",
    "- The **grave accent:** `",
    "- Or the **caret:** ^ key.",
    "",
    "Regardless, the key should be located immediately under the **esc** key.",
    "",
    "For more details and console commands, do see the [console guide](https://" + base_url + "guides/console/)."}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_dxr() (string, string) {
    header := "Installing DirectX/Redist, error 0xc000007b"

    var output = []string{"Errors referencing `d3d9.dll` and/or `xinput1_3.dll` are a result of a missing or broken DirectX installation.",
    "",
    "1. Make sure you haven't downloaded any of the dll files from the internet, if you have, **please delete them**.",
    "2. Download DirectX from Microsoft [here](https://www.microsoft.com/en-us/download/details.aspx?id=35).",
    "3. Run `dxwebsetup.exe` by double-clicking it, and follow the instructions.",
    "4. Try running IW4x again. If you are still facing problems, please restart your computer.",
    "",
    "If you are experiencing error `0xc000007b`, be sure to install both DirectX from the link above, as well as the VC++ Redistributable 2005 from Microsoft [here](https://www.microsoft.com/en-us/download/details.aspx?id=26347).",
    "",
    "If this did not solve your issue, please describe the problem in the <#1420088697960796170> channel."}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_rawfiles() (string, string) {
    header := "Installing/repairing rawfiles"

    var output = []string{"Errors referencing missing files or, for example, `Couldn't load image 'button_a'` are a result of a broken/missing rawfiles installation.",
    "",
    "To fix this, download `release.zip` and `.iwd` files from https://github.com/iw4x/iw4x-rawfiles/releases/latest",
    "",
    "1. Extract `release.zip` into the root directory of your game folder, and **not** into a directory called `release` or otherwise.",
    "2. Copy all `.iwd` files to the `iw4x` directory in the root of your game folder.",
    "",
    "If you are replacing an older rawfiles installation, replace all conflicting files with the new ones.",
    "",
    "For more information, see the `!iw4x install` command."}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_game() (string, string) {
    header := "Obtaining a copy of Call of Duty Modern Warfare 2"

    var output = []string{"The **only** supported copy of MW2 is from [Steam](https://store.steampowered.com/app/10180/Call_of_Duty_Modern_Warfare_2_2009/). The Microsoft Store version of MW2 will not work.",
    "",
    "Support will not be provided for non-Steam copies of the game.",
    ""}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_dlc() (string, string) {
    header := "MW2/IW4x DLC"

    var output = []string{"There are two different DLCs- the official Modern Warfare 2 DLC packs and the IW4x DLCs.",
    "",
    "In the case of the MW2 DLC packs, Resurgence and Stimulus, they can be purchased from Steam:",
    "- [Resurgence](https://store.steampowered.com/app/10196/Call_of_Duty_Modern_Warfare_2_Resurgence_Pack/)",
    "- [Stimulus](https://store.steampowered.com/app/10195/Call_of_Duty_Modern_Warfare_2_Stimulus_Package/)",
    "",
    "In the case of IW4x DLCs, they will be installed by the launcher when you install IW4x."}

    body := strings.Join(output[:], "\n")

    return header, body
}

// STAFF COMMANDS BELOW THIS POINT

func command_staffhelp() (string, string) {
    header := "Available Staff Commands"

    var output = []string{"- `!iw4x staffhelp` - Displays this help dialog",
    "- `!iw4x restart` - Sends the bot a signal to restart itself",
    "- `!iw4x querydb -m <messageid> -c <channelid> -a <authorid> -u <authorusername> -n <authornickname> -d -e -t` - Query the message log database",
    "    - This does not require all options, but requires at least one. In the case of `-d`, `-e`, and `-t`, this will filter the output to deleted, edited, and messages with attachments only, respectively.",
    "- `!iw4x logstat` - Displays statistics about the message log"}

    body := strings.Join(output[:], "\n")

    return header, body
}

func command_logstat(message_count int, location string) (string, string) {
    header := "Log Statistics"

    // convert to string for reply
    count_output := strconv.Itoa(message_count)
    
    logfile_stat, err := os.Stat(filepath.Join(location, "chatlog.json"))
    if err != nil {
        log.Print("iw4x-discord-bot: failed to stat active logfile: ", err)
        return "", ""
    }
    logfile_size := logfile_stat.Size()
    logfile_size_kilobytes := float64(logfile_size) / 1024
    logfile_size_output := strconv.FormatFloat(logfile_size_kilobytes, 'f', 2, 64) 
    
    var output = []string{"Active entries: "+ count_output,
    "Size of active logfile: " + logfile_size_output + "KB"}

    body := strings.Join(output[:], "\n")

    return header, body
}
