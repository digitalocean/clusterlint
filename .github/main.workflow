workflow "Pull Request" {
  # TODO(awg): This won't run checks on PRs from forks. For now that seems to be
  # a limitation of GH actions.
  on = "push"
  resolves = [
    "vet",
    "lint",
    "test",
  ]
}

action "vet" {
  uses = "docker://golang:1.12.6"
  env = {
    GOFLAGS = "-mod=vendor"
  }
  runs = ["go", "vet", "./..."]
}

action "lint" {
  uses = "./.github/golint"
  env = {
    GOFLAGS = "-mod=vendor"
  }
  runs = ["sh", "-c", "golint -set_exit_status $(go list ./...)"]
}

action "test" {
  uses = "docker://golang:1.12.6"
  env = {
    GOFLAGS = "-mod=vendor"
  }
  runs = ["go", "test", "-race", "-cover", "./..."]
}
