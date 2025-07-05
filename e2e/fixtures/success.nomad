job "e2e-success-test" {
  type = "batch"
  
  group "app" {
    task "java-app" {
      driver = "nomad-driver-milo"
      
      artifact {
        source = "file://test-artifacts/hello-world.jar"
      }
      
      resources {
        cpu    = 100
        memory = 64
      }
    }
  }
}