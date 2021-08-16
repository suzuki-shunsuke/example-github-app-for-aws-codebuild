package handler

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/sirupsen/logrus"
)

func (handler *Handler) startBuilds(ctx context.Context, startBuildInputs []*codebuild.StartBuildInput) []*StartBuildError {
	var startBuildErrors []*StartBuildError
	for _, input := range startBuildInputs {
		input := input
		if err := handler.startBuild(ctx, input); err != nil {
			logrus.WithError(err).Error("start a build")
			startBuildErrors = append(startBuildErrors, &StartBuildError{
				StatusContext: aws.StringValue(input.BuildStatusConfigOverride.Context),
				Error:         err,
			})
		}
	}
	return startBuildErrors
}

func (handler *Handler) startBuild(ctx context.Context, input *codebuild.StartBuildInput) error {
	cb := handler.CodeBuild
	if _, err := cb.StartBuildWithContext(ctx, input); err != nil {
		return err //nolint:wrapcheck
	}
	return nil
}
