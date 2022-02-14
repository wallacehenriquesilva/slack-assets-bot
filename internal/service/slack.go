package service

import (
	"context"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/wallacehenriquesilva/slack-assets-bot/internal/model"
	"github.com/wallacehenriquesilva/slack-assets-bot/internal/util/requestutil"
	"log"
	"os"
	"time"
)

type SlackClient interface {
	StartSocket(ctx context.Context, processFunction ProcessFunction) error
	PublishMessage(pretext, text string, color model.SlackMessageColor) error
	DownloadFile(file model.SlackFile) (string, error)
}

type ProcessFunction func(slackevents.EventsAPIEvent) error

type SlackClientImpl struct {
	authToken string
	channelID string
	client    *slack.Client
}

func NewSlackClient(authToken, appToken, channelID string) SlackClient {
	client := slack.New(authToken, slack.OptionDebug(false), slack.OptionAppLevelToken(appToken))

	return &SlackClientImpl{
		authToken: authToken,
		channelID: channelID,
		client:    client,
	}
}

func (slackClient *SlackClientImpl) StartSocket(ctx context.Context, processFunction ProcessFunction) error {
	socketClient := socketmode.New(
		slackClient.client,
		socketmode.OptionDebug(false),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client, process ProcessFunction) {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-socketClient.Events:
				if event.Type != socketmode.EventTypeEventsAPI {
					continue
				}
				socketClient.Ack(*event.Request)
				eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Printf("Could not type cast the event to the EventsAPIEvent: %v\n", event)
					continue
				}

				err := process(eventsAPIEvent)
				if err != nil {
					log.Printf("error to process event: %v\n", err)
				}
			}
		}
	}(ctx, slackClient.client, socketClient, processFunction)

	if err := socketClient.Run(); err != nil {
		return err
	}

	return nil

}

func (slackClient *SlackClientImpl) PublishMessage(pretext, text string, color model.SlackMessageColor) error {
	attachment := slack.Attachment{
		Pretext: pretext,
		Text:    text,
		Color:   string(color),
		Fields: []slack.AttachmentField{
			{
				Title: "Date",
				Value: time.Now().String(),
			},
		},
	}

	_, _, err := slackClient.client.PostMessage(
		slackClient.channelID,
		slack.MsgOptionAttachments(attachment),
	)
	if err != nil {
		return err
	}

	return nil
}

func (slackClient *SlackClientImpl) DownloadFile(file model.SlackFile) (string, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + slackClient.authToken,
	}
	return requestutil.DownloadFile(file.Url, headers, file.FileType)
}
