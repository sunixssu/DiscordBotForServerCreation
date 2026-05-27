package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type Ai_Json struct {
	Model       string  `json:"model"`
	Messages    []Msgs  `json:"messages"`
	Temperature float64 `json:"temperature"`
	Max_tokens  int     `json:"max_tokens"`
}

type Msgs struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AiResponse struct {
	Choices []Choices `json:"choices"`
}

type Choices struct {
	Message Msgs `json:"message"`
}

var (
	Commands = []*discordgo.ApplicationCommand{
		{
			Name:        "setup-channel",
			Description: "creates channels based on your prompt",
		},
	}

	CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"setup-channel": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "setup server modal",
					Title:    "Первичная информация",
					Flags:    discordgo.MessageFlagsIsComponentsV2,
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "text_chan_amount",
									Label:       "Сколько текстовых каналов создать?",
									Style:       discordgo.TextInputShort,
									Placeholder: "Например 0, 2, ...",
									Required:    true,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "text_chan_description",
									Label:       "Что там будет находиться?",
									Style:       discordgo.TextInputParagraph,
									Placeholder: "Например первый канал для общения, второй для мемов, третий для...",
									Required:    false,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "voice_chan_amount",
									Label:       "Сколько голосовых каналов создать?",
									Style:       discordgo.TextInputShort,
									Placeholder: "Например 0, 2, ...",
									Required:    true,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "voice_chan_description",
									Label:       "Что там будет находиться?",
									Style:       discordgo.TextInputParagraph,
									Placeholder: "Например первый канал для игр, второй для просмотра фильмов, третий для...",
									Required:    false,
								},
							},
						},
					},
				},
			})
			if err != nil {
				panic(err)
			}
		},
	}
)

func CreateChannels(s *discordgo.Session, interact *discordgo.InteractionCreate, textAmount string, textDesc string, voiceAmount string, voiceDesc string) {
	if err := godotenv.Load("data.env"); err != nil {
		fmt.Println("Error while trying to load data.env")
	}
	openrouter_ai_api := os.Getenv("AI_API_KEY")

	client := &http.Client{}

	quest_p1 := "I am giving you my requirements for the server in Discord. I want you to write me the names of the channels that I need to specify."
	quest_p2 := "My requirement is for " + textAmount + " text channels and " + voiceAmount + " voice channels. Also I want to see emojis in brackets, that would suite the purpose of channel"
	quest_p3 := "My requirements for text channels is: " + textDesc + ", and for voice channels: " + voiceDesc
	quest_p4 := "Please write me only the names of the channels that you create in single line, at first - text channels, than - voice channels. For example: [emoji]text1, [emoji]text2, [emoji]text3, [emoji]voice1, [emoji]voice2."
	/*
		question := `I am giving you my requirements for the server in Discord. I want you to write me the names of the channels that I need to specify.
			 	My requirement is for three channels, one for communication and the other for important matters, the third one for funny pictures.
			 	Please write to me first the number of channels that I have specified as a number, and then only two names separated by a comma, and nothing else. All of this should be written in one line`
	*/
	question := quest_p1 + "\n" + quest_p2 + "\n" + quest_p3 + "\n" + quest_p4

	new_msgs_for_body_json_temp := Msgs{Role: "user", Content: question}
	msg_map := make([]Msgs, 0)
	msg_map = append(msg_map, new_msgs_for_body_json_temp)
	new_body_json_temp := &Ai_Json{Model: "mistral-large-latest", Messages: msg_map, Temperature: 0.7, Max_tokens: 1000}

	b, err := json.Marshal(new_body_json_temp)
	if err != nil {
		fmt.Println("error while trying to convert struct to json")
	}

	req, err := http.NewRequest("POST", "https://api.mistral.ai/v1/chat/completions", bytes.NewBuffer(b))
	if err != nil {
		fmt.Println("err while trying to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openrouter_ai_api)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("client error:", err)
	}
	resp_b, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading resp body:", err)
	}

	var AiResponse AiResponse
	json.Unmarshal(resp_b, &AiResponse)

	var aiReponseStr string = AiResponse.Choices[0].Message.Content
	aiReponseStr = strings.Replace(aiReponseStr, "\n", ",", -1)
	aiReponseArr := strings.Split(aiReponseStr, ",")

	// Создаю категорию с заданным названием
	s.GuildChannelCreate(interact.GuildID, "Main Category", discordgo.ChannelTypeGuildCategory)
	channels, err := s.GuildChannels(interact.GuildID)
	if err != nil {
		fmt.Println("can't get channels list")
	}

	for _, c := range channels {
		if c.Name == "Main Category" {
			number_of_text_channels_int, err := strconv.Atoi(textAmount)
			if err != nil {
				fmt.Println(err)
			}
			for i := 0; i < number_of_text_channels_int; i++ {
				channelData := discordgo.GuildChannelCreateData{
					Name:     aiReponseArr[i],
					Type:     discordgo.ChannelTypeGuildText,
					ParentID: c.ID,
				}
				s.GuildChannelCreateComplex(interact.GuildID, channelData)
			}
			number_of_voice_channels_int, err := strconv.Atoi(voiceAmount)
			if err != nil {
				fmt.Println(err)
			}
			for i := 0; i < number_of_voice_channels_int; i++ {
				channelData := discordgo.GuildChannelCreateData{
					Name:     aiReponseArr[number_of_text_channels_int+i],
					Type:     discordgo.ChannelTypeGuildVoice,
					ParentID: c.ID,
				}
				s.GuildChannelCreateComplex(interact.GuildID, channelData)
			}
		}
	}
}
