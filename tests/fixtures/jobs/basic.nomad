job "hello-world-test" {
  datacenters = ["dc1"]
  type = "batch"

  group "hello" {
    task "java-app" {
      driver = "milo"

      config {
        dummy = ""
      }

      artifact {
        source = "file:///home/sweeand/andrewesweet/nomad-driver-milo/tests/fixtures/hello-world.jar"
        destination = "local/"
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}