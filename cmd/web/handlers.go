package main

import (
	"myapp/internal/cards"
	"myapp/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

/***************************************
ROUTES BLOCK - START
****************************************/
/********************
Handler for "/" route
*********************/
func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "home", nil); err != nil {
		app.errorLog.Println(err)
	}
}

/*
*******************
Handler for "/virtual-terminal" route
********************
*/
func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "terminal", nil); err != nil {
		app.errorLog.Println(err)
	}
}

/*
*******************
Handler for "/payment-succeeded" route
********************
*/
func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	var txnData TransactionData

	err := r.ParseForm()

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// read posted data
	widgetID, _ := strconv.Atoi(r.Form.Get("product_id")) // not applicable for virtual-terminal

	txnData, err = app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create a new customer
	customerID, err := app.SaveCustomer(txnData.FirstName, txnData.LastName, txnData.Email)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create a new transaction
	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		TransactionStatusID: models.TransactionCleared,
		PaymentIntentID:     txnData.PaymentIntentID,
		PaymentMethodID:     txnData.PaymentMethodID,
	}

	txnID, err := app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create a new order
	order := models.Order{
		WidgetID:      widgetID,
		TransactionID: txnID,
		CustomerID:    customerID,
		StatusID:      models.OrderCleared,
		Quantity:      1,
		Amount:        txnData.PaymentAmount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// write this data to session, and then redirect user to new page
	txnData.Type = ReceiptBuyOnce
	app.Session.Put(r.Context(), "receipt", txnData)

	http.Redirect(w, r, "/receipt", http.StatusSeeOther)
}

/*
*******************
Handler for "/virtual-payment-succeeded" route
VirtualTerminalPaymentSucceeded displays the receipt page for virtual terminal transactions
********************
*/
func (app *application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// create a new transaction
	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		TransactionStatusID: models.TransactionCleared,
		PaymentIntentID:     txnData.PaymentIntentID,
		PaymentMethodID:     txnData.PaymentMethodID,
	}

	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// write this data to session, and then redirect user to new page
	txnData.Type = ReceiptVirtual
	app.Session.Put(r.Context(), "receipt", txnData)

	http.Redirect(w, r, "/receipt", http.StatusSeeOther)
}

/*
*******************
Handler for "/widgets/{id}" route
ChargeOnce displays the page to buy one widget
********************
*/
func (app *application) ChargeOnce(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	if err := app.renderTemplate(w, r, "buy-once", &templateData{Data: data}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}

/*
*******************
Handler for "/plans/bronze" route
********************
*/
func (app *application) BronzePlan(w http.ResponseWriter, r *http.Request) {
	widget, err := app.DB.GetWidget(4)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	if err := app.renderTemplate(w, r, "bronze-plan", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err)
	}
}

// BronzePlanReceipt displays the receipt for bronze plans
func (app *application) BronzePlanReceipt(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "receipt-plan", &templateData{}); err != nil {
		app.errorLog.Print(err)
	}
}

/***************************************
ROUTES BLOCK - END
****************************************/

/***************************************
UTILITIES BLOCK - START
****************************************/

type TransactionData struct {
	FirstName       string
	LastName        string
	Email           string
	CardHolderName  string
	PaymentIntentID string
	PaymentMethodID string
	PaymentAmount   int
	PaymentCurrency string
	LastFour        string
	ExpiryMonth     int
	ExpiryYear      int
	BankReturnCode  string
	Type            ReceiptType
}

/*
*******************
SaveCustomer saves a customer and returns id
********************
*/
func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	id, err := app.DB.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}
	return id, nil
}

/*
*******************
SaveTransaction saves a txn and returns id
********************
*/
func (app *application) SaveTransaction(txn models.Transaction) (int, error) {
	id, err := app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}
	return id, nil
}

/*
*******************
SaveOrder saves a order and returns id
********************
*/
func (app *application) SaveOrder(order models.Order) (int, error) {
	id, err := app.DB.InsertOrder(order)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Receipt displays a receipt
func (app *application) Receipt(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn
	data["ReceiptVirtual"] = ReceiptVirtual
	data["ReceiptBuyOnce"] = ReceiptBuyOnce

	app.Session.Remove(r.Context(), "receipt")
	if err := app.renderTemplate(w, r, "receipt", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

// GetTransactionData gets txn data from post and stripe
func (app *application) GetTransactionData(r *http.Request) (TransactionData, error) {
	var txnData TransactionData

	err := r.ParseForm()

	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	// read posted data
	firstName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	email := r.Form.Get("email")
	cardHolderName := r.Form.Get("card_holder_name")
	paymentIntentId := r.Form.Get("payment_intent_id")
	paymentMethodId := r.Form.Get("payment_method_id")
	paymentAmount, _ := strconv.Atoi(r.Form.Get("payment_amount"))
	paymentCurrency := r.Form.Get("payment_currency")

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	pi, err := card.RetrievePaymentIntent(paymentIntentId)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	pm, err := card.GetPaymentMethod(paymentMethodId)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	txnData = TransactionData{
		FirstName:       firstName,
		LastName:        lastName,
		Email:           email,
		CardHolderName:  cardHolderName,
		PaymentIntentID: paymentIntentId,
		PaymentMethodID: paymentMethodId,
		PaymentAmount:   paymentAmount,
		PaymentCurrency: paymentCurrency,
		LastFour:        lastFour,
		ExpiryMonth:     int(expiryMonth),
		ExpiryYear:      int(expiryYear),
		BankReturnCode:  pi.Charges.Data[0].ID,
	}
	return txnData, nil
}

// LoginPage displays the login page
func (app *application) LoginPage(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "login", &templateData{}); err != nil {
		app.errorLog.Print(err)
	}
}

// PostLoginPage handles the posted login form
func (app *application) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	app.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	id, err := app.DB.Authenticate(email, password)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	app.Session.Put(r.Context(), "userID", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout logs the user out
func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	app.Session.Destroy(r.Context())
	app.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ForgotPassword shows the forgot password page
func (app *application) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "forgot-password", &templateData{}); err != nil {
		app.errorLog.Print(err)
	}
}

/***************************************
UTILITIES BLOCK - START
****************************************/
