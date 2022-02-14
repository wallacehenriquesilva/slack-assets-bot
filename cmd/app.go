package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/wallacehenriquesilva/slack-assets-bot/internal/adapter"
	coreservice "github.com/wallacehenriquesilva/slack-assets-bot/internal/core/service"
	extservice "github.com/wallacehenriquesilva/slack-assets-bot/internal/service"
	"log"
	"os"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("error to load env file")
	}

	token := os.Getenv("SLACK_AUTH_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")
	channelID := os.Getenv("SLACK_CHANNEL_ID")
	githubToken := os.Getenv("GITHUB_TOKEN")
	owner := os.Getenv("GITHUB_OWNER")
	repository := os.Getenv("GITHUB_REPOSITORY")
	authorName := os.Getenv("GITHUB_AUTHOR_NAME")
	authorEmail := os.Getenv("GITHUB_AUTHOR_EMAIL")

	slackService := extservice.NewSlackClient(token, appToken, channelID)
	githubService := extservice.NewGithubClient(githubToken, owner, repository, authorName, authorEmail)

	slackAdapter := adapter.NewSlackAdapter(slackService)
	githubAdapter := adapter.NewGithubAdapter(githubService)

	baseBranch := "main"
	commitMessage := "bot: add new assets"
	prTitle := ":robot: New assets"
	prDescription := `
## Description
- Adds the new assests using the slack bot.
`

	assetService := coreservice.NewAssetService(slackAdapter, githubAdapter, baseBranch, commitMessage, prTitle, prDescription)
	assetAdapter := adapter.NewAssetAdapter(assetService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = slackService.StartSocket(ctx, assetAdapter.Process)
	if err != nil {
		log.Fatalln("error to start the socket")
	}

}
