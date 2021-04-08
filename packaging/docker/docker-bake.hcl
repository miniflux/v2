group "default" {
  targets = ["build"]
}

target "ghaction-docker-meta" {}

target "build" {
  dockerfile = "./packaging/docker/Dockerfile"
  inherits = ["ghaction-docker-meta"]
  platforms = [
    "linux/amd64",
    "linux/arm/v6",
    "linux/arm/v7",
    "linux/arm64"
  ]
}
