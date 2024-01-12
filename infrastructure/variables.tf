variable "cloudflare_email" {
  description = "Cloudflare API email"
  type        = string
  sensitive   = false
  nullable    = false
}

variable "cloudflare_api_key" {
  description = "Cloudflare API key"
  type        = string
  sensitive   = true
  nullable    = false
}

variable "domain_name" {
  type     = string
  nullable = false
  default  = "git-top-repos.net"
}

variable "r2_buckets_name_base" {
  description = "Base name for Cloudflare R2 buckets"
  type        = string
  nullable    = false
}

variable "forward_all_from_domain_to_email" {
  description = "Email to forward all email messages received to the domain to"
  type        = string
  nullable    = false
}

variable "monitoring_warning_emails_from" {
  description = "Email to send monitoring warning emails from"
  type        = string
  nullable    = false
}

variable "monitoring_warning_emails_to" {
  description = "Email to send monitoring warning emails to"
  type        = string
  nullable    = false
}
