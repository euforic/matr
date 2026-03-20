package harness

import "testing"

func TestValidateConventionalCommitSubject(t *testing.T) {
	testCases := []struct {
		name    string
		subject string
		wantErr bool
	}{
		{
			name:    "valid conventional commit",
			subject: "docs: add harness foundations",
			wantErr: false,
		},
		{
			name:    "valid longer conventional commit",
			subject: "docs: scaffold harness foundations with matr workflows",
			wantErr: false,
		},
		{
			name:    "valid breaking change marker",
			subject: "feat!: change target naming",
			wantErr: false,
		},
		{
			name:    "valid scoped commit",
			subject: "test(parser): cover build tag parsing",
			wantErr: false,
		},
		{
			name:    "missing type separator",
			subject: "add harness foundations",
			wantErr: true,
		},
		{
			name:    "capitalized type",
			subject: "Docs: add harness foundations",
			wantErr: true,
		},
		{
			name:    "missing description",
			subject: "docs:",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateConventionalCommitSubject(tc.subject)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for %q", tc.subject)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.subject, err)
			}
		})
	}
}
