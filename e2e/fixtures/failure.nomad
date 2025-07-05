job "e2e-failure-test" {
  type = "batch"
  
  group "app" {
    task "java-app" {
      driver = "nomad-driver-milo"
      
      artifact {
        source = "file://non-existent-file.jar"
      }
      
      resources {
        cpu    = 100
        memory = 64
      }
    }
  }
}