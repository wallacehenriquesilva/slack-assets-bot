package out

import "github.com/wallacehenriquesilva/slack-assets-bot/internal/core/model"

type VersionControlSystem interface {
	CreateCommit(commitBranch, baseBranch, message string, sourceFiles []model.VCSFile) error
	CreatePullRequest(headBranch, baseBranch, title, description string) (string, error)
}
