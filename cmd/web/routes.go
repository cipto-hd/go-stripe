package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(SessionLoad)

	mux.Get("/", app.Home)
	mux.Get("/virtual-terminal", app.VirtualTerminal)
	mux.Post("/payment-succeeded", app.PaymentSucceeded)
	mux.Post("/virtual-payment-succeeded", app.VirtualTerminalPaymentSucceeded)
	mux.Get("/receipt", app.Receipt)
	mux.Get("/widgets/{id}", app.ChargeOnce)

	// the path is relative from working directory
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*",
		http.StripPrefix("/static/", fileServer),
	)

	return mux
}
