package in

import "github.com/wallacehenriquesilva/slack-assets-bot/internal/core/model"

type MessageSystem interface {
	PublishMessage(message model.Message) error
	DownloadFile(file model.MessageFile) (string, error)
}
