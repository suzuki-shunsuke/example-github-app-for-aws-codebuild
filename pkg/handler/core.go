package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/google/go-github/v38/github"
	"github.com/suzuki-shunsuke/go-github-app-for-aws-codebuild/ghapputil"
)

func (handler *Handler) getBuilds(ctx context.Context, prEvent *github.PullRequestEvent, ghClient *github.Client, appID, instID int64) ([]*codebuild.StartBuildInput, error) {
	pr := prEvent.PullRequest
	repo := prEvent.Repo
	files, err := ghapputil.GetPRFiles(ctx, ghClient, repo.Owner.GetLogin(), repo.GetName(), pr.GetNumber(), pr.GetChangedFiles())
	if err != nil {
		return nil, fmt.Errorf("get pull request files: %w", err)
	}
	services := map[string]struct{}{
		"foo": {},
	}
	changedServices := map[string]struct{}{}
	for svc := range services {
		for _, file := range files {
			if strings.HasPrefix(file.GetFilename(), svc+"/") {
				changedServices[svc] = struct{}{}
				break
			}
		}
	}

	allChangedFileNames := make([]string, 0, len(files))
	for _, file := range files {
		allChangedFileNames = append(allChangedFileNames, file.GetFilename())
		if a := file.GetPreviousFilename(); a != "" {
			allChangedFileNames = append(allChangedFileNames, a)
		}
	}

	sourceVersion := aws.String(ghapputil.GetSourceVersion(prEvent, false))

	inputs := make([]*codebuild.StartBuildInput, 0, len(changedServices))
	envs := append(append(ghapputil.GetPREnv(prEvent), ghapputil.GetGitHubAppEnv(appID, instID)...), &codebuild.EnvironmentVariable{
		Name:  aws.String("CODEBUILDER_PR_CHANGED_FILES"),
		Value: aws.String(strings.Join(allChangedFileNames, "\n")),
	})
	for svc := range changedServices {
		inputs = append(inputs, &codebuild.StartBuildInput{
			ProjectName: aws.String("example-github-app-for-aws-codebuild"),
			BuildStatusConfigOverride: &codebuild.BuildStatusConfig{
				Context: aws.String("hello (" + svc + ")"),
			},
			SourceVersion: sourceVersion,
			EnvironmentVariablesOverride: append(envs, &codebuild.EnvironmentVariable{
				Name:  aws.String("SERVICE"),
				Value: aws.String(svc),
			}),
		})
	}
	return inputs, nil
}
