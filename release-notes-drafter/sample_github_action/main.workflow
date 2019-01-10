workflow "Release Notes Drafter" {
  on = "pull_request"
  resolves = ["Draft Release Notes"]
}

action "Draft Release Notes" {
  uses = "./release_notes_drafter"
  args = "action"
  secrets = ["GITHUB_TOKEN", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"]
}
