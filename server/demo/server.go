package demo

import (
	"fmt"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"io/ioutil"
	"sort"
)

func NewServer(api plugin.API) *Server {
	s := &Server{
		api: api,
	}

	return s
}

type Server struct {
	api           plugin.API
	botUser       *model.User
	scriptManager *ScriptManager
}

func (s *Server) Start() error {
	var err error
	s.scriptManager, err = NewScriptManager(s.api, "plugins/com.dschalla.matterdemo-plugin/scripts")

	if err != nil {
		return err
	}

	err = s.RegisterBotUser()

	if err != nil {
		return err
	}

	s.scriptManager.SetBotId(s.botUser.Id)

	return nil
}

func (s *Server) RegisterBotUser() error {
	var err *model.AppError

	user := &model.User{
		Username: "DemoBot",
		Nickname: "DemoBot",
		Email:    "demobot@demo.mattermost.com",
		Password: model.NewId(),
	}

	s.botUser, err = s.api.GetUserByUsername("DemoBot")

	if s.botUser == nil {
		s.botUser, err = s.api.CreateUser(user)

		if err != nil {
			return err
		}
	}

	data, err2 := ioutil.ReadFile("plugins/com.dschalla.matterdemo-plugin/pictures/demobot.png")

	if err2 != nil {
		s.api.LogError(fmt.Sprintf("Error reading bot picture: %s", err))
		return err
	}

	err = s.api.SetProfileImage(s.botUser.Id, data)
	if err != nil {
		s.api.LogError(fmt.Sprintf("Error setting profile picture: %s", err))
		return err
	}

	_, err = s.api.UpdateUser(s.botUser)
	if err != nil {
		return err
	}

	teams, err := s.api.GetTeams()

	for _, team := range teams {
		member, err := s.api.GetTeamMember(team.Id, s.botUser.Id)

		if err != nil {
			s.api.LogError(fmt.Sprintf("Error getting team membership: %s", err))
		}

		if member == nil {
			_, err := s.api.CreateTeamMember(team.Id, s.botUser.Id)

			if err != nil {
				s.api.LogError(fmt.Sprintf("Error creating team membership: %s", err))
			}
		}
	}

	return nil
}

func (s *Server) SendWelcomePost(channelId string) *model.Post {

	post := &model.Post{}
	post.ChannelId = channelId
	post.UserId = s.botUser.Id
	channelMember, _ := s.api.GetChannelMember(channelId, s.botUser.Id)

	if channelMember == nil {
		_, err := s.api.AddChannelMember(channelId, s.botUser.Id)
		if err != nil {
			s.api.LogError(fmt.Sprintf("Error adding channel member in welcome post: %s", err))
		}
	}

	url := *s.api.GetConfig().ServiceSettings.SiteURL

	s.api.LogInfo("PREPARING TO SEND DEMOBOT INTRODUCTION")
	post.Props = model.StringInterface{}

	attachments := []*model.SlackAttachment{
		{
			Title:      "DemoBot Introduction",
			AuthorName: "DemoBot",
			AuthorIcon: "http://www.mattermost.org/wp-content/uploads/2016/04/icon_WS.png",
			Text:       "Welcome to Acme Bank! :wave: Acme Bank is a simulation of Mattermost in action. We are going to give you a tour of the product and show you why Mattermost is the premier choice to making your team more productive through high trust collaboration.\n\nTo get started, choose a demo from the options below and click “Start Demo”. You will be shown a short example scenario to give you some ideas on how other teams use Mattermost. Feel free to click around and interact with what you see! If you need more time for your organization to try Mattermost, please request a [trial](https://mattermost.com/trial/).",
		},
	}

	i := 1

	var scripts []Script

	for _, script := range s.scriptManager.GetScripts() {
		scripts = append(scripts, script)
	}

	sort.Slice(scripts, func(i, j int) bool { return scripts[i].Priority < scripts[j].Priority })

	for _, script := range scripts {
		attachments = append(attachments, &model.SlackAttachment{
			Text: fmt.Sprintf("**%s**: %s", script.Name, script.Description),
			Actions: []*model.PostAction{
				{
					Name: "START DEMO",
					Integration: &model.PostActionIntegration{
						URL: url + "/plugins/com.dschalla.matterdemo-plugin/start_script",
						Context: map[string]interface{}{
							"script_id": script.Id,
						},
					},
				},
			},
		})
		i++
	}
	post.AddProp("attachments", attachments)

	config := s.api.GetConfig()
	*config.TeamSettings.ExperimentalTownSquareIsReadOnly = false
	err := s.api.SaveConfig(config)
	if err != nil {
		s.api.LogError(fmt.Sprintf("Error updating config in welcome post: %s", err))
	}

	post, err = s.api.CreatePost(post)

	if err != nil {
		s.api.LogError(fmt.Sprintf("Error creating welcome post: %s", err))
	}

	*config.TeamSettings.ExperimentalTownSquareIsReadOnly = true
	err = s.api.SaveConfig(config)
	if err != nil {
		s.api.LogError(fmt.Sprintf("Error updating config in welcome post: %s", err))
	}

	return post
}

func (s *Server) StartScript(teamId, userId, scriptId string) {
	go s.scriptManager.StartScript(teamId, userId, scriptId)
}

func (s *Server) TriggerResponse(channelId, userId, scriptId, responseId string) error {
	s.api.LogDebug(fmt.Sprintf("Trigger new response for script id %s for team %s and user %s", channelId, userId, responseId))

	err := s.scriptManager.TriggerResponse(responseId, channelId, userId)

	if err != nil {
		s.api.LogError(fmt.Sprintf("Error triggering response: %s", err))
		return err
	}
	s.api.LogDebug(fmt.Sprintf("End new response for script id %s for team %s and user %s", channelId, userId, responseId))

	return nil
}
