package log

import (
	"context"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestViewer_RendersEntries(t *testing.T) {
	entries := []Entry{
		{Timestamp: "10:42:05.100", Level: LevelInfo, Message: "Test message"},
	}

	var buf strings.Builder
	err := Viewer(Props{}, entries).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "10:42:05.100") {
		t.Error("expected timestamp not found")
	}
	if !strings.Contains(html, "INFO") {
		t.Error("expected level not found")
	}
	if !strings.Contains(html, "Test message") {
		t.Error("expected message not found")
	}
}

func TestViewer_EmptyState(t *testing.T) {
	var buf strings.Builder
	err := Viewer(Props{}, []Entry{}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "No logs available") {
		t.Error("expected empty state message not found")
	}
}

func TestViewer_MultipleEntries(t *testing.T) {
	entries := []Entry{
		{Timestamp: "10:42:05.100", Level: LevelInfo, Message: "First"},
		{Timestamp: "10:42:05.101", Level: LevelError, Message: "Second"},
		{Timestamp: "10:42:05.102", Level: LevelWarn, Message: "Third"},
	}

	var buf strings.Builder
	err := Viewer(Props{}, entries).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "First") {
		t.Error("expected first message not found")
	}
	if !strings.Contains(html, "Second") {
		t.Error("expected second message not found")
	}
	if !strings.Contains(html, "Third") {
		t.Error("expected third message not found")
	}
}

func TestViewer_WithID(t *testing.T) {
	var buf strings.Builder
	err := Viewer(Props{ID: "my-logs"}, []Entry{}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	container := doc.Find("#my-logs")
	if container.Length() == 0 {
		t.Error("expected container with ID 'my-logs' not found")
	}
}

func TestViewer_WithMaxHeight(t *testing.T) {
	var buf strings.Builder
	err := Viewer(Props{MaxHeight: "200px"}, []Entry{}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "max-height: 200px") {
		t.Error("expected max-height style not found")
	}
}

func TestViewer_LevelInfo(t *testing.T) {
	entries := []Entry{
		{Timestamp: "10:42:05.100", Level: LevelInfo, Message: "Info message"},
	}

	var buf strings.Builder
	err := Viewer(Props{}, entries).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "INFO") {
		t.Error("expected INFO level not found")
	}
}

func TestViewer_LevelError(t *testing.T) {
	entries := []Entry{
		{Timestamp: "10:42:05.100", Level: LevelError, Message: "Error message"},
	}

	var buf strings.Builder
	err := Viewer(Props{}, entries).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "ERROR") {
		t.Error("expected ERROR level not found")
	}
}

func TestViewer_LevelWarn(t *testing.T) {
	entries := []Entry{
		{Timestamp: "10:42:05.100", Level: LevelWarn, Message: "Warn message"},
	}

	var buf strings.Builder
	err := Viewer(Props{}, entries).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "WARN") {
		t.Error("expected WARN level not found")
	}
}

func TestViewer_LevelDebug(t *testing.T) {
	entries := []Entry{
		{Timestamp: "10:42:05.100", Level: LevelDebug, Message: "Debug message"},
	}

	var buf strings.Builder
	err := Viewer(Props{}, entries).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "DEBUG") {
		t.Error("expected DEBUG level not found")
	}
}

func TestEntryRow_RendersAllParts(t *testing.T) {
	entry := Entry{
		Timestamp: "10:42:05.100",
		Level:     LevelInfo,
		Message:   "Test message",
	}

	var buf strings.Builder
	err := EntryRow(entry, false).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "10:42:05.100") {
		t.Error("expected timestamp not found")
	}
	if !strings.Contains(html, "INFO") {
		t.Error("expected level not found")
	}
	if !strings.Contains(html, "Test message") {
		t.Error("expected message not found")
	}
}

func TestGetLevelClass_ReturnsCorrectClass(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelInfo, "logLevelInfo"},
		{LevelError, "logLevelError"},
		{LevelWarn, "logLevelWarn"},
		{LevelDebug, "logLevelDebug"},
		{"UNKNOWN", "logLevelInfo"}, // defaults to info
	}

	for _, tt := range tests {
		class := getLevelClass(tt.level)
		className := class.ClassName()
		if !strings.Contains(className, tt.expected) {
			t.Errorf("getLevelClass(%s) = %s, expected to contain %s", tt.level, className, tt.expected)
		}
	}
}
