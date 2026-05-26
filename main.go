package main

import (
	"context"
	"discordBot/handlers"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v3"
)

type AI struct {
	client openai.Client
	ctx    context.Context
}

func CreateNewAI(client openai.Client) *AI {
	return &AI{client: client, ctx: context.Background()}
}

func main() {
	if err := godotenv.Load("data.env"); err != nil {
		fmt.Println("Error while trying to load data.env")
	}
	bot_token := os.Getenv("BOT_TOKEN")
	public_key := os.Getenv("PUBLIC_KEY")
	client_id := os.Getenv("CLIENT_ID")

	_ = bot_token
	_ = public_key
	_ = client_id

	session, err := discordgo.New("Bot " + bot_token)
	if err != nil {
		fmt.Println("Error while trying to create session")
	}

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Bot is ready!")
	})

	session.AddHandler(handlers.FirstCommand)

	session.Identify.Intents = discordgo.IntentGuildMessages

	err = session.Open()
	if err != nil {
		fmt.Printf("Error while trying to open session: %v", err)
	}
	defer session.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-sc
	fmt.Println("Graceful Shutdown")
}
