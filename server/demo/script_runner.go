package demo

import (
	"errors"
	"fmt"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"math/rand"
	"strconv"
	"time"
)

func NewScriptRunner(api plugin.API, script Script, botId, teamId, userId string) (*ScriptRunner, error) {
	s := &ScriptRunner{}
	s.api = api
	s.script = script
	s.botId = botId
	s.teamId = teamId
	s.creatorId = userId
	s.randomNr = strconv.Itoa(rand.Intn(99999))

	err := s.createChannel()

	if err != nil {
		return nil, err
	}

	return s, nil
}

type ScriptRunner struct {
	api       plugin.API
	script    Script
	botId     string
	teamId    string
	creatorId string
	channelId string
	randomNr  string
}

func (sr *ScriptRunner) Start() error {

	for i, user := range sr.script.Users {
		systemUser, err := sr.api.GetUserByUsername(user.Id)

		if systemUser == nil {
			newUser := &model.User{
				Username: user.Id,
				Nickname: user.Name,
				Email:    user.Id + "-sample-mail@example.com",
				Password: user.Id + "thisshouldbechanged",
			}

			systemUser, err = sr.api.CreateUser(newUser)
			sr.api.LogInfo("Creating User for script")
			if err != nil {
				sr.api.LogError(fmt.Sprintf("Error creating user for Script: %s", err.Message))
				continue
			}
		}

		teamMember, _ := sr.api.GetTeamMember(sr.teamId, systemUser.Id)

		if teamMember == nil {
			_, err := sr.api.CreateTeamMember(sr.teamId, systemUser.Id)

			if err != nil {
				sr.api.LogError(fmt.Sprintf("Error creating team member for Script: %s", err.Message))
				continue
			}
		}

		sr.api.AddChannelMember(sr.channelId, systemUser.Id)
		sr.script.Users[i].SystemId = systemUser.Id
	}

	sr.api.AddChannelMember(sr.channelId, sr.creatorId)
	sr.sendScriptProlog()
	time.Sleep(time.Second * time.Duration(10))

	sr.api.LogDebug("Starting Post Generation...")

	for _, message := range sr.script.Messages {
		sr.sendMessage(message, "")
	}

	return nil
}

func (sr *ScriptRunner) GetChannelId() string {
	return sr.channelId
}

func (sr *ScriptRunner) createChannel() error {

	channelExists, _ := sr.api.GetChannelByName(sr.teamId, sr.script.Channel.Id+sr.randomNr, false)
	if channelExists != nil {
		sr.api.DeleteChannel(channelExists.Id)
	}

	channel := &model.Channel{
		Name:        sr.script.Channel.Id + sr.randomNr,
		DisplayName: sr.script.Channel.Name + " (#" + sr.randomNr + ")",
		Header:      sr.script.Channel.Description,
		TeamId:      sr.teamId,
		Type:        model.CHANNEL_PRIVATE,
	}

	var err *model.AppError
	channel, err = sr.api.CreateChannel(channel)

	if err != nil {
		return errors.New(fmt.Sprintf("error creating channel for script %s and user %s: %s", sr.script.Id, sr.creatorId, err.Message))
	}

	sr.channelId = channel.Id
	return nil
}

func (sr *ScriptRunner) TriggerResponse(responseId, userId string) error {
	var response ScriptResponse
	for _, tmpResponse := range sr.script.Responses {
		if tmpResponse.Id == responseId {
			response = tmpResponse
		}
	}

	if response.Id == "" {
		return errors.New("response not found")
	}

	return sr.sendMessage(response.Message, "")
}

func (sr *ScriptRunner) sendScriptProlog() {
	user, _ := sr.api.GetUser(sr.creatorId)

	post := &model.Post{}
	post.ChannelId = sr.channelId
	post.UserId = sr.botId
	post.AddProp("attachments", []*model.SlackAttachment{
		{
			Title:      "Script: " + sr.script.Name,
			AuthorName: "DemoBot",
			AuthorIcon: "http://www.mattermost.org/wp-content/uploads/2016/04/icon_WS.png",
			Text:       "Hello @" + user.Username + " and welcome to the " + sr.script.Name + " demonstration. Starting in 10 seconds.",
		},
	})
	sr.api.CreatePost(post)
}

func (sr *ScriptRunner) sendMessage(message ScriptMessage, rootId string) error {

	url := *sr.api.GetConfig().ServiceSettings.SiteURL

	post := &model.Post{}
	post.ChannelId = sr.channelId
	post.Message = message.Text
	post.RootId = rootId

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
				Integration: &model.PostActionIntegration{
					URL: url + "/plugins/com.dschalla.matterdemo-plugin/trigger_response",
					Context: map[string]interface{}{
						"response_id": action.ResponseId,
						"script_id":   sr.script.Id,
					},
				},
			})
		}
		attachments = append(attachments, &slackAttachment)
	}

	post.AddProp("attachments", attachments)

	user := sr.script.GetUserById(message.UserId)

	if user.SystemId == "" {
		sr.api.LogDebug("User " + message.UserId + " not found!")
		return errors.New("user not found in users map")
	}

	post.UserId = user.SystemId

	if user.Bot {
		post.AddProp("from_webhook", "true")
	}

	post, err := sr.api.CreatePost(post)

	if err != nil {
		return errors.New("error creating post")
	}

	go sr.createReactions(post.Id, message.Reactions)

	time.Sleep(time.Second * time.Duration(message.PostDelay))

	for _, reply := range message.Replies {
		sr.sendMessage(reply, post.Id)
	}

	return nil
}

func (sr *ScriptRunner) createReactions (postId string, reactions []ScriptReaction) {
	for _, reaction := range reactions {
		go func(reaction ScriptReaction) {

			user := sr.script.GetUserById(reaction.UserId)

			if user.SystemId == "" {
				sr.api.LogWarn("Error getting user id for reaction")
				return
			}

			if reaction.Delay != 0 {
				time.Sleep(time.Second * time.Duration(reaction.Delay))
			}

			r := &model.Reaction{
				UserId:    user.Id,
				PostId:    postId,
				EmojiName: reaction.Id,
			}
			sr.api.AddReaction(r)
		}(reaction)
	}
}
