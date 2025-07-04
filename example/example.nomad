# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

job "example" {
  datacenters = ["dc1"]
  type        = "batch"

  group "example" {
    task "milo-world" {
      driver = "milo-world-example"

      config {
        greeting = "hello"
      }
    }
  }
}
