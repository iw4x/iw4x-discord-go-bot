package main

import "strings"

// all of the functions here need to return a title and body of type string
// and main.go will construct a message and spit it out
//
// the iw4x domain is contained in a variable here to make it easier
// to change in the future, if there are any more "events"
const base_url string = "iw4x.io/"

func command_help() (string, string) {
	header := "Available Commands"

	// create a list of strings for each line of the body
	var output = []string{"- `!iw4x help` - Displays this help dialog",
	"- `!iw4x install` - Information about installing IW4x.",
	"- `!iw4x docs` - IW4x documentation",
	"- `!iw4x discord` - Displays the invite link for this server",
	"- `!iw4x github` - Displays the link to IW4x source code.",
	"- `!iw4x redist` - Information on installing DirectX and VC++ Redistributables (for error 0xc000007b)",
	"- `!iw4x repair` - Information on repairing your game files",
	"- `!iw4x dedicated` - Information on setting up a dedicated server",
	"- `!iw4x vcredist` - Information on installing VC++ Redistributables",
	"- `!iw4x unlockstats` - Information on unlocking all items in the game",
	"- `!iw4x performance` - Information on improving the games performance",
	"- `!iw4x fps` - Information on changing the FPS limit",
	"- `!iw4x fov` - Information on changing your FOV",
	"- `!iw4x nickname` - Information on changing your in-game name",
	"- `!iw4x console` - Information on the in-game command console",
	"- `!iw4x directx` - Information on installing DirectX",
	"",
	"If you would like more information about IW4x, visit the official documentation at https://docs." + base_url}

	// join `output` into one string to be passed back
	// \n is our field separator here.
	body := strings.Join(output[:], "\n")

	return header, body
}

func command_install() (string, string) {
	header := "Installing/updating IW4x"

	var output = []string{"Currently, installing and updating IW4x should be done manually.",
	"",
	"- Installing/updating",
	"  1. Download the current release of `iw4x.dll` from: https://github.com/iw4x/iw4x-client/releases/latest",
	"  2. Download the current release of `release.zip` **and** `iw4x.exe` from: https://github.com/iw4x/iw4x-rawfiles/releases/latest",
	"  3. On your Steam client, right-click on **Call of Duty: Modern Warfare 2** in your Steam library.",
	"  4. Hover your cursor over **Manage**.",
	"  5. Left-click on **Browse local files**, this will open your game folder.",
	"  6. Move the `iw4x.dll` file downloaded in step 1 to this folder.",
	"    - If you are updating from an old release of the client, replace the old `iw4x.dll` in this folder with the new one",
	"  7. Move the `iw4x.exe` file downloaded in step 2 to this folder.",
	"  8. Unzip/extract the `release.zip` downloaded in step 2 into your game folder- the same location as `iw4x.dll`.",
	"    - If you are updating from an old release of the rawfiles, replace all conflicting files in this folder with those from the new `release.zip`.",
	"",
	"- Launching IW4x",
	"  - Windows: Double left-click on `iw4x.exe` in your games directory.",
	"  - Linux: You may add `iw4x.exe` as a non-Steam game in your Steam client, and run it with proton."}

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

func command_redist() (string, string) {
	header := "Installing redistributables / Solving error 0xc000007b"

	var output = []string{"Be sure to install both DirectX and VC++ Redistributable 2005 from:",
	"- https://www.microsoft.com/en-us/download/details.aspx?id=26347",
	"    - Be sure you download the `x86` version.",
	"- https://www.microsoft.com/en-us/download/details.aspx?id=35",
	"",
	"If you need more information, please see `!iw4x vcredist` and `!iw4x directx`"}

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
	"    - Remove IW4x rawfiles, or replace by re-extracting release.zip (available at https://github.com/iw4x/iw4x-rawfiles/releases) in the same directory.",
	"    - Replace iw4x.dll with a new copy, available at: https://github.com/iw4x/iw4x-client/releases",
	"",
	"If this did not solve your issue, please describe the problem in the <#1382046854753026079> channel."}

	body := strings.Join(output[:], "\n")

	return header, body
}

func command_dedicated() (string, string) {
	header := "Setting up a dedicated server"

	var output = []string{"There are detailed instructions at:",
	"- https://" + base_url + "servers/dedicated-server/"}

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
	"If this did not solve your issue, please describe the problem in the <#1382046854753026079> channel.",
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
	"- https://" + base_url + "guides/performance"}

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

func command_directx() (string, string) {
	header := "Installing DirectX"

	var output = []string{"Errors referencing `d3d9.dll` and/or `xinput1_3.dll` are a result of a missing or broken DirectX installation.",
	"",
	"1. Make sure you haven't downloaded any of the dll files from the internet, if you have, **please delete them**.",
	"2. Download DirectX from Microsoft at: https://www.microsoft.com/en-us/download/details.aspx?id=35",
	"3. Run `dxwebsetup.exe` by double-clicking it, and follow the instructions.",
	"4. Try running IW4x again. If you are still facing problems, please restart your computer.",
	"",
	"If this did not solve your issue, please describe the problem in the <#1382046854753026079> channel."}

	body := strings.Join(output[:], "\n")

	return header, body
}
