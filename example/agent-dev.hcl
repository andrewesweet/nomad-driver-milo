# Development configuration for testing the milo driver
# Usage: nomad agent -dev -config=example/agent-dev.hcl -plugin-dir=/tmp/nomad-plugins

# Enable both server and client for dev mode
server {
  enabled          = true
  bootstrap_expect = 1
}

client {
  enabled = true
  
  # Disable unnecessary fingerprinters for faster startup
  options = {
    "fingerprint.denylist" = "env_aws,env_gce,env_azure,env_digitalocean"
  }
}

# UI configuration
ui {
  enabled = true
  show_cli_hints = false
}

# Logging
log_level = "DEBUG"

# Plugin configuration
plugin "nomad-driver-milo" {
  config {
    shell = "bash"
  }
}