provider "cloudflare" {
  email   = var.cloudflare_email
  api_key = var.cloudflare_api_key
}

data "cloudflare_accounts" "this" {}

locals {
  cloudflare_account_id = data.cloudflare_accounts.this.accounts[0].id
}

data "cloudflare_zone" "this" {
  account_id = local.cloudflare_account_id
  name       = var.domain_name
}

resource "cloudflare_zone_settings_override" "this" {
  zone_id = data.cloudflare_zone.this.zone_id

  settings {
    always_use_https         = "on"
    automatic_https_rewrites = "on"
    browser_check            = "off"
    early_hints              = "on"
    security_level           = "essentially_off"
    ssl                      = "strict"
    zero_rtt                 = "on"
    security_header {
      enabled            = true
      include_subdomains = true
      max_age            = 31536000
      preload            = true
    }
  }
}

resource "cloudflare_zone_dnssec" "this" {
  zone_id = data.cloudflare_zone.this.zone_id
}

resource "cloudflare_r2_bucket" "rendered_api" {
  account_id = local.cloudflare_account_id
  name       = "${var.r2_buckets_name_base}-rendered-api"
  location   = "ENAM"
}

resource "cloudflare_r2_bucket" "database_backup" {
  account_id = local.cloudflare_account_id
  name       = "${var.r2_buckets_name_base}-database-backup"
  location   = "ENAM"
}

resource "cloudflare_pages_project" "this" {
  account_id        = local.cloudflare_account_id
  name              = "git-top-repos"
  production_branch = "master"

  source {
    type = "github"
    config {
      owner                         = "karolba"
      repo_name                     = "top-of-github"
      production_branch             = "master"
      deployments_enabled           = true
      production_deployment_enabled = true
      preview_branch_includes       = ["*"]
    }
  }
  build_config {
    build_command   = "npm run build-production"
    destination_dir = "/dist"
    root_dir        = "/frontend"
  }
  deployment_configs {
    production {
      compatibility_date = "2023-08-16"
      environment_variables = {
        "NODE_VERSION" = "20.2.0"
      }
      # TODO: It's not yet possible to set the build image version automatically.
      # See: https://github.com/cloudflare/terraform-provider-cloudflare/issues/2438
      # Until that is fixed, this step must be done manually
      #build_image_major_version = 2
    }
    preview {
      compatibility_date = "2023-08-16"
      environment_variables = {
        "NODE_VERSION" = "20.2.0"
      }
      # TODO: It's not yet possible to set the build image version automatically.
      # See: https://github.com/cloudflare/terraform-provider-cloudflare/issues/2438
      # Until that is fixed, this step must be done manually
      #build_image_major_version = 2
    }
  }
}

resource "cloudflare_pages_domain" "this" {
  account_id   = local.cloudflare_account_id
  domain       = var.domain_name
  project_name = cloudflare_pages_project.this.name
}

resource "cloudflare_record" "to_pages" {
  zone_id = data.cloudflare_zone.this.zone_id
  name    = "@"
  type    = "CNAME"
  value   = "${cloudflare_pages_project.this.name}.pages.dev."
  proxied = true
}

resource "cloudflare_record" "www_redirect" {
  zone_id = data.cloudflare_zone.this.zone_id
  name    = "www"
  type    = "CNAME"
  value   = var.domain_name
  proxied = true
}

resource "cloudflare_record" "spf" {
  zone_id = data.cloudflare_zone.this.zone_id
  name    = "@"
  type    = "TXT"
  value   = "v=spf1 include:_spf.mx.cloudflare.net -all"
}

resource "cloudflare_page_rule" "www_redirect" {
  zone_id = data.cloudflare_zone.this.zone_id
  target  = "www.${var.domain_name}/*"
  actions {
    forwarding_url {
      url         = "https://${var.domain_name}/$1"
      status_code = 302
    }
  }
}

resource "cloudflare_email_routing_settings" "this" {
  zone_id = data.cloudflare_zone.this.zone_id
  enabled = "true"
}

resource "cloudflare_email_routing_address" "this" {
  account_id = local.cloudflare_account_id
  email      = var.forward_all_from_domain_to_email
}

resource "cloudflare_email_routing_catch_all" "this" {
  zone_id = data.cloudflare_zone.this.zone_id
  enabled = true
  name    = "catch-all"

  matcher {
    type = "all"
  }

  action {
    type  = "forward"
    value = [cloudflare_email_routing_address.this.email]
  }
}

# TODO: it's not yet possible to set a custom domain for an R2 bucket via either the API
# or terraform.
# See: https://github.com/cloudflare/terraform-provider-cloudflare/issues/2537
# Until that is fixed, there's an extra manual step to be done after deploying this:
# point "data.${var.domain_name}" to the ${cloudflare_r2_bucket.rendered_api} bucket.
#
#resource "cloudflare_record" "data" {
#  zone_id = data.cloudflare_zone.this.zone_id
#  name    = "data"
#  type    = "R2"
#  value   = cloudflare_r2_bucket.rendered_api.id
#}

resource "cloudflare_ruleset" "transform_modify_response_headers_for_cors" {
  zone_id = data.cloudflare_zone.this.zone_id
  name    = "data-cors"
  kind    = "zone"
  phase   = "http_response_headers_transform"

  rules {
    description = "Set CORS headers for the data subdomain"

    action = "rewrite"
    action_parameters {
      headers {
        name      = "Access-Control-Allow-Headers"
        operation = "set"
        value     = "*"
      }
      headers {
        name      = "Access-Control-Allow-Methods"
        operation = "set"
        value     = "GET, HEAD"
      }
      headers {
        name      = "Access-Control-Allow-Origin"
        operation = "set"
        value     = "*"
      }
    }
    expression = "(http.host eq \"data.${var.domain_name}\")"
    enabled    = true
  }
}

# As this is pretty much a static website, disable all unnecessary "security" layers that might prevent
# the website from loading
resource "cloudflare_ruleset" "disable_all_firewall_steps" {
  zone_id = data.cloudflare_zone.this.zone_id
  kind    = "zone"
  name    = "default"
  phase   = "http_request_firewall_custom"

  rules {
    description = "skip all unneccessary security steps for a static website"
    action      = "skip"

    action_parameters {
      # disable everything that can be disabled
      phases   = ["http_ratelimit", "http_request_firewall_managed", "http_request_sbfm"]
      products = ["zoneLockdown", "uaBlock", "bic", "hot", "securityLevel", "rateLimit", "waf"]
      ruleset  = "current"
    }
    logging {
      enabled = true
    }

    # a bogus expression that's always true to match everything
    expression = "(http.request.method eq \"GET\") or (http.request.method ne \"GET\")"
    enabled    = true
  }
}

resource "cloudflare_managed_headers" "this" {
  zone_id = data.cloudflare_zone.this.zone_id
  managed_response_headers {
    id      = "add_security_headers"
    enabled = true
  }
}

# Cache known js and css assets that don't change:
locals {
  known_static_files_expression = <<-EOT
      (http.host eq "${var.domain_name}" and (
        starts_with(http.request.uri.path, "/external-libraries/")
        or (starts_with(http.request.uri.path, "/index-") and ends_with(http.request.uri.path, ".js"))))
    EOT
  cache_static_files_for        = 28 * 24 * 60 * 60 # a month
}
resource "cloudflare_ruleset" "cache_rules_for_static_assets" {
  zone_id = data.cloudflare_zone.this.zone_id
  kind    = "zone"
  name    = "cache-known-static-assets"
  phase   = "http_request_cache_settings"
  rules {
    description = "Cache static javascript and css files for longer than the default"
    action      = "set_cache_settings"
    action_parameters {
      cache = true
      browser_ttl {
        mode    = "override_origin"
        default = local.cache_static_files_for
      }
      cache_key {
        cache_deception_armor = true
        custom_key {
          query_string {
            exclude = ["*"]
          }
        }
      }
      edge_ttl {
        default = 86400
        mode    = "override_origin"
        status_code_ttl {
          status_code = 200
          value       = local.cache_static_files_for
        }
      }
      serve_stale {
        disable_stale_while_updating = true
      }
    }
    expression = local.known_static_files_expression
    enabled    = true
  }
}


