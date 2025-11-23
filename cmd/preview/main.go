package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dimiro1/faas-go/internal/ui/components/badge"
	"github.com/dimiro1/faas-go/internal/ui/components/button"
	"github.com/dimiro1/faas-go/internal/ui/components/card"
	"github.com/dimiro1/faas-go/internal/ui/components/code"
	"github.com/dimiro1/faas-go/internal/ui/components/code_example"
	"github.com/dimiro1/faas-go/internal/ui/components/api_reference"
	"github.com/dimiro1/faas-go/internal/ui/components/diff"
	"github.com/dimiro1/faas-go/internal/ui/components/env_editor"
	"github.com/dimiro1/faas-go/internal/ui/components/form"
	"github.com/dimiro1/faas-go/internal/ui/components/icons"
	"github.com/dimiro1/faas-go/internal/ui/components/kbd"
	logcomponent "github.com/dimiro1/faas-go/internal/ui/components/log"
	"github.com/dimiro1/faas-go/internal/ui/components/navbar"
	"github.com/dimiro1/faas-go/internal/ui/components/pagination"
	"github.com/dimiro1/faas-go/internal/ui/components/preview"
	"github.com/dimiro1/faas-go/internal/ui/components/request_builder"
	"github.com/dimiro1/faas-go/internal/ui/components/table"
	"github.com/dimiro1/faas-go/internal/ui/components/tabs"
	"github.com/dimiro1/faas-go/internal/ui/pages"
	"github.com/dimiro1/faas-go/internal/ui/pages/login"
)

func main() {
	mux := http.NewServeMux()

	// Serve static CSS files
	fs := http.FileServer(http.Dir("internal/ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Preview routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		_ = pages.PreviewDashboard().Render(r.Context(), w)
	})

	mux.HandleFunc("/functions/new", func(w http.ResponseWriter, r *http.Request) {
		_ = pages.PreviewCreateFunction().Render(r.Context(), w)
	})

	mux.HandleFunc("/functions/hello", func(w http.ResponseWriter, r *http.Request) {
		_ = pages.PreviewFunctionDetails().Render(r.Context(), w)
	})

	mux.HandleFunc("/functions/hello/code", func(w http.ResponseWriter, r *http.Request) {
		_ = pages.PreviewCodeTab().Render(r.Context(), w)
	})

	mux.HandleFunc("/functions/hello/settings", func(w http.ResponseWriter, r *http.Request) {
		_ = pages.PreviewSettingsTab().Render(r.Context(), w)
	})

	mux.HandleFunc("/functions/hello/executions", func(w http.ResponseWriter, r *http.Request) {
		_ = pages.PreviewExecutionsTab().Render(r.Context(), w)
	})

	mux.HandleFunc("/functions/hello/test", func(w http.ResponseWriter, r *http.Request) {
		_ = pages.PreviewTestTab().Render(r.Context(), w)
	})

	mux.HandleFunc("/functions/hello/executions/exec_12345abcde", func(w http.ResponseWriter, r *http.Request) {
		_ = pages.PreviewExecutionDetails().Render(r.Context(), w)
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		_ = login.PreviewLogin().Render(r.Context(), w)
	})

	mux.HandleFunc("/login-error", func(w http.ResponseWriter, r *http.Request) {
		_ = login.PreviewLoginWithError().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/badge", func(w http.ResponseWriter, r *http.Request) {
		_ = badge.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/button", func(w http.ResponseWriter, r *http.Request) {
		_ = button.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/card", func(w http.ResponseWriter, r *http.Request) {
		_ = card.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/code", func(w http.ResponseWriter, r *http.Request) {
		_ = code.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/form", func(w http.ResponseWriter, r *http.Request) {
		_ = form.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/pagination", func(w http.ResponseWriter, r *http.Request) {
		_ = pagination.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/table", func(w http.ResponseWriter, r *http.Request) {
		_ = table.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/tabs", func(w http.ResponseWriter, r *http.Request) {
		_ = tabs.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/icons", func(w http.ResponseWriter, r *http.Request) {
		_ = icons.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/kbd", func(w http.ResponseWriter, r *http.Request) {
		_ = kbd.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/log", func(w http.ResponseWriter, r *http.Request) {
		_ = logcomponent.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/navbar", func(w http.ResponseWriter, r *http.Request) {
		_ = navbar.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/code-example", func(w http.ResponseWriter, r *http.Request) {
		_ = code_example.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/diff", func(w http.ResponseWriter, r *http.Request) {
		_ = diff.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/api-reference", func(w http.ResponseWriter, r *http.Request) {
		_ = api_reference.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/env-editor", func(w http.ResponseWriter, r *http.Request) {
		_ = env_editor.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/request-builder", func(w http.ResponseWriter, r *http.Request) {
		_ = request_builder.Preview().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview", func(w http.ResponseWriter, r *http.Request) {
		_ = preview.Index().Render(r.Context(), w)
	})

	// Wrap with CSS middleware - this serves /styles/templ.css automatically
	handler := templ.NewCSSMiddleware(mux)

	fmt.Println("Preview server running at http://localhost:8080")
	fmt.Println("Routes:")
	fmt.Println("  - http://localhost:8080/                        (Dashboard)")
	fmt.Println("  - http://localhost:8080/functions/new           (Create Function)")
	fmt.Println("  - http://localhost:8080/functions/hello         (Metrics Tab)")
	fmt.Println("  - http://localhost:8080/functions/hello/code    (Code Tab)")
	fmt.Println("  - http://localhost:8080/functions/hello/settings (Settings Tab)")
	fmt.Println("  - http://localhost:8080/functions/hello/executions (Executions Tab)")
	fmt.Println("  - http://localhost:8080/functions/hello/test    (Test Tab)")
	fmt.Println("  - http://localhost:8080/functions/hello/executions/exec_12345abcde (Execution Details)")
	fmt.Println("  - http://localhost:8080/login                    (Login)")
	fmt.Println("  - http://localhost:8080/login-error              (Login with Error)")
	fmt.Println("  - http://localhost:8080/preview                   (Component Index)")
	fmt.Println("  - http://localhost:8080/preview/component/badge  (Badge Component)")
	fmt.Println("  - http://localhost:8080/preview/component/button (Button Component)")
	fmt.Println("  - http://localhost:8080/preview/component/card   (Card Component)")
	fmt.Println("  - http://localhost:8080/preview/component/code   (Code Component)")
	fmt.Println("  - http://localhost:8080/preview/component/form   (Form Component)")
	fmt.Println("  - http://localhost:8080/preview/component/pagination (Pagination Component)")
	fmt.Println("  - http://localhost:8080/preview/component/table  (Table Component)")
	fmt.Println("  - http://localhost:8080/preview/component/tabs   (Tabs Component)")
	fmt.Println("  - http://localhost:8080/preview/component/icons  (Icons Component)")
	fmt.Println("  - http://localhost:8080/preview/component/kbd    (Kbd Component)")
	fmt.Println("  - http://localhost:8080/preview/component/log    (Log Component)")
	fmt.Println("  - http://localhost:8080/preview/component/navbar (Navbar Component)")
	fmt.Println("  - http://localhost:8080/preview/component/code-example (Code Example Component)")
	fmt.Println("  - http://localhost:8080/preview/component/diff (Diff Component)")
	fmt.Println("  - http://localhost:8080/preview/component/api-reference (API Reference Component)")
	fmt.Println("  - http://localhost:8080/preview/component/request-builder (Request Builder Component)")

	log.Fatal(http.ListenAndServe(":8080", handler))
}
