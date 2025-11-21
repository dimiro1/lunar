package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/dimiro1/faas-go/internal/ui/components/button"
	"github.com/dimiro1/faas-go/internal/ui/pages"
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
		pages.PreviewDashboard().Render(r.Context(), w)
	})

	mux.HandleFunc("/functions/new", func(w http.ResponseWriter, r *http.Request) {
		pages.PreviewCreateFunction().Render(r.Context(), w)
	})

	mux.HandleFunc("/functions/hello", func(w http.ResponseWriter, r *http.Request) {
		pages.PreviewFunctionDetails().Render(r.Context(), w)
	})

	mux.HandleFunc("/functions/hello/code", func(w http.ResponseWriter, r *http.Request) {
		pages.PreviewCodeTab().Render(r.Context(), w)
	})

	mux.HandleFunc("/functions/hello/settings", func(w http.ResponseWriter, r *http.Request) {
		pages.PreviewSettingsTab().Render(r.Context(), w)
	})

	mux.HandleFunc("/functions/hello/executions", func(w http.ResponseWriter, r *http.Request) {
		pages.PreviewExecutionsTab().Render(r.Context(), w)
	})

	mux.HandleFunc("/functions/hello/test", func(w http.ResponseWriter, r *http.Request) {
		pages.PreviewTestTab().Render(r.Context(), w)
	})

	mux.HandleFunc("/functions/hello/executions/exec_12345abcde", func(w http.ResponseWriter, r *http.Request) {
		pages.PreviewExecutionDetails().Render(r.Context(), w)
	})

	mux.HandleFunc("/preview/component/button", func(w http.ResponseWriter, r *http.Request) {
		button.Preview().Render(r.Context(), w)
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
	fmt.Println("  - http://localhost:8080/preview/component/button (Button Component)")

	log.Fatal(http.ListenAndServe(":8080", handler))
}
