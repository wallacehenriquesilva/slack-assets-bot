package adapter

import (
	"encoding/json"
	"fmt"
	"github.com/slack-go/slack/slackevents"
	coremodel "github.com/wallacehenriquesilva/slack-assets-bot/internal/core/model"
	coreservice "github.com/wallacehenriquesilva/slack-assets-bot/internal/core/service"
	extmodel "github.com/wallacehenriquesilva/slack-assets-bot/internal/model"
)

type AssetAdapter struct {
	assetService coreservice.AssetSetvice
}

var (
	MaxNumberofFilesError = fmt.Errorf("invalid number of files. Only one file is allowed")
	MinNumberofFilesError = fmt.Errorf("invalid number of files. One file is required")
)

func NewAssetAdapter(assetService coreservice.AssetSetvice) *AssetAdapter {
	return &AssetAdapter{
		assetService: assetService,
	}
}

func (assetAdapter *AssetAdapter) Process(event slackevents.EventsAPIEvent) error {
	bytes, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}

	var slackEvent extmodel.SlackEvent

	err = json.Unmarshal(bytes, &slackEvent)
	if err != nil {
		return err
	}

	if len(slackEvent.Event.Files) > 1 {
		_ = assetAdapter.assetService.SendErrorMessage(MaxNumberofFilesError)
		return MaxNumberofFilesError
	} else if len(slackEvent.Event.Files) == 0 {
		_ = assetAdapter.assetService.SendErrorMessage(MinNumberofFilesError)
		return MinNumberofFilesError
	}

	file := slackEvent.Event.Files[0]

	coreFile := coremodel.AssetFile{
		Url:       file.Url,
		Extension: file.FileType,
	}
	return assetAdapter.assetService.Process(coreFile)
}
