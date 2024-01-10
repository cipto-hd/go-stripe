package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(SessionLoad)

	mux.Get("/", app.Home)
	mux.Get("/ws", app.WsEndPoint)

	/* routes for buy/subscribe and receipt */
	mux.Get("/widgets/{id}", app.ChargeOnce)             // show buy-once page
	mux.Post("/payment-succeeded", app.PaymentSucceeded) // store customer,transaction, order in the db
	mux.Get("/receipt", app.Receipt)                     // show receipt

	// all details about subscription process happen at backend API
	mux.Get("/plans/bronze", app.BronzePlan)          // show bronze-plan page
	mux.Get("/receipt/bronze", app.BronzePlanReceipt) // show bronze-plan receipt

	// mux.Post("/virtual-payment-succeeded", app.VirtualTerminalPaymentSucceeded)

	/* auth routes */
	mux.Get("/login", app.LoginPage)
	mux.Post("/login", app.PostLoginPage)
	mux.Get("/logout", app.Logout)
	mux.Get("/forgot-password", app.ForgotPassword)
	mux.Get("/reset-password", app.ShowResetPassword)

	/* admin routes */
	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.Auth)
		mux.Get("/virtual-terminal", app.VirtualTerminal)
		mux.Get("/sales", app.AllSales)
		mux.Get("/subscriptions", app.AllSubscriptions)
		mux.Get("/sales/{id}", app.ShowSale)
		mux.Get("/subscriptions/{id}", app.ShowSubscription)
		mux.Get("/users", app.AllUsers)
		mux.Get("/users/{id}", app.OneUser)
	})

	// the path is relative from working directory
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*",
		http.StripPrefix("/static/", fileServer),
	)

	return mux
}
