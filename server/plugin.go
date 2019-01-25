package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DSchalla/MatterDemo-Plugin/server/demo"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

type Plugin struct {
	plugin.MattermostPlugin

	server *demo.Server
}

func (p *Plugin) OnActivate() error{
	p.server = demo.NewServer(p.API)
	err := p.server.Start()

	if err != nil && (reflect.ValueOf(err).Kind() == reflect.Ptr && !reflect.ValueOf(err).IsNil()){
		return err
	}

	// Register fallback command

	introCommand := &model.Command{
		Trigger: "demobot_intro",
	}
	p.API.RegisterCommand(introCommand)

	// Send welcome post to town square

	teams, err2 := p.API.GetTeams()
	if err2 != nil {
		return errors.New(err2.Message)
	}

	for _, team := range teams {
		data, err2 := p.API.KVGet("welcomePostTownSquare-" + team.Id)

		if err2 != nil {
			return errors.New(err2.Message)
		}

		channel, err2 := p.API.GetChannelByName(team.Id, "town-square", false)

		if err2 != nil {
			return errors.New(err2.Message)
		}

		var post *model.Post

		if data == nil {
			post = p.server.SendWelcomePost(channel.Id)
		} else {
			post, err2 = p.API.GetPost(string(data))

			if err2 != nil || post == nil {
				post = p.server.SendWelcomePost(channel.Id)
			}
		}

		p.API.KVSet("welcomePostTownSquare-" + team.Id, []byte(post.Id))

	}

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
		err := json.Unmarshal(bodyBytes, &requestData)
		if err != nil {
			p.API.LogError(fmt.Sprintf("Error decoding JSON: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte{})
			if err != nil {
				p.API.LogError(fmt.Sprintf("Error sending response: %s", err))
			}
			return
		}
		p.server.StartScript(requestData.TeamId, requestData.UserId, requestData.Context["script_id"])
	} else if strings.HasPrefix(path, "/trigger_response") {
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(bodyBytes, &requestData)

		if err != nil {
			p.API.LogError(fmt.Sprintf("Error decoding JSON: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte{})
			if err != nil {
				p.API.LogError(fmt.Sprintf("Error sending response: %s", err))
			}
			return
		}

		err = p.server.TriggerResponse(requestData.ChannelId, requestData.UserId, requestData.Context["script_id"], requestData.Context["response_id"])
		if err != nil {

			p.API.LogError(fmt.Sprintf("Error decoding JSON: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte{})

			if err != nil {
				p.API.LogError(fmt.Sprintf("Error sending response: %s", err))
			}
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte{})
	if err != nil {
		p.API.LogError(fmt.Sprintf("Error sending response: %s", err))
	}
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
