package service

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/wallacehenriquesilva/slack-assets-bot/internal/model"
	"golang.org/x/oauth2"
	"io/ioutil"
	"time"
)

type GithubClient interface {
	CreateCommit(commitBranch, baseBranch, message string, sourceFiles []model.GithubFile) error
	CreatePullRequest(headBranch, baseBranch, title, description string) (string, error)
}

var (
	InvalidPrTitleError    = fmt.Errorf("the pr title is required")
	InvalidHeadBranchError = fmt.Errorf("the pr head branch is required")
	InvalidBaseBranchError = fmt.Errorf("the pr base branch is required")
	SameBranchError        = fmt.Errorf("the pr base and head branch are the same, it is not allowed")
	InvalidLocalPathError  = fmt.Errorf("invalid local path. The local path can't be empty")
	InvalidRemotePathError = fmt.Errorf("invalid remote path. The remote path can't be empty")
)

type GithubClientImpl struct {
	client      *github.Client
	owner       string
	repository  string
	authorName  string
	authorEmail string
}

func NewGithubClient(accessToken, owner, repository, authorName, authorEmail string) GithubClient {
	client := getGithubClient(context.Background(), accessToken)
	return &GithubClientImpl{
		client:      client,
		owner:       owner,
		repository:  repository,
		authorName:  authorName,
		authorEmail: authorEmail,
	}
}

func getGithubClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)

}

func (githubClient *GithubClientImpl) CreateCommit(commitBranch, baseBranch, message string, sourceFiles []model.GithubFile) error {
	ctx := context.Background()

	ref, err := githubClient.getRef(ctx, githubClient.client, commitBranch, baseBranch)
	if err != nil {
		return err
	}

	tree, err := githubClient.getTree(ctx, githubClient.client, ref, sourceFiles)
	if err != nil {
		return err
	}

	parent, _, err := githubClient.client.Repositories.GetCommit(
		context.Background(),
		githubClient.owner,
		githubClient.repository,
		*ref.Object.SHA,
	)
	if err != nil {
		return err
	}

	parent.Commit.SHA = parent.SHA

	date := time.Now()
	author := &github.CommitAuthor{Date: &date, Name: &githubClient.authorName, Email: &githubClient.authorEmail}
	commit := &github.Commit{Author: author, Message: &message, Tree: tree, Parents: []github.Commit{*parent.Commit}}
	newCommit, _, err := githubClient.client.Git.CreateCommit(ctx, githubClient.owner, githubClient.repository, commit)
	if err != nil {
		return err
	}

	ref.Object.SHA = newCommit.SHA
	_, _, err = githubClient.client.Git.UpdateRef(ctx, githubClient.owner, githubClient.repository, ref, false)
	return err
}

func (githubClient *GithubClientImpl) CreatePullRequest(headBranch, baseBranch, title, description string) (string, error) {
	if title == "" {
		return "", InvalidPrTitleError
	}

	if headBranch == "" {
		return "", InvalidHeadBranchError
	}

	commitBranch := fmt.Sprintf("%s:%s", githubClient.owner, headBranch)

	pullRequestPayload := &github.NewPullRequest{
		Title:               &title,
		Head:                &commitBranch,
		Base:                &baseBranch,
		Body:                &description,
		MaintainerCanModify: github.Bool(true),
	}

	pullRequest, _, err := githubClient.client.PullRequests.Create(context.Background(), githubClient.owner, githubClient.repository, pullRequestPayload)
	if err != nil {
		return "", err
	}

	return pullRequest.GetHTMLURL(), nil
}

// getRef returns the commit branch reference object if it exists or creates it
// from the base branch before returning it.
func (githubClient *GithubClientImpl) getRef(ctx context.Context, client *github.Client, commitBranch, baseBranch string) (ref *github.Reference, err error) {
	if ref, _, err = client.Git.GetRef(ctx, githubClient.owner, githubClient.repository, "refs/heads/"+commitBranch); err == nil {
		return ref, nil
	}

	if commitBranch == baseBranch {
		return nil, SameBranchError
	}

	if baseBranch == "" {
		return nil, InvalidBaseBranchError
	}

	var baseRef *github.Reference
	if baseRef, _, err = client.Git.GetRef(ctx, githubClient.owner, githubClient.repository, "refs/heads/"+baseBranch); err != nil {
		return nil, err
	}
	newRef := &github.Reference{Ref: github.String("refs/heads/" + commitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err = client.Git.CreateRef(ctx, githubClient.owner, githubClient.repository, newRef)
	return ref, err
}

// getTree generates the tree to commit based on the given files and the commit
// of the ref you got in getRef.
func (githubClient *GithubClientImpl) getTree(ctx context.Context, client *github.Client, ref *github.Reference, sourceFiles []model.GithubFile) (tree *github.Tree, err error) {
	var entries []github.TreeEntry

	for _, fileArg := range sourceFiles {
		file, content, err := getFileContent(fileArg)
		if err != nil {
			return nil, err
		}

		entries = append(entries, github.TreeEntry{Path: github.String(file), Type: github.String("blob"), Content: github.String(string(content)), Mode: github.String("100644")})
	}

	tree, _, err = client.Git.CreateTree(ctx, githubClient.owner, githubClient.repository, *ref.Object.SHA, entries)
	return tree, err
}

// getFileContent loads the local content of a file and return the target name
// of the file in the target repository and its contents.
func getFileContent(file model.GithubFile) (string, []byte, error) {
	if file.LocalPath == "" {
		return "", nil, InvalidLocalPathError
	}

	if file.RemotePath == "" {
		return "", nil, InvalidRemotePathError
	}

	bytes, err := ioutil.ReadFile(file.LocalPath)
	return file.RemotePath, bytes, err
}
