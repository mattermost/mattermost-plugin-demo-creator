package demo

import "github.com/mattermost/mattermost-server/plugin"

func NewServer (api plugin.API) *Server {
	s := &Server{
		api: api,
	}

	return s
}

type Server struct {
	api plugin.API
}

func (s *Server) Start() {

}