job "milo-test" {
  datacenters = ["dc1"]
  type = "batch"

  group "greeting" {
    task "hello" {
      driver = "nomad-driver-milo"

      config {
        greeting = "Hello from Milo driver!"
      }

      resources {
        cpu    = 100
        memory = 64
      }
    }
  }
}