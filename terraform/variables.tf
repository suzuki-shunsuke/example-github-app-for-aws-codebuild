variable "region" {
  type = string
}

variable "zip_path" {
  type        = string
  description = ""
  default     = "codebuilder_linux_amd64.zip"
}

variable "function_name" {
  type        = string
  description = "Lambda Function Name"
  default     = "codebuilder-example-github-app-for-aws-codebuild"
}

variable "lambda_role_name" {
  type        = string
  description = ""
  default     = "lambda-codebuilder-example-github-app-for-aws-codebuild"
}

variable "repo_full_name" {
  type        = string
  description = "source repository full name"
}

variable "codebuild_role_name" {
  type        = string
  description = "IAM Role Name for CodeBuild"
  default     = "codebuild-example-github-app-for-aws-codebuild"
}

variable "api_gateway_name" {
  type        = string
  description = "API Gateway name"
  default     = "example-github-app-for-aws-codebuild"
}

variable "project_name" {
  type        = string
  description = "CodeBuild Project name"
  default     = "example-github-app-for-aws-codebuild"
}

variable "secret_id" {
  type    = string
  default = "lambda-codebuilder-example-github-app-for-aws-codebuild"
}

variable "webhook_secret" {
  type      = string
  sensitive = true
}

variable "github_app_id" {
  type = number
}
