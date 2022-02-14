package model

type SlackMessageColor string

const (
	Success SlackMessageColor = "#36a64f"
	Error   SlackMessageColor = "#e63939"
)

type SlackEvent struct {
	Token string `json:"token"`
	Event Event  `json:"event"`
}

type Event struct {
	Type      string      `json:"type"`
	TimeStamp string      `json:"ts"`
	Text      string      `json:"text"`
	Channel   string      `json:"channel"`
	User      string      `json:"user"`
	Files     []SlackFile `json:"files"`
}

type SlackFile struct {
	ID        string `json:"id"`
	CreatedAt int    `json:"created"`
	Name      string `json:"name"`
	Mimetype  string `json:"mimetype"`
	FileType  string `json:"filetype"`
	User      string `json:"user"`
	Url       string `json:"url_private_download"`
	Size      int    `json:"size"`
}
