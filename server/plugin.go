package main

import (
	"encoding/json"
	"fmt"
	"github.com/DSchalla/MatterDemo-Plugin/server/demo"
	"github.com/mattermost/mattermost-server/model"
	ioutil "io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-server/plugin"
)

type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the Configuration.
	configurationLock sync.RWMutex

	// Configuration is the active plugin Configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *Configuration

	server *demo.Server
}

func (p *Plugin) OnActivate() error{
	p.server = demo.NewServer(p.API)
	err := p.server.Start()

	if err != nil {
		return err
	}

	introCommand := &model.Command{
		Trigger: "demobot_intro",
		AutoComplete: true,
		AutoCompleteDesc: "Start Introduction to Demobot capabilities and menu of available demonstrations",
	}
	p.API.RegisterCommand(introCommand)

	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	p.API.LogInfo("COMMAND RECEIVED: ", args.Command, args.SiteURL, args.UserId, args.ChannelId)

	if strings.HasPrefix(args.Command, "/demobot_intro") {
		p.server.SendWelcomePost(args.ChannelId)
	}
	return &model.CommandResponse{}, nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.API.LogInfo(fmt.Sprintf("REQUEST URL: %s", r.URL.Path))

	path := r.URL.Path

	if strings.HasPrefix(path, "/start_script") {
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		p.API.LogDebug(fmt.Sprintf("BODY: %+v", bodyBytes))
		requestData := struct{
			UserId string `json:"user_id"`
			PostId string `json:"post_id"`
			ChannelId string `json:"channel_id"`
			TeamId string `json:"team_id"`
			Context map[string]string
		}{}
		json.Unmarshal(bodyBytes, &requestData)
		p.API.LogDebug(fmt.Sprintf("PARSED: %+v", requestData))
		p.server.StartScript(requestData.TeamId, requestData.UserId, requestData.Context["script_id"])
	}
	w.WriteHeader(http.StatusOK)
}

func (p *Plugin) createSamplePost() {
	user, _ := p.API.GetUserByEmail("daniel@schalla.me")
	channel, _ := p.API.GetChannel("999x9itiob8j5extprm4w6aana")
	post := model.Post{}
	post.UserId = user.Id
	post.ChannelId = channel.Id
	post.Props = model.StringInterface{}
	post.Props["attachments"] = []*model.SlackAttachment{
		{
			Title: "Zendesk Issue #12345",
			TitleLink: "http://zendesk.com/12345",
			Color: "#ff0000",
			AuthorName: "ZenDesk",
			AuthorIcon: "https://d1eipm3vz40hy0.cloudfront.net/images/p-brand/zendesk-wordmark.svg",
			Fields: []*model.SlackAttachmentField{
				{
					Title: "No#",
					Value: "12345",
					Short: true,
				},
				{
					Title: "SLA Assigned",
					Value: "15 Min",
					Short: true,
				},
				{
					Title: "Severity",
					Value: "S1",
					Short: true,
				},
			},
			Actions: []*model.PostAction{
				{
					Name: "Confirm Incident",
					Integration: &model.PostActionIntegration{
						URL: "http://localhost:8065/plugins/",
					},
				},
				{
					Name: "Close Incident as False Positive",
					Integration: &model.PostActionIntegration{
						URL: "http://localhost:8065/plugins/",
					},
				},
			},
		},
	}
	p.API.CreatePost(&post)
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
