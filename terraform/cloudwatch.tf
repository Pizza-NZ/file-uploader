
# Creates a CloudWatch Log Group for the application.
resource "aws_cloudwatch_log_group" "main" {
  name = "/ecs/${var.app_name}"

  tags = {
    Name = "${var.app_name}-logs"
  }
}
