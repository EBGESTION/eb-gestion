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
			Width:  1180,
			Height: 740,
			Center: true,
		},
	})
	if window == nil {
		return errors.New("WebView2 no está disponible; instala Microsoft Edge WebView2 Runtime")
	}
	defer window.Destroy()

	window.SetTitle("EB Gestión - Entre Bahías")
	// Una ventana inicial más compacta evita que Windows/WebView2 la escale
	// fuera del área útil en notebooks o pantallas con ampliación de DPI.
	window.SetSize(960, 620, webview2.HintMin)
	window.SetSize(1180, 740, webview2.HintNone)
	window.Navigate(url)
	window.Run()
	return nil
}
