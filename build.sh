#!/usr/bin/env bash

set -eu
set -o pipefail

echo "HELLO: $HELLO"

# Install go-github-apps to generate GitHub App's token
curl -sSLf https://raw.githubusercontent.com/nabeken/go-github-apps/master/install-via-release.sh | bash -s -- -v v0.0.3
cp go-github-apps /usr/local/bin
eval $(go-github-apps -export -app-id "$CODEBUILDER_GITHUB_APP_APP_ID" -inst-id "$CODEBUILDER_GITHUB_APP_INSTALLATION_ID")

curl \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  "https://api.github.com/repos/$CODEBUILDER_REPO_OWNER/$CODEBUILDER_REPO_NAME/pulls/$CODEBUILDER_PR_NUMBER"
