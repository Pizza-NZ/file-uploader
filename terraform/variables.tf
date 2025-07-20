variable "aws_region" {
  description = "The AWS region to deploy resources in."
  type        = string
  default     = "ap-southeast-2"
}

variable "app_name" {
  description = "The name of the application."
  type        = string
  default     = "file-uploader"
}
