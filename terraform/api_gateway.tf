resource "aws_apigatewayv2_api" "main" {
  name          = var.api_gateway_name
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.main.id
  name        = "$default"
  auto_deploy = true

  default_route_settings {
    detailed_metrics_enabled = true
  }

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.apigateway_default.arn
    format = jsonencode(
      {
        httpMethod     = "$context.httpMethod"
        ip             = "$context.identity.sourceIp"
        protocol       = "$context.protocol"
        requestId      = "$context.requestId"
        requestTime    = "$context.requestTime"
        responseLength = "$context.responseLength"
        routeKey       = "$context.routeKey"
        status         = "$context.status"
      }
    )
  }

  route_settings {
    route_key                = aws_apigatewayv2_route.main.route_key
    detailed_metrics_enabled = true
  }
}

resource "aws_cloudwatch_log_group" "apigateway_default" {
  name              = "/aws/apigateway/${var.api_gateway_name}/default"
  retention_in_days = 7
}

# per function

resource "aws_apigatewayv2_deployment" "main" {
  api_id      = aws_apigatewayv2_route.main.api_id
  description = "example-github-app-for-aws-codebuild"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_apigatewayv2_integration" "main" {
  api_id                 = aws_apigatewayv2_api.main.id
  integration_type       = "AWS_PROXY"
  payload_format_version = "2.0"

  description        = "example-github-app-for-aws-codebuild"
  integration_method = "POST"
  integration_uri    = aws_lambda_function.main.invoke_arn
}

resource "aws_apigatewayv2_route" "main" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "POST /repos/example-github-app-for-aws-codebuild"

  target = "integrations/${aws_apigatewayv2_integration.main.id}"
}

resource "aws_lambda_permission" "lambda_permission" {
  statement_id  = "AllowCodebuilderInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.main.function_name
  principal     = "apigateway.amazonaws.com"

  # The /*/*/* part allows invocation from any stage, method and resource path
  # within API Gateway HTTP API.
  source_arn = "${aws_apigatewayv2_api.main.execution_arn}/*/*/repos/example-github-app-for-aws-codebuild"
}
