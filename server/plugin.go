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
	requestData := struct{
		UserId string `json:"user_id"`
		PostId string `json:"post_id"`
		ChannelId string `json:"channel_id"`
		TeamId string `json:"team_id"`
		Context map[string]string
	}{}

	if strings.HasPrefix(path, "/start_script") {
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(bodyBytes, &requestData)
		p.server.StartScript(requestData.TeamId, requestData.UserId, requestData.Context["script_id"])
	} else if strings.HasPrefix(path, "/trigger_response") {
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(bodyBytes, &requestData)
		p.server.TriggerResponse(requestData.ChannelId, requestData.UserId, requestData.Context["script_id"], requestData.Context["response_id"])
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
