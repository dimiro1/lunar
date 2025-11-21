package tabs

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func renderAndParse(t *testing.T, component interface {
	Render(context.Context, io.Writer) error
}) *goquery.Document {
	t.Helper()
	pr, pw := io.Pipe()
	go func() {
		_ = component.Render(context.Background(), pw)
		_ = pw.Close()
	}()
	doc, err := goquery.NewDocumentFromReader(pr)
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}
	return doc
}

func TestTabs_RendersAllTabs(t *testing.T) {
	doc := renderAndParse(t, Tabs(TabsProps{
		Tabs: []Tab{
			{ID: "metrics", Name: "Metrics", Href: "/function/123", Active: true},
			{ID: "code", Name: "Code", Href: "/function/123/code", Active: false},
			{ID: "settings", Name: "Settings", Href: "/function/123/settings", Active: false},
		},
	}))

	links := doc.Find("a")
	if links.Length() != 3 {
		t.Errorf("expected 3 tabs, got %d", links.Length())
	}
}

func TestTabs_ActiveTabHasCorrectClass(t *testing.T) {
	doc := renderAndParse(t, Tabs(TabsProps{
		Tabs: []Tab{
			{ID: "metrics", Name: "Metrics", Href: "/function/123", Active: true},
			{ID: "code", Name: "Code", Href: "/function/123/code", Active: false},
		},
	}))

	// First tab should be active
	firstTab := doc.Find("a").First()
	class, _ := firstTab.Attr("class")
	if !strings.Contains(class, "tabActive") {
		t.Error("expected active tab to have tabActive class")
	}

	// Second tab should not be active
	secondTab := doc.Find("a").Last()
	class, _ = secondTab.Attr("class")
	if strings.Contains(class, "tabActive") {
		t.Error("expected inactive tab to not have tabActive class")
	}
}

func TestTabs_UsesHref(t *testing.T) {
	doc := renderAndParse(t, Tabs(TabsProps{
		Tabs: []Tab{
			{ID: "metrics", Name: "Metrics", Href: "/function/123", Active: true},
		},
	}))

	link := doc.Find("a").First()
	href, _ := link.Attr("href")
	if href != "/function/123" {
		t.Errorf("expected tab href to be /function/123, got %s", href)
	}
}

func TestTabs_HasAriaAttributes(t *testing.T) {
	doc := renderAndParse(t, Tabs(TabsProps{
		Tabs: []Tab{
			{ID: "metrics", Name: "Metrics", Href: "#", Active: true},
			{ID: "code", Name: "Code", Href: "#", Active: false},
		},
	}))

	// Container should have role="tablist"
	container := doc.Find("[role='tablist']")
	if container.Length() != 1 {
		t.Error("expected container to have role=tablist")
	}

	// Tabs should have role="tab"
	tabs := doc.Find("[role='tab']")
	if tabs.Length() != 2 {
		t.Errorf("expected 2 tabs with role=tab, got %d", tabs.Length())
	}

	// Active tab should have aria-selected="true"
	firstTab := doc.Find("a").First()
	ariaSelected, _ := firstTab.Attr("aria-selected")
	if ariaSelected != "true" {
		t.Errorf("expected active tab to have aria-selected=true, got %s", ariaSelected)
	}

	// Inactive tab should have aria-selected="false"
	secondTab := doc.Find("a").Last()
	ariaSelected, _ = secondTab.Attr("aria-selected")
	if ariaSelected != "false" {
		t.Errorf("expected inactive tab to have aria-selected=false, got %s", ariaSelected)
	}
}

func TestTabs_DisabledTab(t *testing.T) {
	doc := renderAndParse(t, Tabs(TabsProps{
		Tabs: []Tab{
			{ID: "admin", Name: "Admin", Href: "#", Disabled: true},
		},
	}))

	tab := doc.Find("a").First()
	class, _ := tab.Attr("class")
	if !strings.Contains(class, "tabDisabled") {
		t.Error("expected disabled tab to have tabDisabled class")
	}

	ariaDisabled, _ := tab.Attr("aria-disabled")
	if ariaDisabled != "true" {
		t.Error("expected disabled tab to have aria-disabled=true")
	}

	tabindex, _ := tab.Attr("tabindex")
	if tabindex != "-1" {
		t.Error("expected disabled tab to have tabindex=-1")
	}
}

func TestTabs_WithBadge(t *testing.T) {
	doc := renderAndParse(t, Tabs(TabsProps{
		Tabs: []Tab{
			{ID: "inbox", Name: "Inbox", Href: "#", Badge: "12", Active: true},
		},
	}))

	badge := doc.Find("span").FilterFunction(func(i int, s *goquery.Selection) bool {
		return s.Text() == "12"
	})
	if badge.Length() != 1 {
		t.Error("expected badge with text '12'")
	}
}

func TestTabContent_HasAriaAttributes(t *testing.T) {
	doc := renderAndParse(t, TabContent(TabContentProps{
		ID:     "test",
		Active: true,
	}))

	div := doc.Find("div")

	role, _ := div.Attr("role")
	if role != "tabpanel" {
		t.Errorf("expected role=tabpanel, got %s", role)
	}

	ariaLabelledby, _ := div.Attr("aria-labelledby")
	if ariaLabelledby != "tab-test" {
		t.Errorf("expected aria-labelledby=tab-test, got %s", ariaLabelledby)
	}
}

func TestTabContent_ActiveDisplaysContent(t *testing.T) {
	doc := renderAndParse(t, TabContent(TabContentProps{
		ID:     "test",
		Active: true,
	}))

	div := doc.Find("div")
	class, _ := div.Attr("class")
	if !strings.Contains(class, "tabContentActive") {
		t.Error("expected active tab content to have tabContentActive class")
	}

	_, hasHidden := div.Attr("hidden")
	if hasHidden {
		t.Error("expected active tab content to not have hidden attribute")
	}
}

func TestTabContent_InactiveHidden(t *testing.T) {
	doc := renderAndParse(t, TabContent(TabContentProps{
		ID:     "test",
		Active: false,
	}))

	div := doc.Find("div")
	class, _ := div.Attr("class")
	if strings.Contains(class, "tabContentActive") {
		t.Error("expected inactive tab content to not have tabContentActive class")
	}

	_, hasHidden := div.Attr("hidden")
	if !hasHidden {
		t.Error("expected inactive tab content to have hidden attribute")
	}
}

func TestTabContainer_RendersChildren(t *testing.T) {
	doc := renderAndParse(t, TabContainer())

	div := doc.Find("div")
	if div.Length() != 1 {
		t.Errorf("expected 1 container div, got %d", div.Length())
	}
}
