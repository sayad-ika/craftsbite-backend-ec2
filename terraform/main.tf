terraform {
  backend "s3" {
    bucket = "trainee-2026-sayad-craftsbite-tfstate"
    key    = "craftsbite/terraform.tfstate"
    region = "ap-southeast-1"
  }
}

provider "aws" {
  region = var.aws_region
}
