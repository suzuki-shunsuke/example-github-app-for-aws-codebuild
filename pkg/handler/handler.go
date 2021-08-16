package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/google/go-github/v38/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-github-app-for-aws-codebuild/ghapputil"
	"github.com/suzuki-shunsuke/go-template-unmarshaler/text"
	"gopkg.in/yaml.v2"
)

func Start(ctx context.Context) error {
	appID, err := strconv.ParseInt(os.Getenv("GITHUB_APP_ID"), 10, 64) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("GITHUB_APP_ID is invalid: %w", err)
	}
	sess := session.Must(session.NewSession())
	handler := &Handler{
		GitHubAppID: appID,
		CodeBuild:   codebuild.New(sess),
	}
	secret := &ghapputil.Secret{}
	if err := ghapputil.ReadSecretFromSecretsManager(ctx, secretsmanager.New(sess), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(os.Getenv("SECRET_ID")),
	}, secret); err != nil {
		return err //nolint:wrapcheck
	}
	handler.Secret = secret

	cfg := &Config{}
	if err := yaml.Unmarshal([]byte(os.Getenv("CONFIG")), cfg); err != nil {
		return fmt.Errorf("parse config as YAML: %w", err)
	}
	handler.Config = cfg

	lambda.StartWithContext(ctx, handler.Start)
	return nil
}

type Handler struct {
	CodeBuild   ghapputil.CodeBuild
	GitHubAppID int64
	Secret      *ghapputil.Secret
	Config      *Config
}

type Config struct {
	Messages map[string]*text.Template
}

type StartBuildErrors struct {
	SHA    string
	Errors []*StartBuildError
}

type StartBuildError struct {
	StatusContext string
	Error         error
}

// Start is the Lambda Function's endpoint.
func (handler *Handler) Start(ctx context.Context, event *ghapputil.Event) (*ghapputil.Response, error) { //nolint:funlen,cyclop
	if err := ghapputil.ValidateSignature(event, []byte(handler.Secret.WebhookSecret)); err != nil {
		logrus.WithError(err).Debug("validate webhook with signature")
		return ghapputil.NewResponse(http.StatusBadRequest, "invalid webhook"), nil
	}
	body, err := ghapputil.ParseWebHook(event.Headers.Event, event.Body)
	if err != nil {
		logrus.WithError(err).Debug("parse webhook request body")
		return ghapputil.NewResponse(http.StatusBadRequest, "the request body is invalid as GitHub Webhook"), nil
	}

	prEvent, ok := body.(*github.PullRequestEvent)
	if !ok {
		logrus.WithError(err).Debug("the event must be pull request")
		return ghapputil.NewResponse(http.StatusBadRequest, "the event must be pull request"), nil
	}

	pr := prEvent.GetPullRequest()
	repo := prEvent.Repo

	if !ghapputil.ExcludePREventByAction(prEvent) {
		logrus.WithError(err).Debug("this aciton is ignored: " + prEvent.GetAction())
		return ghapputil.NewResponse(http.StatusBadRequest, "this aciton is ignored: "+prEvent.GetAction()), nil
	}

	instID := prEvent.GetInstallation().GetID()
	logrus.WithFields(logrus.Fields{
		"app_id":          handler.GitHubAppID,
		"installation_id": instID,
	}).Debug("start creating github client")
	ghClient, err := ghapputil.NewGitHubClient(handler.GitHubAppID, prEvent, []byte(handler.Secret.GitHubAppPrivateKey))
	if err != nil {
		logrus.WithError(err).Error("create a GitHub client")
		return ghapputil.NewResponse(http.StatusInternalServerError, "Internal Server Error"), nil
	}

	startBuildInputs, err := handler.getBuilds(ctx, prEvent, ghClient, handler.GitHubAppID, instID)
	if err != nil {
		if _, _, e := ghClient.Issues.CreateComment(ctx, repo.GetOwner().GetLogin(), repo.GetName(), pr.GetNumber(), &github.IssueComment{
			Body: github.String(""),
		}); e != nil {
			logrus.WithError(e).Error("create a pull request comment by GitHub API")
			return ghapputil.NewResponse(http.StatusInternalServerError, "failed to start builds"), err
		}
	}
	logrus.WithFields(logrus.Fields{
		"count": len(startBuildInputs),
	}).Info("start builds")
	startBuildErrors := handler.startBuilds(ctx, startBuildInputs)
	if len(startBuildErrors) != 0 {
		sha := prEvent.GetAfter()
		if prEvent.PullRequest.GetMerged() {
			sha = prEvent.PullRequest.GetMergeCommitSHA()
		}
		s, err := handler.Config.Messages["start_build"].Execute(StartBuildErrors{
			SHA:    sha,
			Errors: startBuildErrors,
		})
		if err != nil {
			logrus.WithError(err).Error("render template")
			return ghapputil.NewResponse(http.StatusInternalServerError, "failed to start builds and render template"), err
		}
		if _, _, e := ghClient.Issues.CreateComment(ctx, repo.GetOwner().GetLogin(), repo.GetName(), pr.GetNumber(), &github.IssueComment{
			Body: github.String(s),
		}); e != nil {
			logrus.WithError(e).Error("create a pull request comment by GitHub API")
			return ghapputil.NewResponse(http.StatusInternalServerError, "failed to start builds"), err
		}
	}
	return nil, nil
}
