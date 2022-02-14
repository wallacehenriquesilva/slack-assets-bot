package service

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/wallacehenriquesilva/slack-assets-bot/internal/core/model"
	"github.com/wallacehenriquesilva/slack-assets-bot/internal/core/port/in"
	"github.com/wallacehenriquesilva/slack-assets-bot/internal/core/port/out"
	"github.com/wallacehenriquesilva/slack-assets-bot/internal/util/fileutil"
	"strings"
)

type AssetSetvice interface {
	Process(assetFile model.AssetFile) error
	SendErrorMessage(er error) error
}

var (
	InvalidURLError           = fmt.Errorf("invalid url to file. The url is required")
	InvalidFileExtensionError = fmt.Errorf("invalid file format. The format allowed is only zip")
)

type AssetSetviceImpl struct {
	vcsClient     out.VersionControlSystem
	messageClient in.MessageSystem
	baseBranch    string
	commitMessage string
	prTitle       string
	prDescription string
}

func NewAssetService(messageClient in.MessageSystem, vcsClient out.VersionControlSystem, baseBranch, commitMessage,
	prTitle, prDescription string) AssetSetvice {
	return &AssetSetviceImpl{
		messageClient: messageClient,
		vcsClient:     vcsClient,
		baseBranch:    baseBranch,
		commitMessage: commitMessage,
		prTitle:       prTitle,
		prDescription: prDescription,
	}
}

func (assetService *AssetSetviceImpl) Process(assetFile model.AssetFile) error {
	if err := assetService.validateAssetFile(assetFile); err != nil {
		_ = assetService.SendErrorMessage(err)
		return err
	}

	messageFile := model.MessageFile{
		Url:       assetFile.Url,
		Extension: assetFile.Extension,
	}

	file, err := assetService.messageClient.DownloadFile(messageFile)
	if err != nil {
		_ = assetService.SendErrorMessage(err)
		return err
	}
	defer fileutil.DeleteFiles(file)

	unzipedFiles, err := fileutil.UnzipFiles(file, ignoreFile)
	if err != nil {
		_ = assetService.SendErrorMessage(err)
		return err
	}
	defer deleteUnzipedFiles(unzipedFiles)

	files := unzipedToVcs(unzipedFiles)

	branchName, err := generateBranchName()
	if err != nil {
		_ = assetService.SendErrorMessage(err)
		return err
	}

	err = assetService.vcsClient.CreateCommit(branchName, assetService.baseBranch, assetService.commitMessage, files)
	if err != nil {
		_ = assetService.SendErrorMessage(err)
		return err
	}

	prUrl, err := assetService.vcsClient.CreatePullRequest(
		branchName,
		assetService.baseBranch,
		assetService.prTitle,
		assetService.prDescription,
	)
	if err != nil {
		_ = assetService.SendErrorMessage(err)
		return err
	}

	err = assetService.sendSuccessMessage(prUrl)
	if err != nil {
		return err
	}

	return nil
}

func (assetService *AssetSetviceImpl) validateAssetFile(assetFile model.AssetFile) error {
	if assetFile.Url == "" {
		return InvalidURLError
	}

	if assetFile.Extension != "zip" {
		return InvalidFileExtensionError
	}

	return nil
}

func (assetService *AssetSetviceImpl) SendErrorMessage(er error) error {
	err := assetService.sendMessage("Error to process the asset", er.Error(), model.ErrorMessage)
	if err != nil {
		return err
	}
	return nil
}

func (assetService *AssetSetviceImpl) sendSuccessMessage(prUrl string) error {
	message := "You can see the PR opened in :arrow_right: " + prUrl
	err := assetService.sendMessage("Asset processed with success", message, model.SuccessMessage)
	if err != nil {
		return err
	}
	return nil
}

func (assetService *AssetSetviceImpl) sendMessage(title, message string, style model.MessageStyle) error {
	messageContent := model.Message{
		Title:   title,
		Message: message,
		Style:   style,
	}

	err := assetService.messageClient.PublishMessage(messageContent)
	if err != nil {
		return err
	}
	return nil
}

func generateBranchName() (string, error) {
	u4, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	return "asset-" + u4.String(), nil
}

func unzipedToVcs(unzipedFiles []fileutil.File) []model.VCSFile {
	files := make([]model.VCSFile, 0, len(unzipedFiles)+1)

	for _, file := range unzipedFiles {

		fl := model.VCSFile{
			LocalPath:  file.LocalPath,
			RemotePath: file.RemotePath,
		}

		files = append(files, fl)
	}

	return files
}

func ignoreFile(file string) bool {
	return strings.HasPrefix(file, "_")
}

func deleteUnzipedFiles(unzipedFiles []fileutil.File) {
	if len(unzipedFiles) > 0 {
		filePath := unzipedFiles[0].LocalPath
		zipFolder := strings.Split(filePath, "/")[0]
		_ = fileutil.DeleteFiles(zipFolder)
	}
}
