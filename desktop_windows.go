//go:build windows

package main

import (
	"errors"

	webview2 "github.com/jchv/go-webview2"
)

// runDesktop abre EB Gestión dentro de una ventana propia de Windows.
// El contenido continúa siendo servido por el servidor local de Go, por lo que
// los teléfonos y otros computadores pueden usar el mismo sistema en paralelo.
func runDesktop(url string) error {
	window := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     false,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  "EB Gestión - Entre Bahías",
			Width:  1366,
			Height: 820,
			Center: true,
		},
	})
	if window == nil {
		return errors.New("WebView2 no está disponible; instala Microsoft Edge WebView2 Runtime")
	}
	defer window.Destroy()

	window.SetTitle("EB Gestión - Entre Bahías")
	window.SetSize(1100, 700, webview2.HintMin)
	window.SetSize(1366, 820, webview2.HintNone)
	window.Navigate(url)
	window.Run()
	return nil
}
