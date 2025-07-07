job "exit-code-test" {
  datacenters = ["dc1"]
  type = "batch"

  group "exit" {
    task "java-app" {
      driver = "milo"

      config {
        dummy = ""
      }

      artifact {
        source = "file:///home/sweeand/andrewesweet/nomad-driver-milo/tests/fixtures/exit-code-test.jar"
        destination = "local/"
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}