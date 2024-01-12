# This file deploy a lambda that check every day if the /metadata file has been pushed by the uploader,
# to monitor if the website is still being updated and current.

provider "aws" {}

resource "aws_cloudwatch_event_rule" "schedule" {
  name                = "checkGitTopReposForCurrentData"
  description         = "Check ${var.domain_name} for current data"
  schedule_expression = "rate(1 day)"
  depends_on = [
    aws_lambda_function.check_last_modified
  ]
}

resource "aws_cloudwatch_event_target" "schedule_lambda" {
  rule = aws_cloudwatch_event_rule.schedule.name
  arn  = aws_lambda_function.check_last_modified.arn
}

resource "aws_lambda_permission" "allow_events_bridge_to_run_lambda" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.check_last_modified.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.schedule.arn
}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
    actions = ["sts:AssumeRole"]
  }
}


data "aws_iam_policy_document" "ses_send_email_access" {
  statement {
    effect    = "Allow"
    actions   = ["ses:SendEmail"]
    resources = ["*"]
    condition {
      test     = "StringLike"
      variable = "ses:FromAddress"
      values   = [var.monitoring_warning_emails_from]
    }
    condition {
      test     = "ForAllValues:StringLike"
      variable = "ses:Recipients"
      values   = [var.monitoring_warning_emails_to]
    }
  }
}

resource "aws_iam_role" "lambda_role" {
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
  inline_policy {
    name   = "allow-sending-emails"
    policy = data.aws_iam_policy_document.ses_send_email_access.json
  }

  managed_policy_arns  = ["arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"]
  max_session_duration = 3600
  name                 = "checkGitTopReposLastModifiedLambdaRole"
  path                 = "/service-role/"
}

data "archive_file" "lambda" {
  type        = "zip"
  source_file = "monitoring_by_aws_lambda.py"
  output_path = "${path.root}/.generated_archives/aws_lambda_function_check_last_modified.zip"
}

resource "aws_lambda_function" "check_last_modified" {
  function_name                  = "checkGitTopReposLastModifiedFunction"
  runtime                        = "python3.12"
  handler                        = "monitoring_by_aws_lambda.lambda_handler"
  memory_size                    = 128
  reserved_concurrent_executions = -1
  architectures                  = ["arm64"]
  role                           = aws_iam_role.lambda_role.arn
  filename                       = data.archive_file.lambda.output_path
  source_code_hash               = data.archive_file.lambda.output_base64sha256

  environment {
    variables = {
      SENDER_EMAIL_ADDRESS = var.monitoring_warning_emails_from
      TARGET_EMAIL_ADDRESS = var.monitoring_warning_emails_to
      METADATA_URL         = "https://data.${var.domain_name}/metadata"
    }
  }

  timeout = 60
}


