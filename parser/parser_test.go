package parser

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type testCase struct {
	name          string
	input         string
	expectedCmds  []Command
	expectedError error
}

func TestParse(t *testing.T) {
	testCases := []testCase{
		{
			name: "valid Matrfile",
			input: `//go:build matr

package main

// Summary for TestFunc1
func TestFunc1(ctc context.Context, args []string) {
}

// Summary for TestFunc2
func TestFunc2(ctx context.Context, args []string) {
}`,
			expectedCmds: []Command{
				{
					Name:       "TestFunc1",
					Summary:    "Summary for TestFunc1",
					Doc:        "Summary for TestFunc1",
					IsExported: true,
				},
				{
					Name:       "TestFunc2",
					Summary:    "Summary for TestFunc2",
					Doc:        "Summary for TestFunc2",
					IsExported: true,
				},
			},
			expectedError: nil,
		},
		{
			name:          "invalid Matrfile",
			input:         `package main`,
			expectedCmds:  []Command{},
			expectedError: errors.New("invalid Matrfile: matr build tag missing or incorrect"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary file for the test case
			tmpFile, err := os.CreateTemp("", "parser_test_*.go")
			if err != nil {
				t.Fatalf("Failed to create temporary file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			// Write the input content to the temporary file
			_, err = tmpFile.WriteString(tc.input)
			if err != nil {
				t.Fatalf("Failed to write to temporary file: %v", err)
			}
			tmpFile.Close()

			cmds, err := Parse(tmpFile.Name())

			if diff := cmp.Diff(tc.expectedCmds, cmds); diff != "" {
				t.Errorf("Commands mismatch (-want +got):\n%s", diff)
			}

			if tc.expectedError != nil {
				if err == nil || err.Error() != tc.expectedError.Error() {
					t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
