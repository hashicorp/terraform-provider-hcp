# Overrides the copywrite config schema version
# Default: 1
schema_version = 1

project {
  # SPDX-compatible license identifier
  # Leave blank if you don't wish to license the project
  # Default: "MPL-2.0"
  # license = ""

  # Represents the year that the project initially began
  # Default: <the year the repo was first created>
  # copyright_year = 0

  # A list of globs that should not have copyright or license headers
  # Supports doublestar glob patterns for more flexibility in defining which
  # files or folders should be ignored
  # Default: []
  header_ignore = [
    "examples/**"
  ]
}
