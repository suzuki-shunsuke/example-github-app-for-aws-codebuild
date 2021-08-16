resource "aws_secretsmanager_secret" "main" {
  name = var.secret_id
}

resource "aws_secretsmanager_secret_version" "codebuilder-example-github-app-for-aws-codebuild" {
  secret_id = aws_secretsmanager_secret.main.id
  secret_string = jsonencode({
    "webhook_secret" : var.webhook_secret,
    "github_app_private_key" : file("${path.module}/github-app-private-key.pem"),
  })
}
