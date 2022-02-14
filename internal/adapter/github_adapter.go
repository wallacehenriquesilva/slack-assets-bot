package adapter

import (
	"github.com/wallacehenriquesilva/slack-assets-bot/internal/core/model"
	"github.com/wallacehenriquesilva/slack-assets-bot/internal/core/port/out"
	extmodel "github.com/wallacehenriquesilva/slack-assets-bot/internal/model"
	"github.com/wallacehenriquesilva/slack-assets-bot/internal/service"
)

type GithubAdapter struct {
	githubService service.GithubClient
}

func NewGithubAdapter(githubService service.GithubClient) out.VersionControlSystem {
	return &GithubAdapter{
		githubService: githubService,
	}
}

func (githubAdapter *GithubAdapter) CreateCommit(commitBranch, baseBranch, message string, sourceFiles []model.VCSFile) error {
	githubFiles := make([]extmodel.GithubFile, 0, len(sourceFiles)+1)

	for _, file := range sourceFiles {
		githubFile := extmodel.GithubFile{
			LocalPath:  file.LocalPath,
			RemotePath: file.RemotePath,
		}
		githubFiles = append(githubFiles, githubFile)
	}

	return githubAdapter.githubService.CreateCommit(commitBranch, baseBranch, message, githubFiles)
}

func (githubAdapter *GithubAdapter) CreatePullRequest(headBranch, baseBranch, title, description string) (string, error) {
	return githubAdapter.githubService.CreatePullRequest(headBranch, baseBranch, title, description)
}
