package button

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/dimiro1/faas-go/internal/ui/components/icons"
)

// Helper to render component and parse with goquery
func renderAndParse(t *testing.T, component interface {
	Render(context.Context, io.Writer) error
},
) *goquery.Document {
	t.Helper()

	pr, pw := io.Pipe()
	go func() {
		_ = component.Render(context.Background(), pw)
		pw.Close()
	}()

	doc, err := goquery.NewDocumentFromReader(pr)
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	return doc
}

func TestButton_RendersAsButton(t *testing.T) {
	doc := renderAndParse(t, Button(Props{
		Variant: Default,
	}))

	// Should render as button element
	buttons := doc.Find("button")
	if buttons.Length() != 1 {
		t.Errorf("expected 1 button, got %d", buttons.Length())
	}

	// Should have default type
	buttonType, _ := buttons.Attr("type")
	if buttonType != "button" {
		t.Errorf("expected type='button', got %q", buttonType)
	}
}

func TestButton_RendersAsAnchor(t *testing.T) {
	doc := renderAndParse(t, Button(Props{
		Variant: Default,
		Href:    "/test",
	}))

	// Should render as anchor element
	anchors := doc.Find("a")
	if anchors.Length() != 1 {
		t.Errorf("expected 1 anchor, got %d", anchors.Length())
	}

	// Should have correct href
	href, _ := anchors.Attr("href")
	if href != "/test" {
		t.Errorf("expected href='/test', got %q", href)
	}
}

func TestButton_ExternalLinkHasNoopener(t *testing.T) {
	doc := renderAndParse(t, Button(Props{
		Variant: Default,
		Href:    "https://example.com",
		Target:  "_blank",
	}))

	anchor := doc.Find("a")
	rel, exists := anchor.Attr("rel")
	if !exists || rel != "noopener noreferrer" {
		t.Errorf("expected rel='noopener noreferrer', got %q", rel)
	}
}

func TestButton_DisabledState(t *testing.T) {
	doc := renderAndParse(t, Button(Props{
		Variant:  Default,
		Disabled: true,
	}))

	button := doc.Find("button")
	_, disabled := button.Attr("disabled")
	if !disabled {
		t.Error("expected button to be disabled")
	}
}

func TestButton_LoadingState(t *testing.T) {
	doc := renderAndParse(t, Button(Props{
		Variant: Default,
		Loading: true,
	}))

	button := doc.Find("button")

	// Should be disabled when loading
	_, disabled := button.Attr("disabled")
	if !disabled {
		t.Error("expected button to be disabled when loading")
	}

	// Should have aria-busy
	ariaBusy, _ := button.Attr("aria-busy")
	if ariaBusy != "true" {
		t.Errorf("expected aria-busy='true', got %q", ariaBusy)
	}

	// Should contain spinner SVG
	svgs := button.Find("svg")
	if svgs.Length() != 1 {
		t.Errorf("expected 1 spinner svg, got %d", svgs.Length())
	}
}

func TestButton_WithIcon(t *testing.T) {
	doc := renderAndParse(t, Button(Props{
		Variant: Default,
		Icon:    icons.Plus(),
	}))

	button := doc.Find("button")
	svgs := button.Find("svg")
	if svgs.Length() != 1 {
		t.Errorf("expected 1 icon svg, got %d", svgs.Length())
	}
}

func TestButton_IconPositionRight(t *testing.T) {
	doc := renderAndParse(t, Button(Props{
		Variant:      Default,
		Icon:         icons.Plus(),
		IconPosition: IconRight,
	}))

	button := doc.Find("button")
	html, _ := button.Html()

	// SVG should come after text content (at the end)
	// This is a simple check - icon should be last child
	if !strings.HasSuffix(strings.TrimSpace(html), "</svg>") {
		t.Error("expected icon to be at the end for IconRight position")
	}
}

func TestButton_AriaLabel(t *testing.T) {
	doc := renderAndParse(t, Button(Props{
		Variant:   Default,
		Size:      SizeIcon,
		Icon:      icons.Plus(),
		AriaLabel: "Add new item",
	}))

	button := doc.Find("button")
	ariaLabel, _ := button.Attr("aria-label")
	if ariaLabel != "Add new item" {
		t.Errorf("expected aria-label='Add new item', got %q", ariaLabel)
	}
}

func TestButton_AriaExpanded(t *testing.T) {
	expanded := true
	doc := renderAndParse(t, Button(Props{
		Variant:      Default,
		AriaExpanded: &expanded,
		AriaControls: "menu",
	}))

	button := doc.Find("button")

	ariaExpanded, _ := button.Attr("aria-expanded")
	if ariaExpanded != "true" {
		t.Errorf("expected aria-expanded='true', got %q", ariaExpanded)
	}

	ariaControls, _ := button.Attr("aria-controls")
	if ariaControls != "menu" {
		t.Errorf("expected aria-controls='menu', got %q", ariaControls)
	}
}

func TestButton_AriaPressed(t *testing.T) {
	pressed := false
	doc := renderAndParse(t, Button(Props{
		Variant:     Default,
		AriaPressed: &pressed,
	}))

	button := doc.Find("button")
	ariaPressed, _ := button.Attr("aria-pressed")
	if ariaPressed != "false" {
		t.Errorf("expected aria-pressed='false', got %q", ariaPressed)
	}
}

func TestButton_SubmitType(t *testing.T) {
	doc := renderAndParse(t, Button(Props{
		Variant: Default,
		Type:    TypeSubmit,
	}))

	button := doc.Find("button")
	buttonType, _ := button.Attr("type")
	if buttonType != "submit" {
		t.Errorf("expected type='submit', got %q", buttonType)
	}
}

func TestButton_CustomID(t *testing.T) {
	doc := renderAndParse(t, Button(Props{
		Variant: Default,
		ID:      "my-button",
	}))

	button := doc.Find("button")
	id, _ := button.Attr("id")
	if id != "my-button" {
		t.Errorf("expected id='my-button', got %q", id)
	}
}

func TestButton_Variants(t *testing.T) {
	variants := []ButtonVariant{
		Default,
		Secondary,
		Destructive,
		Outline,
		Ghost,
		Link,
	}

	for _, variant := range variants {
		t.Run(string(variant), func(t *testing.T) {
			doc := renderAndParse(t, Button(Props{
				Variant: variant,
			}))

			button := doc.Find("button")
			if button.Length() != 1 {
				t.Errorf("expected 1 button for variant %s", variant)
			}

			// Check that class contains the variant-specific class
			class, _ := button.Attr("class")
			if class == "" {
				t.Errorf("expected button to have classes for variant %s", variant)
			}
		})
	}
}

func TestButton_Sizes(t *testing.T) {
	sizes := []ButtonSize{
		SizeDefault,
		SizeSm,
		SizeLg,
		SizeIcon,
	}

	for _, size := range sizes {
		t.Run(string(size), func(t *testing.T) {
			doc := renderAndParse(t, Button(Props{
				Variant: Default,
				Size:    size,
			}))

			button := doc.Find("button")
			if button.Length() != 1 {
				t.Errorf("expected 1 button for size %s", size)
			}

			class, _ := button.Attr("class")
			if class == "" {
				t.Errorf("expected button to have classes for size %s", size)
			}
		})
	}
}

func TestBackButton(t *testing.T) {
	doc := renderAndParse(t, BackButton("/home"))

	anchor := doc.Find("a")
	if anchor.Length() != 1 {
		t.Errorf("expected 1 anchor, got %d", anchor.Length())
	}

	href, _ := anchor.Attr("href")
	if href != "/home" {
		t.Errorf("expected href='/home', got %q", href)
	}

	// Should contain chevron icon
	svgs := anchor.Find("svg")
	if svgs.Length() != 1 {
		t.Errorf("expected 1 icon svg, got %d", svgs.Length())
	}

	// Should contain "Back" text
	text := anchor.Text()
	if !strings.Contains(text, "Back") {
		t.Errorf("expected text to contain 'Back', got %q", text)
	}
}
