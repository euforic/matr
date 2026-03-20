package matr

import (
	"bytes"
	"strings"
	"testing"

	"github.com/euforic/matr/parser"
)

func TestGenerateDoesNotRequireExecutionTimeoutHelper(t *testing.T) {
	var out bytes.Buffer

	err := generate([]parser.Command{
		{Name: "Sleepy", Summary: "Sleepy", Doc: "Sleepy", IsExported: true},
	}, &out)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	rendered := out.String()
	if strings.Contains(rendered, "matr.ExecutionTimeout()") {
		t.Fatalf("generated code should not depend on matr.ExecutionTimeout for compatibility")
	}
	if !strings.Contains(rendered, "os.Getenv(\"MATR_TIMEOUT\")") {
		t.Fatalf("generated code should read MATR_TIMEOUT directly")
	}
}
