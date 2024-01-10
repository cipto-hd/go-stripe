package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	mux.Get("/api/check", app.Check)
	mux.Post("/api/payment-intent", app.GetPaymentIntent)

	mux.Get("/api/widgets/{id}", app.GetWidgetByID)

	mux.Post("/api/create-customer-and-subscribe-to-plan", app.CreateCustomerAndSubscribeToPlan)

	mux.Get("/api/test", func(w http.ResponseWriter, r *http.Request) {
		user, err := app.DB.GetUserByEmail("fool@example.com")
		app.infoLog.Printf("%+v, %+v", user, err)
		w.Write([]byte("Hit api admin"))
	})

	/* auth routes */
	mux.Post("/api/authenticate", app.CreateAuthToken)
	mux.Post("/api/is-authenticated", app.CheckAuthentication)
	mux.Post("/api/forgot-password", app.SendPasswordResetEmail)
	mux.Post("/api/reset-password", app.ResetPassword)

	/* admin routes */
	mux.Route("/api/admin", func(mux chi.Router) {
		mux.Use(app.Auth)

		mux.Post("/virtual-terminal-succeeded", app.VirtualTerminalPaymentSucceeded)
		mux.Post("/sales", app.AllSales)
		mux.Post("/subscriptions", app.AllSubscriptions)

		mux.Post("/sales/{id}", app.GetSale)

		mux.Post("/refund", app.RefundCharge)
		mux.Post("/cancel-subscription", app.CancelSubscription)

		mux.Post("/users", app.AllUsers)
		mux.Post("/users/{id}", app.OneUser)
		mux.Post("/users/edit/{id}", app.EditUser)
		mux.Post("/users/delete/{id}", app.DeleteUser)

	})

	return mux
}
