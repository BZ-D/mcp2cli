package schema

import (
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
)

func TestBuildArgumentsStrictValidation(t *testing.T) {
	tool := mcp.Tool{
		Name: "cityInfo",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "City name",
				},
				"days": map[string]any{
					"type":    "integer",
					"default": 1,
				},
				"units": map[string]any{
					"type": "string",
					"enum": []any{"c", "f"},
				},
			},
			Required: []string{"name"},
		},
	}

	spec, err := ParseToolInputSchema(tool)
	if err != nil {
		t.Fatalf("ParseToolInputSchema error: %v", err)
	}

	t.Run("missing required field", func(t *testing.T) {
		cmd := &cobra.Command{Use: "cityInfo"}
		if err := RegisterFlags(cmd, spec); err != nil {
			t.Fatalf("RegisterFlags error: %v", err)
		}
		if err := cmd.ParseFlags([]string{"--units", "c"}); err != nil {
			t.Fatalf("ParseFlags error: %v", err)
		}
		_, err := BuildArguments(cmd, spec)
		if err == nil {
			t.Fatalf("expected missing required error")
		}
		if !strings.Contains(err.Error(), "--name") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("enum validation", func(t *testing.T) {
		cmd := &cobra.Command{Use: "cityInfo"}
		if err := RegisterFlags(cmd, spec); err != nil {
			t.Fatalf("RegisterFlags error: %v", err)
		}
		if err := cmd.ParseFlags([]string{"--name", "hk", "--units", "k"}); err != nil {
			t.Fatalf("ParseFlags error: %v", err)
		}
		_, err := BuildArguments(cmd, spec)
		if err == nil {
			t.Fatalf("expected enum validation error")
		}
		if !strings.Contains(err.Error(), "enum") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		cmd := &cobra.Command{Use: "cityInfo"}
		if err := RegisterFlags(cmd, spec); err != nil {
			t.Fatalf("RegisterFlags error: %v", err)
		}
		if err := cmd.ParseFlags([]string{"--name", "hk", "--units", "c"}); err != nil {
			t.Fatalf("ParseFlags error: %v", err)
		}
		args, err := BuildArguments(cmd, spec)
		if err != nil {
			t.Fatalf("BuildArguments error: %v", err)
		}
		if got, want := args["name"], "hk"; got != want {
			t.Fatalf("name = %#v, want %#v", got, want)
		}
		// default should be included when not explicitly set.
		if got, want := args["days"], float64(1); got != want {
			// json defaults may preserve int or float depending source;
			// accept both numeric representations.
			if got != 1 {
				t.Fatalf("days = %#v, want %#v", got, want)
			}
		}
		if got, want := args["units"], "c"; got != want {
			t.Fatalf("units = %#v, want %#v", got, want)
		}
	})
}

func TestNestedArraySchemaUsesJSONFlag(t *testing.T) {
	tool := mcp.Tool{
		Name: "update_range",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"backgroundColors": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
			},
		},
	}

	spec, err := ParseToolInputSchema(tool)
	if err != nil {
		t.Fatalf("ParseToolInputSchema error: %v", err)
	}
	if got, want := len(spec.Fields), 1; got != want {
		t.Fatalf("len(spec.Fields) = %d, want %d", got, want)
	}
	field := spec.Fields[0]
	if got, want := field.Kind, KindArray; got != want {
		t.Fatalf("field.Kind = %s, want %s", got, want)
	}
	if got, want := field.FlagName, "backgroundColors-json"; got != want {
		t.Fatalf("field.FlagName = %q, want %q", got, want)
	}

	cmd := &cobra.Command{Use: "update_range"}
	if err := RegisterFlags(cmd, spec); err != nil {
		t.Fatalf("RegisterFlags error: %v", err)
	}
	if err := cmd.ParseFlags([]string{"--backgroundColors-json", `[["#fff","#000"],["#abc","#def"]]`}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}
	args, err := BuildArguments(cmd, spec)
	if err != nil {
		t.Fatalf("BuildArguments error: %v", err)
	}

	rows, ok := args["backgroundColors"].([]any)
	if !ok {
		t.Fatalf("backgroundColors = %#v, want []any", args["backgroundColors"])
	}
	if got, want := len(rows), 2; got != want {
		t.Fatalf("len(backgroundColors) = %d, want %d", got, want)
	}
	firstRow, ok := rows[0].([]any)
	if !ok {
		t.Fatalf("backgroundColors[0] = %#v, want []any", rows[0])
	}
	if got, want := firstRow[0], "#fff"; got != want {
		t.Fatalf("backgroundColors[0][0] = %#v, want %#v", got, want)
	}
}
