package demo

type Script struct {
	Id          string
	Name        string
	Description string
	Priority 	int
	Channel     ScriptChannel
	Users       []ScriptUser
	Messages    []ScriptMessage
	Responses   []ScriptResponse
}

type ScriptChannel struct {
	Id          string
	Name        string
	Description string
}

type ScriptUser struct {
	Id       string
	Name     string
	Position string
	Bot      bool
	SystemId string
}

type ScriptMessage struct {
	UserId      string `yaml:"user_id"`
	Text        string
	Attachments []ScriptAttachment
	Reactions   []ScriptReaction
	PostDelay   int `yaml:"post_delay"`
	Replies		[]ScriptMessage
}

type ScriptAttachment struct {
	Title       string
	TitleLink   string
	Color       string
	AuthorName  string `yaml:"author_name"`
	AuthorImage string `yaml:"author_image"`
	Fields      []ScriptAttachmentField
	Actions     []ScriptAttachmentAction
}

type ScriptAttachmentField struct {
	Title string
	Value string
	Short bool
}

type ScriptAttachmentAction struct {
	Name       string
	ResponseId string `yaml:"response_id"`
}

type ScriptReaction struct {
	Id     string
	UserId string `yaml:"user_id"`
	Delay  int
}

type ScriptResponse struct {
	Id      string
	Message ScriptMessage
}

func (s *Script) GetUserById(id string) ScriptUser {
	for _, user := range s.Users {
		if id == user.Id {
			return user
		}
	}

	return ScriptUser{}
}

func (s *Script) GetUserBySystemId(systemId string) ScriptUser {
	for _, user := range s.Users {
		if systemId == user.SystemId {
			return user
		}
	}

	return ScriptUser{}
}