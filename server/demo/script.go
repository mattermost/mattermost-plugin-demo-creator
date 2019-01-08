package demo

import (
	"fmt"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"strconv"
	"time"
)

type Script struct {
	Id          string
	Name        string
	Description string
	Channel     ScriptChannel
	Users       []ScriptUser
	Messages    []ScriptMessage
	Responses   []ScriptResponses
}

type ScriptChannel struct {
	Id          string
	Name        string
	Description string
}

type ScriptUser struct {
	Id       string
	Name     string
	Position string
	Bot      bool
}

type ScriptMessage struct {
	UserId      string `yaml:"user_id"`
	Message     string
	Attachments []ScriptAttachment
	PostDelay   int `yaml:"post_delay"`
}

type ScriptAttachment struct {
	Title       string
	TitleLink   string
	Color       string
	AuthorName  string `yaml:"author_name"`
	AuthorImage string `yaml:"author_image"`
	Fields      []ScriptAttachmentField
	Actions     []ScriptAttachmentAction
}

type ScriptAttachmentField struct {
	Title string
	Value string
	Short bool
}

type ScriptAttachmentAction struct {
	Name      string
	ResponseId string `yaml:"response_id"`
}

type ScriptResponses struct {
	Id      string
	UserId  string
	Message string
}

func (s *Script) RunScript(teamId, botId, userId string, api plugin.API) {
	rand.Seed(time.Now().Unix())
	randomNr := rand.Intn(99999)

	api.LogWarn("TEAM ID: " + teamId)
	api.LogWarn("USER ID: " + userId)

	channelExists, _ := api.GetChannelByName(teamId, s.Channel.Id+strconv.Itoa(randomNr), false)
	if channelExists != nil {
		api.DeleteChannel(channelExists.Id)
	}

	channel := &model.Channel{
		Name:        s.Channel.Id + strconv.Itoa(randomNr),
		DisplayName: s.Channel.Name + " " + strconv.Itoa(randomNr),
		TeamId:      teamId,
		Type:        model.CHANNEL_OPEN,
	}

	var err *model.AppError
	channel, err = api.CreateChannel(channel)

	if err != nil {
		api.LogError(fmt.Sprintf("Error creating channel for Script: %s", err.Message))
		return
	}

	users := map[string]*model.User{}

	for _, user := range s.Users {
		systemUser, _ := api.GetUserByUsername(user.Id)
		api.LogInfo("Fetching User for script")
		if systemUser == nil {
			newUser := &model.User{
				Username: user.Id,
				Nickname: user.Name,
				Email:    user.Id + "-sample-mail@example.com",
				Password: user.Id + "thisshouldbechanged",
			}

			systemUser, err = api.CreateUser(newUser)
			api.LogInfo("Creating User for script")
			if err != nil {
				api.LogError(fmt.Sprintf("Error creating user for Script: %s", err.Message))
				continue
			}
		}

		teamMember, _ := api.GetTeamMember(teamId, systemUser.Id)

		if teamMember == nil {
			_, err = api.CreateTeamMember(teamId, systemUser.Id)

			if err != nil {
				api.LogError(fmt.Sprintf("Error creating team member for Script: %s", err.Message))
				continue
			}
		}

		api.AddChannelMember(channel.Id, systemUser.Id)
		users[user.Id] = systemUser
	}

	api.AddChannelMember(channel.Id, userId)

	user, _ := api.GetUser(userId)

	post := &model.Post{}
	post.ChannelId = channel.Id
	post.UserId = botId
	post.AddProp("attachments", []*model.SlackAttachment{
		{
			Title:      "Script: " + s.Name,
			AuthorName: "DemoBot",
			AuthorIcon: "http://www.mattermost.org/wp-content/uploads/2016/04/icon_WS.png",
			Text:       "Hello @" + user.Username + " and welcome to the " + s.Name + " demonstration. Starting in 10 seconds.",
		},
	})
	api.CreatePost(post)

	time.Sleep(time.Second * time.Duration(10))
	api.LogDebug("Starting Post Generation...")
	for _, message := range s.Messages {
		post := &model.Post{}
		post.ChannelId = channel.Id
		post.Message = message.Message
		var attachments []*model.SlackAttachment
		for _, attachment := range message.Attachments {
			slackAttachment := model.SlackAttachment{}
			slackAttachment.Title = attachment.Title
			slackAttachment.TitleLink = attachment.TitleLink
			slackAttachment.AuthorName = attachment.AuthorName
			slackAttachment.Color = attachment.Color

			for _, field := range attachment.Fields {
				slackAttachment.Fields = append(slackAttachment.Fields, &model.SlackAttachmentField{
					Title: field.Title,
					Value: field.Value,
					Short: field.Short,
				})
			}

			for _, action := range attachment.Actions {
				slackAttachment.Actions = append(slackAttachment.Actions, &model.PostAction{
					Name: action.Name,
				})
			}
			attachments = append(attachments, &slackAttachment)
		}

		post.AddProp("attachments", attachments)

		tmpUser, ok := users[message.UserId]
		if !ok {
			api.LogDebug("User " + message.UserId + " not found!")
			continue
		}
		post.UserId = tmpUser.Id

		for _, tmpBots := range s.Users {
			if message.UserId == tmpBots.Id && tmpBots.Bot {
				post.AddProp("from_webhook", true)
				break
			}
		}

		reaction := &model.Reaction{

		}
		api.AddReaction()

		api.CreatePost(post)
		time.Sleep(time.Second * time.Duration(message.PostDelay))
	}
}

func LoadScriptsFromFile(filepath string) ([]Script, error) {
	data, err := ioutil.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	helper := struct {
		Scripts []Script
	}{}

	err = yaml.Unmarshal(data, &helper)

	if err != nil {
		return nil, err
	}

	return helper.Scripts, nil
}
