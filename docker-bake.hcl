variable "IMAGE_NAME" {
  default = "ghcr.io/makepad-fr/alerteconso.com"
}

variable "REPOSITORY" {
  default = "Makepad-fr/alerteconso.com"
}

variable "GO_VERSION" {
  default = "1.24.3"
}

variable "BUILD_DATE" {
  default = "1970-01-01T00:00:00Z"
}

variable "VCS_REF" {
  default = "local"
}

variable "VERSION" {
  default = "local"
}

target "common" {
  context    = "."
  dockerfile = "Dockerfile"
  args = {
    BUILD_DATE = "${BUILD_DATE}"
    GO_VERSION = "${GO_VERSION}"
    VCS_REF    = "${VCS_REF}"
    VERSION    = "${VERSION}"
  }
  labels = {
    "org.opencontainers.image.created"  = "${BUILD_DATE}"
    "org.opencontainers.image.revision" = "${VCS_REF}"
    "org.opencontainers.image.source"   = "https://github.com/${REPOSITORY}"
    "org.opencontainers.image.version"  = "${VERSION}"
  }
}

target "runtime" {
  inherits = ["common"]
  target   = "runtime"
  tags     = ["${IMAGE_NAME}:local"]
}

target "publish" {
  inherits = ["runtime"]
  push     = true
  tags     = ["${IMAGE_NAME}:latest", "${IMAGE_NAME}:${VERSION}"]
}

target "local-image" {
  inherits = ["runtime"]
  load     = true
  tags     = ["${IMAGE_NAME}:local"]
}

group "local" {
  targets = ["local-image"]
}
