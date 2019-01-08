package main

import (
	"github.com/DSchalla/MatterDemo-Plugin/server/demo"
	"github.com/mattermost/mattermost-server/model"
	"net/http"
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
	p.server.Start()
	return nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
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
