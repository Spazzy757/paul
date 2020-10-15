package test

import (
	"context"
	"io"

	paulclient "github.com/Spazzy757/paul/pkg/client"
	"github.com/google/go-github/v32/github"
)

//MockPullRequestService
type MockPullRequestService struct {
	CreateReviewResp *github.PullRequestReview
	Err              error
}

func (m *MockPullRequestService) CreateReview(
	ctx context.Context,
	owner, repo string,
	number int,
	review *github.PullRequestReviewRequest,
) (*github.PullRequestReview, *github.Response, error) {
	return m.CreateReviewResp, nil, m.Err
}

//MockGitService
type MockGitService struct {
	DeleteRefResp *github.Response
	Err           error
}

func (m *MockGitService) DeleteRef(
	ctx context.Context,
	owner, repo, ref string,
) (*github.Response, error) {
	return m.DeleteRefResp, m.Err
}

//MockRepoService
type MockRepoService struct {
	DownloadContentsResp io.ReadCloser
	Err                  error
}

func (m *MockRepoService) DownloadContents(
	ctx context.Context,
	owner, repo, filepath string,
	opt *github.RepositoryContentGetOptions,
) (io.ReadCloser, error) {
	return m.DownloadContentsResp, m.Err
}

// MockIssueService
type MockIssueService struct {
	CreateCommentResp       *github.IssueComment
	AddLabelsToIssueResp    []*github.Label
	RemoveLabelForIssueResp *github.Response
	Err                     error
}

func (m *MockIssueService) CreateComment(
	ctx context.Context,
	owner, repo string,
	number int,
	review *github.IssueComment,
) (*github.IssueComment, *github.Response, error) {
	return m.CreateCommentResp, nil, m.Err
}

func (m *MockIssueService) AddLabelsToIssue(
	ctx context.Context,
	owner, repo string,
	number int,
	labels []string,
) ([]*github.Label, *github.Response, error) {
	return m.AddLabelsToIssueResp, nil, m.Err
}

func (m *MockIssueService) RemoveLabelForIssue(
	ctx context.Context,
	owner, repo string,
	number int,
	label string,
) (*github.Response, error) {
	return m.RemoveLabelForIssueResp, m.Err
}

func GetMockClient() *paulclient.GithubClient {
	ctx := context.Background()
	return &paulclient.GithubClient{
		Ctx:                ctx,
		GitService:         &MockGitService{},
		RepoService:        &MockRepoService{},
		PullRequestService: &MockPullRequestService{},
		IssueService:       &MockIssueService{},
	}
}
