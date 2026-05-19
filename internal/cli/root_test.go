package cli

import (
	"strings"
	"testing"

	"github.com/hengyunabc/mcp2cli/internal/schema"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestToolExampleUsesValidJSONForGenericArray(t *testing.T) {
	example := toolExample("sheet", mcp.Tool{Name: "update_range"}, schema.Spec{
		Fields: []schema.Field{
			{
				Name:     "values",
				FlagName: "values-json",
				Required: true,
				Kind:     schema.KindArray,
			},
		},
	})

	if !strings.Contains(example, `sheet update_range --values-json '["value"]'`) {
		t.Fatalf("example does not contain JSON array sample:\n%s", example)
	}
	if strings.Contains(example, "--values-json value") {
		t.Fatalf("example contains invalid generic value sample:\n%s", example)
	}
}
