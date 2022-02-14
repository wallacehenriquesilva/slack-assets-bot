package adapter

import (
	coremodel "github.com/wallacehenriquesilva/slack-assets-bot/internal/core/model"
	"github.com/wallacehenriquesilva/slack-assets-bot/internal/core/port/in"
	extmodel "github.com/wallacehenriquesilva/slack-assets-bot/internal/model"
	"github.com/wallacehenriquesilva/slack-assets-bot/internal/service"
)

type SlackAdapter struct {
	slackService service.SlackClient
}

func NewSlackAdapter(slackService service.SlackClient) in.MessageSystem {
	return &SlackAdapter{
		slackService: slackService,
	}
}

func (slackAdapter *SlackAdapter) PublishMessage(message coremodel.Message) error {
	var messageColor extmodel.SlackMessageColor
	if message.Style == coremodel.SuccessMessage {
		messageColor = extmodel.Success
	} else {
		messageColor = extmodel.Error
	}
	return slackAdapter.slackService.PublishMessage(message.Title, message.Message, messageColor)
}

func (slackAdapter *SlackAdapter) DownloadFile(file coremodel.MessageFile) (string, error) {
	slackFile := extmodel.SlackFile{
		Url:      file.Url,
		FileType: file.Extension,
	}
	return slackAdapter.slackService.DownloadFile(slackFile)
}
