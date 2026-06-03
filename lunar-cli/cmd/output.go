package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"
)

// outputWriter is where all rendered output is written.
// Defaults to os.Stdout; tests swap this to capture output.
var outputWriter io.Writer = os.Stdout

// outputFormat is set by the --output / -o flag on rootCmd.
var outputFormat string

// showCode controls whether code fields are printed when rendering objects.
var showCode bool

// preferredColumns are shown first (in order) when rendering tables.
var preferredColumns = []string{
	"id", "name", "version", "number", "status", "is_active",
	"disabled", "cron_status", "save_response",
	"created_at", "updated_at",
}

// skipInListView skips these fields when rendering a list table.
var skipInListView = map[string]bool{
	"code":           true,
	"active_version": true,
	"env_vars":       true,
	"kv":             true,
	"scoped_data":    true,
	"global_data":    true,
	"function_id":    true,
}

// skipInObjectView skips fields from the main scalar table; they are rendered
// as labelled sections below.
var skipInObjectView = map[string]bool{
	"code":        true,
	"env_vars":    true,
	"scoped_data": true,
	"global_data": true,
}

// renderAsKVTable renders these map fields as key/value tables instead of raw JSON.
var renderAsKVTable = map[string]bool{
	"env_vars":    true,
	"scoped_data": true,
	"global_data": true,
}

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			Padding(0, 1)

	valStyle = lipgloss.NewStyle().Padding(0, 1)

	dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99"))

	verticalKeyStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99")).
				Width(18)

	rowSepStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("237"))
)

// printRawJSON pretty-prints JSON to outputWriter.
func printRawJSON(data []byte) error {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		fmt.Fprintln(outputWriter, string(data))
		return nil
	}
	fmt.Fprintln(outputWriter, buf.String())
	return nil
}

// printPretty detects the shape of the JSON and renders styled output.
func printPretty(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Fprintln(outputWriter, string(data))
		return nil
	}

	switch t := v.(type) {
	case []any:
		if len(t) == 0 {
			fmt.Fprintln(outputWriter, dimStyle.Render("(no results)"))
			return nil
		}
		return renderList(t)
	case map[string]any:
		if items, pagination, ok := detectEnvelope(t); ok {
			return renderEnvelope(items, pagination)
		}
		return renderObject(t)
	default:
		enc, _ := json.Marshal(v)
		fmt.Fprintln(outputWriter, string(enc))
	}
	return nil
}

// detectEnvelope checks if obj is {"<key>": [...], "pagination": {...}}.
func detectEnvelope(obj map[string]any) ([]any, map[string]any, bool) {
	var arrayKey string
	var items []any
	var pagination map[string]any
	found := false

	for k, v := range obj {
		if k == "pagination" {
			if p, ok := v.(map[string]any); ok {
				pagination = p
			}
			continue
		}
		switch val := v.(type) {
		case []any:
			if found {
				return nil, nil, false
			}
			arrayKey = k
			items = val
			found = true
		case nil:
			if found {
				return nil, nil, false
			}
			arrayKey = k
			items = nil
			found = true
		default:
			return nil, nil, false
		}
	}
	_ = arrayKey
	return items, pagination, found
}

// renderEnvelope renders a paginated list response: table/vertical + footer.
func renderEnvelope(items []any, pagination map[string]any) error {
	if len(items) == 0 {
		fmt.Fprintln(outputWriter, dimStyle.Render("(no results)"))
	} else {
		if err := renderList(items); err != nil {
			return err
		}
	}
	if pagination != nil {
		total := int64(pagination["total"].(float64))
		limit := int64(pagination["limit"].(float64))
		offset := int64(pagination["offset"].(float64))
		if len(items) == 0 {
			fmt.Fprintln(outputWriter, dimStyle.Render(fmt.Sprintf("showing 0 of %d  (limit %d, offset %d)", total, limit, offset)))
			return nil
		}
		end := offset + int64(len(items))
		fmt.Fprintln(outputWriter, dimStyle.Render(fmt.Sprintf("showing %d–%d of %d  (limit %d)", offset+1, end, total, limit)))
	}
	return nil
}

// renderList renders a JSON array. Switches to vertical layout if the table
// would exceed the terminal width.
func renderList(items []any) error {
	keySet := make(map[string]struct{})
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for k, v := range obj {
			if skipInListView[k] {
				continue
			}
			if isSimpleValue(v) {
				keySet[k] = struct{}{}
			}
		}
	}
	headers := orderedKeys(keySet)
	if len(headers) == 0 {
		b, _ := json.MarshalIndent(items, "", "  ")
		fmt.Fprintln(outputWriter, string(b))
		return nil
	}

	rows := make([][]string, 0, len(items))
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		row := make([]string, len(headers))
		for i, h := range headers {
			row[i] = formatValue(h, obj[h])
		}
		rows = append(rows, row)
	}

	return renderVerticalList(headers, rows)
}

// renderVerticalList renders each item as a compact key: value block.
func renderVerticalList(headers []string, rows [][]string) error {
	sep := rowSepStyle.Render(strings.Repeat("─", 40))
	for i, row := range rows {
		if i > 0 {
			fmt.Fprintln(outputWriter, sep)
		}
		for j, h := range headers {
			fmt.Fprintf(outputWriter, "%s %s\n", verticalKeyStyle.Render(h), row[j])
		}
	}
	return nil
}

// renderObject renders a JSON object as a two-column key/value table,
// with complex nested fields printed as sections below.
func renderObject(obj map[string]any) error {
	keySet := make(map[string]struct{})
	for k, v := range obj {
		if skipInObjectView[k] {
			continue
		}
		if isSimpleValue(v) {
			keySet[k] = struct{}{}
		}
	}
	// Also include small nested objects that fit in one line.
	for k, v := range obj {
		if skipInObjectView[k] {
			continue
		}
		if m, ok := v.(map[string]any); ok {
			b, _ := json.Marshal(m)
			if len(b) <= 80 {
				keySet[k] = struct{}{}
			}
		}
	}

	keys := orderedKeys(keySet)
	if len(keys) > 0 {
		rows := make([][]string, 0, len(keys))
		for _, k := range keys {
			rows = append(rows, []string{k, formatValue(k, obj[k])})
		}
		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(borderStyle).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return headerStyle
				}
				if col == 0 {
					return keyStyle
				}
				return valStyle
			}).
			Headers("field", "value").
			Rows(rows...)
		fmt.Fprintln(outputWriter, t)
	}

	// Print complex nested fields as sections.
	complexKeys := make([]string, 0)
	for k := range obj {
		if _, inTable := keySet[k]; inTable {
			continue
		}
		complexKeys = append(complexKeys, k)
	}
	sort.Strings(complexKeys)

	for _, k := range complexKeys {
		v := obj[k]
		fmt.Fprintln(outputWriter, sectionStyle.Render("── "+k+" ──"))
		switch val := v.(type) {
		case string:
			if k == "code" {
				if !showCode {
					fmt.Fprintln(outputWriter, dimStyle.Render("  (use --show-code to display)"))
				} else {
					printCode(val)
				}
			} else {
				fmt.Fprintln(outputWriter, val)
			}
		case map[string]any:
			if renderAsKVTable[k] {
				renderKVMap(val)
			} else {
				if err := renderObject(val); err != nil {
					return err
				}
			}
		default:
			b, _ := json.MarshalIndent(v, "", "  ")
			fmt.Fprintln(outputWriter, string(b))
		}
	}

	return nil
}

// renderKVMap renders a flat map[string]any as a two-column key/value table.
// Empty maps show a dim "(empty)" line instead of a blank table.
func renderKVMap(m map[string]any) {
	if len(m) == 0 {
		fmt.Fprintln(outputWriter, dimStyle.Render("  (empty)"))
		return
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	rows := make([][]string, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, []string{k, formatValue(k, m[k])})
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(borderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			if col == 0 {
				return keyStyle
			}
			return valStyle
		}).
		Headers("key", "value").
		Rows(rows...)
	fmt.Fprintln(outputWriter, t)
}

// printCode renders Lua source with syntax highlighting (falls back to plain).
func printCode(code string) {
	lexer := lexers.Get("lua")
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Use terminal256 formatter; fall back to plain if stdout is not a tty.
	var formatterName string
	if isTerminal() {
		formatterName = "terminal256"
	} else {
		formatterName = "noop"
	}

	formatter := formatters.Get(formatterName)
	if formatter == nil {
		fmt.Fprintln(outputWriter, code)
		return
	}

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		fmt.Fprintln(outputWriter, code)
		return
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		fmt.Fprintln(outputWriter, code)
		return
	}
	fmt.Fprintln(outputWriter, buf.String())
}

// isSimpleValue returns true for scalars that fit nicely in a table cell.
func isSimpleValue(v any) bool {
	switch v.(type) {
	case string, bool, float64, nil:
		return true
	}
	return false
}

// orderedKeys returns keys sorted: preferredColumns first, then alphabetically.
func orderedKeys(keySet map[string]struct{}) []string {
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		pi := slices.Index(preferredColumns, keys[i])
		pj := slices.Index(preferredColumns, keys[j])
		switch {
		case pi >= 0 && pj >= 0:
			return pi < pj
		case pi >= 0:
			return true
		case pj >= 0:
			return false
		default:
			return keys[i] < keys[j]
		}
	})
	return keys
}

// formatValue converts a value to a display string.
func formatValue(key string, v any) string {
	if v == nil {
		return ""
	}
	if isTimestampKey(key) {
		if ts, ok := v.(float64); ok && ts > 0 {
			return time.Unix(int64(ts), 0).Local().Format("2006-01-02 15:04")
		}
	}
	switch t := v.(type) {
	case string:
		if len(t) > 72 {
			return t[:69] + "..."
		}
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	case []any, map[string]any:
		b, _ := json.Marshal(v)
		s := string(b)
		if len(s) > 72 {
			s = s[:69] + "..."
		}
		return s
	default:
		b, _ := json.Marshal(v)
		return strings.Trim(string(b), `"`)
	}
}

// isTimestampKey returns true for field names that hold Unix timestamps.
func isTimestampKey(key string) bool {
	return strings.HasSuffix(key, "_at") || key == "last_used" || key == "timestamp"
}

// isTerminal returns true when stdout is an interactive terminal.
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
