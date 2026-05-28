package main

import (
	"context"
	"discordBot/handlers"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
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

type ComponentsJson struct {
	Component []GetValueJSON `json:"components"`
}

type GetValueJSON struct {
	Value string `json:"value"`
}

func CreateNewComponentsJson() *ComponentsJson {
	return &ComponentsJson{}
}

func CreateNewGetValueJSON() *GetValueJSON {
	return &GetValueJSON{Value: ""}
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

	//session.AddHandler(handlers.FirstCommand)
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		fmt.Println("type:", i.Type)
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := handlers.CommandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionModalSubmit:
			data := i.ModalSubmitData()

			if data.CustomID == "setup server modal" {
				if err != nil {
					panic(err)
				}
				textAmntByte, err := data.Components[0].MarshalJSON()
				if err != nil {
					fmt.Println(1)
					panic(err)
				}
				textDescByte, err := data.Components[1].MarshalJSON()
				if err != nil {
					fmt.Println(2)
					panic(err)
				}
				voiceAmntByte, err := data.Components[2].MarshalJSON()
				if err != nil {
					fmt.Println(3)
					panic(err)
				}
				voiceDescByte, err := data.Components[3].MarshalJSON()
				if err != nil {
					fmt.Println(4)
					panic(err)
				}

				textAmntComponent := CreateNewComponentsJson()
				json.Unmarshal(textAmntByte, &textAmntComponent)
				textDescComponent := CreateNewComponentsJson()
				json.Unmarshal(textDescByte, &textDescComponent)
				voiceAmntComponent := CreateNewComponentsJson()
				json.Unmarshal(voiceAmntByte, &voiceAmntComponent)
				voiceDescComponent := CreateNewComponentsJson()
				json.Unmarshal(voiceDescByte, &voiceDescComponent)

				textAmntComponentInt, err := strconv.Atoi(textAmntComponent.Component[0].Value)
				voiceAmntComponentInt, err := strconv.Atoi(voiceAmntComponent.Component[0].Value)

				if err != nil {
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Число каналов должно быть числом",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
				} else if textAmntComponentInt < 0 || voiceAmntComponentInt < 0 {
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Я не могу создать отрицательное число каналов",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
				} else if textAmntComponentInt == 0 && voiceAmntComponentInt == 0 {
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Ну.. окей, я ничего не буду делать, раз ты каналы не хочешь видеть",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					handlers.CreateChannels(s, i, textAmntComponent.Component[0].Value, textDescComponent.Component[0].Value, voiceAmntComponent.Component[0].Value, voiceDescComponent.Component[0].Value)
				} else {
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Требования понял, создаю каналы!",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					handlers.CreateChannels(s, i, textAmntComponent.Component[0].Value, textDescComponent.Component[0].Value, voiceAmntComponent.Component[0].Value, voiceDescComponent.Component[0].Value)
				}
			} else if data.CustomID == "delete server modal" {
				if err != nil {
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Я тя понял, тестируй на здоровье",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
				}
			}
		}
	})

	session.Identify.Intents = discordgo.IntentGuildMessages

	err = session.Open()
	if err != nil {
		fmt.Printf("Error while trying to open session: %v", err)
	}
	defer session.Close()

	registeredCommands := make([]*discordgo.ApplicationCommand, len(handlers.Commands))
	for i, v := range handlers.Commands {
		cmd, err := session.ApplicationCommandCreate(session.State.User.ID, client_id, v)
		if err != nil {
			fmt.Printf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-sc
	fmt.Println("Graceful Shutdown")
}
