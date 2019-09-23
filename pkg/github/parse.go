package github

import (
	"strings"
)

// Owner extracts the owner from the provided repository HTTPS or SSH URL.
func Owner(repo string) string {
	// e.g. github.com/marccarre/my-gitops-repo -> marccarre
	return strings.Split(repo, "/")[1]
}
