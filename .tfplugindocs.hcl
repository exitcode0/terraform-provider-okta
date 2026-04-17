# Ensures terraform-plugin-docs uses the correct provider name even when this repo
# is checked out into a non-standard directory name (e.g. a worktree).
#
# Without this, tfplugindocs derives the provider name from the directory name,
# which can cause generation failures like:
#   data source entitled "<dir>", or "<dir>_<name>" does not exist
provider {
  name          = "okta"
  rendered_name = "Okta"
}
