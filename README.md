# IW4x Discord bot

This is the bot for the [IW4x](https://iw4x.io/) Discord server, and at the moment primarily provides functionality to users seeking basic help / information.

# Current valid command list:

- `!iw4x help` - Displays this help dialog
- `!iw4x install` - Information about installing IW4x.
- `!iw4x docs` - IW4x documentation
- `!iw4x discord` - Displays the invite link for this server
- `!iw4x github` - Displays the link to IW4x source code.
- `!iw4x repair` - Information on repairing your game files
- `!iw4x dedicated` - Information on setting up a dedicated server
- `!iw4x vcredist` - Information on installing VC++ Redistributables
- `!iw4x unlockstats` - Information on unlocking all items in the game
- `!iw4x performance` - Information on improving the games performance
- `!iw4x fps` - Information on changing the FPS limit
- `!iw4x fov` - Information on changing your FOV
- `!iw4x nickname` - Information on changing your in-game name
- `!iw4x console` - Information on the in-game command console
- `!iw4x dxr` - Information on installing DirectX and VC++ Redistributables (Error 0xc000007b)
- `!iw4x rawfiles` - Information on installing/repairing iw4x-rawfiles
- `!iw4x game` - Information on supported copies of MW2
- `!iw4x dxvk` - Information on setting up DXVK
- `!iw4x dlc` - Information on MW2 and IW4x DLC

# Building

System dependencies: `go`, (see go.mod for versioning info), though this should build fine on older versions (within reason) assuming the minimum requirement for the discordgo version defined in go.mod is met. 

1. Clone the source, move to its directory.
2. `go get github.com/bwmarrin/discordgo` to pull the discordgo library
3. `go build -ldflags="-s -w"` in the source directory, the `ldflags` here strips the binary to reduce resource usage
4. After building, you should have a resulting `iw4x-discord-bot` binary, done!

# Running

As this produces a single binary, it can simply be run as (preferably) part of a system service (it is service manager agnostic), or with a script, or in a tmux/screen session, or however else desired.

The environment variable `IW4X_DISCORD_BOT_TOKEN` must be set- and should of course contain the token for the bot.

# Misc

https://discord.com/invite/pV2qJscTXf
