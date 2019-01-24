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
	api     plugin.API
	botUser *model.User
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
	var err error

	user := &model.User{
		Username: "DemoBot",
		Nickname: "DemoBot",
		Email:    "daniel+demobot@schalla.me",
		Password: "12308ßi12ß380sadjhnipoashdjas09dhj",
	}

	s.botUser, err = s.api.GetUserByUsername("DemoBot")

	if s.botUser == nil {
		s.botUser, err = s.api.CreateUser(user)

		if err != nil {
			return err
		}
	}

	data, err := ioutil.ReadFile("plugins/com.dschalla.matterdemo-plugin/server/dist/mattermost_logo.jpg")

	if err != nil {
		return err
	}

	s.api.SetProfileImage(s.botUser.Id, data)
	s.api.UpdateUser(s.botUser)

	teams, err := s.api.GetTeams()

	for _, team := range teams {
		s.api.CreateTeamMember(team.Id, user.Id)
	}

	return nil
}

func (s *Server) SendWelcomePost(channelId string) {

	post := &model.Post{}
	post.ChannelId = channelId
	post.UserId = s.botUser.Id
	channelMember, _ := s.api.GetChannelMember(channelId, s.botUser.Id)

	if channelMember == nil {
		s.api.AddChannelMember(channelId, s.botUser.Id)
	}

	url := *s.api.GetConfig().ServiceSettings.SiteURL

	s.api.LogInfo("PREPARING TO SEND DEMOBOT INTRODUCTION")
	post.Props = model.StringInterface{}
	post.AddProp("from_webhook", "true")

	attachments := []*model.SlackAttachment{
		{
			Title:      "DemoBot Introduction",
			AuthorName: "DemoBot",
			AuthorIcon: "http://www.mattermost.org/wp-content/uploads/2016/04/icon_WS.png",
			Text:       "Welcome to Palo Alto Bank!  Palo Alto Bank is a simulation of Mattermost in action.  We are going to give you a tour of the product and show you why Mattermost is the premier choice to making your team more productive through high trust collaboration.\n\nThis demo consists of a variety of 2-3 minute workflows to highlight Mattermost features and use cases.\n\nViewing all workflows takes about 15 minutes, and you will have access to this site until 6:00 AM UTC so that you can try it out yourself. If you would like to preview Mattermost longer, please reach request a trial at Mattermost.com/trial or via the link at the top of the demo instance.\n\nChoose a script below to start the demo for that particular workflow.  You will see the channel get created in the left-hand channel menu under private channels. Once in the channel, please read the posts and interact with buttons as they become available. You are able to run each workflow multiple times.",
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
			Title:      fmt.Sprintf("Script #%d: %s", i, script.Name),
			Text:       script.Description,
			Actions: []*model.PostAction{
				{
					Name: "Start Script",
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
	s.api.CreatePost(post)
}

func (s *Server) StartScript(teamId, userId, scriptId string) {
	go s.scriptManager.StartScript(teamId, userId, scriptId)
}

func (s *Server) TriggerResponse(channelId, userId, scriptId, responseId string) error {
	err := s.scriptManager.TriggerResponse(responseId, channelId, userId)

	if err != nil {
		s.api.LogError(fmt.Sprintf("Error triggering response: %s", err))
		return err
	}

	return nil
}
