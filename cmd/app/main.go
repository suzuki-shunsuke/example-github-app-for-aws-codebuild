package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/example-github-app-for-aws-codebuild/pkg/handler"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.Info("start program")
	if err := core(); err != nil {
		logrus.Fatal(err)
	}
}

func setLogLevel() error {
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			return fmt.Errorf("parse LOG_LEVEL (%s): %w", logLevel, err)
		}
		logrus.SetLevel(lvl)
	}
	return nil
}

func core() error {
	if err := setLogLevel(); err != nil {
		return fmt.Errorf("set a log level: %w", err)
	}
	ctx := context.Background()
	logrus.Debug("start handler")
	return handler.Start(ctx) //nolint:wrapcheck
}
