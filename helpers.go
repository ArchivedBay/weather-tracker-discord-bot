package main

import (
  "fmt"
  "os"
  "strings"
  "regexp"
  "errors"

  "github.com/bwmarrin/discordgo"
)

var (
  prefix      string = "!"
  botID       string
  channelID   string
  commandList []string = []string{"greet"}
)

// Starts the bot and connects to the Discord server
func StartClient()  {
  client, err := discordgo.New("Bot " + os.Getenv("WEATHERBOT_TOKEN"))
  user, err := client.User("@me")
  handleError("BOT_CONNECT", err, true)
  botID = user.ID
  client.AddHandler(commandHandler)
  client.AddHandler(func(client *discordgo.Session, ready *discordgo.Ready){
    err = client.UpdateStatus(0, "A Bot test")
    handleError("BOT_STATUS_UPDATE_FAIL", err, true)

    servers := client.State.Guilds
    if count := len(servers); count > 1 {
      lmsg := fmt.Sprintf("Bot started on %d servers...", count)
      logMsg("SERVER", lmsg)
    }
  })

  err = client.Open()
  handleError("DISCORD_CONNECT", err, true)
  defer client.Close()

  // keeps the "server" alive without consuming CPU
  <-make(chan struct{})
}

// tells the bot how to handle an incoming message
func commandHandler(client *discordgo.Session, msg *discordgo.MessageCreate) {
  user      := msg.Author
  userMsg   := msg.Content
  channelID = msg.ChannelID
  if user.ID == botID || user.Bot {
    // No-op since bot is talking
    return
  }

  // establishes that the user might have tried to type a command
  if string(userMsg[0]) == prefix {
    userMsg       = removeSpecialChars(userMsg)
    command, err := findCommandFromMsg(userMsg)

    if err != nil {
      handleError("COMMAND_NOT_FOUND", err, false)
      _, err := client.ChannelMessageSend(channelID, "Uh oh, I don't know that command!")
      handleError("MESSAGE_SEND_FAIL", err, true)
    } else {
      lmsg := fmt.Sprintf("[%s] %s", msg.Author, command)
      logMsg("MESSAGE", lmsg)
    }
  }
}

// parses out any punctuation from a command
func removeSpecialChars(s string) (fs string){
  reg, err := regexp.Compile("[^\\w ]+")
  handleError("REGEX_PARSE_FAIL", err, false)

  fs = reg.ReplaceAllString(s, "")
  return
}

// looks for commands based on the first word of a message
func findCommandFromMsg(s string) (c string, err error) {
  first := strings.Split(s, " ")[0]
  e := fmt.Sprintf("There is no command called: %s", first)
  err = errors.New(e)
  for _, command := range commandList {
    if first == command {
      c = command
      err = nil
    }
  }

  return
}

// logs a message to the logger
func logMsg(logType string, msg string) {
  msg = fmt.Sprintf("[%s] %s", logType, msg)
  fmt.Println(msg)
}

// logs error messages to the logger
func handleError(userMsg string, err error, shouldExit bool) {
  if err != nil {
    e := fmt.Sprintf("[ERROR - %s] Reason: %s", userMsg, err.Error())
    fmt.Println(e)
    if shouldExit {
      os.Exit(1)
    }
  }
}
