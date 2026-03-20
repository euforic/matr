package harness

import (
	"fmt"
	"regexp"
	"strings"
)

var conventionalCommitSubjectPattern = regexp.MustCompile(`^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\([a-z0-9._/-]+\))?(!)?: .+`)

// ValidateConventionalCommitSubject checks whether a commit subject follows the
// conventional commits format required by this repository.
func ValidateConventionalCommitSubject(subject string) error {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return fmt.Errorf("commit subject is empty")
	}

	if !conventionalCommitSubjectPattern.MatchString(subject) {
		return fmt.Errorf("commit subject %q must match conventional commits", subject)
	}

	return nil
}
