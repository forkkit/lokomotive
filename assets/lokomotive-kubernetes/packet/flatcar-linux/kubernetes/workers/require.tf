# Terraform version and plugin versions

terraform {
  required_version = ">= 0.12.0"

  required_providers {
    ct       = "= 0.5.0"
    local    = "~> 1.2"
    template = "~> 2.1"
    tls      = "~> 2.0"
    packet   = "~> 2.7.3"
  }
}
