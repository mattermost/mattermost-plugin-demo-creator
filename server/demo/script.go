package demo

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Script struct {
	Name        string
	Description string
	Channel     ScriptChannel
	Users       []ScriptUser
	Bots        []ScriptUser
	Messages    []ScriptMessage
	Responses   []ScriptResponses
}

type ScriptChannel struct {
	Id          string
	Name        string
	Description string
}

type ScriptUser struct {
	Name     string
	Position string
}

type ScriptMessage struct {
	UserId      string
	Message     string
	Attachments []ScriptAttachment
	PostDelay   int
}

type ScriptAttachment struct {
	Title string
	TitleLink string
	Color string
	AuthorName string `yaml:"author_name"`
	AuthorImage string `yaml:"author_image"`
	Fields []ScriptAttachmentField
	Actions []ScriptAttachmentAction
}

type ScriptAttachmentField struct {
	Title string
	Value string
	Short bool
}

type ScriptAttachmentAction struct {
	Label string
	ResponseId string `yaml:"response_id"`
}

type ScriptResponses struct {
	Id      string
	UserId  string
	Message string
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