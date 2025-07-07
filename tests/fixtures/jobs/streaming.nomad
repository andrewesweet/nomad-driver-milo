job "streaming-test" {
  datacenters = ["dc1"]
  type = "service"

  group "streaming" {
    task "java-app" {
      driver = "milo"

      config {
        dummy = ""
      }

      artifact {
        source = "file:///home/sweeand/andrewesweet/nomad-driver-milo/tests/fixtures/long-running.jar"
        destination = "local/"
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}