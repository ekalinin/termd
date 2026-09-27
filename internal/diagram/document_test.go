package diagram_test

import (
	"strings"
	"testing"

	mdiagram "github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	mrender "github.com/AlexanderGrooff/mermaid-ascii/pkg/render"

	"github.com/ekalinin/termd/internal/diagram"
	"github.com/ekalinin/termd/internal/render"
)

// TestPanicDoesNotStopDocument injects a panicking renderer and checks that
// the diagram becomes framed source and the next block is still rendered.
func TestPanicDoesNotStopDocument(t *testing.T) {
	orig := *diagram.RenderDiagramHook
	defer func() { *diagram.RenderDiagramHook = orig }()
	*diagram.RenderDiagramHook = func(string, *mdiagram.Config) (string, mrender.WidthStatus, error) {
		panic("boom")
	}
	src := "```mermaid\nsequenceDiagram\n  A->>B: hi\n```\n\nThe paragraph after.\n"
	out := render.Render([]byte(src), render.Options{Width: 80})
	if !strings.Contains(out, "renderer failed: boom") || !strings.Contains(out, "A->>B: hi") {
		t.Errorf("diagram is not shown as framed source:\n%s", out)
	}
	if !strings.Contains(out, "The paragraph after.") {
		t.Errorf("the next block was not rendered:\n%s", out)
	}
}
