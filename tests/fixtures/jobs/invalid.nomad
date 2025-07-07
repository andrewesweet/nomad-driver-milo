job "invalid-test" {
  datacenters = ["dc1"]
  type = "batch"

  group "invalid" {
    task "java-app" {
      driver = "milo"

      config {
        dummy = ""
      }

      artifact {
        source = "file:///home/sweeand/andrewesweet/nomad-driver-milo/tests/fixtures/src/HelloWorld.java"
        destination = "local/"
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}