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

func FirstCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if err := godotenv.Load("data.env"); err != nil {
		fmt.Println("Error while trying to load data.env")
	}
	openrouter_ai_api := os.Getenv("AI_API_KEY")

	client := &http.Client{}

	question := `I am giving you my requirements for the server in Discord. I want you to write me the names of the channels that I need to specify.
		 	My requirement is for two three channels, one for communication and the other for important matters, the third one for funny pictures.
		 	Please write to me first the number of channels that I have specified as a number, and then only two names separated by a comma, and nothing else. All of this should be written in one line`

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

	// Бот не будет реагировать на свои же сообщения
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "create channel" {
		msg, err := s.ChannelMessageSend(m.ChannelID, "Окей, я понял, создаю канал!")
		if err != nil {
			fmt.Println("Error while trying to send message")
		}
		msgID := msg.ID

		// Создаю категорию с заданным названием
		s.GuildChannelCreate(m.GuildID, "Main Category", discordgo.ChannelTypeGuildCategory)
		channels, err := s.GuildChannels(m.GuildID)
		if err != nil {
			fmt.Println("can't get channels list")
		}

		for _, c := range channels {
			if c.Name == "Main Category" {
				number_of_channels_str := aiReponseArr[0]
				number_of_channels_int, err := strconv.Atoi(number_of_channels_str)
				if err != nil {
					fmt.Println(err)
				}
				for i := 0; i < number_of_channels_int; i++ {
					channelData := discordgo.GuildChannelCreateData{
						Name:     aiReponseArr[i+1],
						Type:     discordgo.ChannelTypeGuildText,
						ParentID: c.ID,
					}
					s.GuildChannelCreateComplex(m.GuildID, channelData)
				}
			}
		}

		s.ChannelMessageDelete(m.ChannelID, msgID)
	}

	if m.Content == "status" {
		for _, guild := range s.State.Guilds {
			channels, _ := s.GuildChannels(guild.ID)
			for _, c := range channels {
				fmt.Println(c.Type, c.Name)
			}
		}
	}
}
