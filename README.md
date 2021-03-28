# Watchalong

A media synchronization project for mpv.

# Usage

| **Requirements** |
| mpv              |
| LuaSockets       |

## Server

Initialize the server `server.go`:
`go run server.go --address <IP_ADDRESS>:<PORT>`

This server will receive the player status and position from the client and broadcast it to everyone else, allowing everyone to synchronize with each other.

## Client

First, you will need to copy the `client.lua` file to your mpv scripts directory.

Playing media files `client.lua`:
`mpv <file/url> --script-opts=address=<IP_ADDRESS>:<PORT>`

You may remove the script when it is not being used or simple rename it to have a `.disable` extension.
