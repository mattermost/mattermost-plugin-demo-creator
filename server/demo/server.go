package demo

import (
	"errors"
	"fmt"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"io/ioutil"
	"strconv"
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
	scripts []Script
}

func (s *Server) Start() error {
	var err error
	s.scripts, err = LoadScriptsFromFile("plugins/com.dschalla.matterdemo-plugin/server/dist/script.yml")

	if err != nil {
		return err
	}

	err = s.RegisterBotUser()

	if err != nil {
		return err
	}

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
			Text:       "Morbi pellentesque enim quis libero congue, vitae congue metus feugiat. Nam justo ex, convallis sit amet dolor vulputate, hendrerit consectetur nulla. Suspendisse potenti. Vestibulum et augue tincidunt, fermentum mi ut, facilisis libero. Interdum et malesuada fames ac ante ipsum primis in faucibus. Aenean eu magna quam. Ut massa nibh, ornare et enim sit amet, efficitur aliquet nunc. Nulla nisi nibh, vehicula ultrices vestibulum sed, blandit in nisl. ",
			Fields: []*model.SlackAttachmentField{
				{
					Title: "Number of Scripts",
					Value: strconv.Itoa(len(s.scripts)),
					Short: true,
				},
			},
		},
	}

	i := 1
	for _, script := range s.scripts {
		attachments = append(attachments, &model.SlackAttachment{
			Title:      fmt.Sprintf("Script #%d: %s", i, script.Name),
			AuthorName: "DemoBot",
			AuthorIcon: "http://www.mattermost.org/wp-content/uploads/2016/04/icon_WS.png",
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

func (s *Server) StartScript(teamId, userId, scriptId string) error {
	var script Script

	for _, tmpScript := range s.scripts {
		if tmpScript.Id == scriptId {
			script = tmpScript
			break
		}
	}

	if script.Id == "" {
		return errors.New("scriptId not found")
	}

	go script.RunScript(teamId, s.botUser.Id, userId, s.api)
	return nil
}

func (s *Server) TriggerResponse(teamId, userId, scriptId, responseId string) error {
	var script Script

	for _, tmpScript := range s.scripts {
		if tmpScript.Id == scriptId {
			script = tmpScript
			break
		}
	}

	if script.Id == "" {
		return errors.New("scriptId not found")
	}

	go script.RunScript(teamId, s.botUser.Id, userId, s.api)
	return nil
}
