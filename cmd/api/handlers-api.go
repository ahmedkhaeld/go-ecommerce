package main

import (
	"encoding/json"
	"fmt"
	"github.com/ahmedkhaeld/ecommerce/internal/cards"
	"github.com/ahmedkhaeld/ecommerce/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v72"
	"net/http"
	"strconv"
	"time"
)

type stripePayload struct {
	Currency      string `json:"currency"`
	Amount        string `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	Email         string `json:"email"`
	CardBrand     string `json:"card_brand"`
	ExpiryMonth   int    `json:"exp_month"`
	ExpiryYear    int    `json:"exp_year"`
	LastFour      string `json:"last_four"`
	Plan          string `json:"plan"`
	ProductID     string `json:"product_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	ID      string `json:"id,omitempty"`
}

func (app *application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	// 1. get the body of the request and decode it to payload type with its fields currency and amount
	var payload stripePayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	// typecast amount to int
	amount, err := strconv.Atoi(payload.Amount)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	// now currency and amount are ready

	// card instance of Card, also has access to Card method
	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: payload.Currency,
	}
	okay := true

	// charge the card
	pi, msg, err := card.Charge(payload.Currency, amount)
	if err != nil {
		okay = false
	}

	// if payment intent success, then send back actual payment intent, else send back json response with false
	if okay {
		out, err := json.MarshalIndent(pi, "", "  ")
		if err != nil {
			app.errorLog.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)

	} else {
		j := jsonResponse{
			OK:      false,
			Message: msg,
			Content: "",
		}

		out, err := json.MarshalIndent(j, "", "  ")
		if err != nil {
			app.errorLog.Println(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	}

}

func (app *application) GetWidgetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	out, err := json.MarshalIndent(widget, "", "  ")
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (app *application) CreateCustomerAndSubscriptionPlan(w http.ResponseWriter, r *http.Request) {
	// 1. deal with the payload that passed to us
	var data stripePayload
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	app.infoLog.Println(data.Email, data.LastFour, data.PaymentMethod, data.LastFour)

	// card instance of Card, also has access to Card method
	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: data.Currency,
	}

	okay := true
	var subscription *stripe.Subscription
	txnMsg := "Transaction successful"

	// call the customer method
	stripeCustomer, msg, err := card.CreateCustomer(data.PaymentMethod, data.Email)
	if err != nil {
		app.errorLog.Println(err)
		okay = false
		txnMsg = msg
	}

	if okay {
		subscription, err = card.SubscribeToPlan(stripeCustomer, data.Plan, data.Email, data.LastFour, "")
		if err != nil {
			app.errorLog.Println(err)
			okay = false
			txnMsg = msg
		}

		app.infoLog.Println("subscription id is:", subscription.ID)
	}

	// if okay true then create and save customer, and create a new txn, last, create an order
	if okay {
		productID, _ := strconv.Atoi(data.ProductID)
		customerID, err := app.SaveCustomer(data.FirstName, data.LastName, data.Email)
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		// create  a new transaction
		amount, _ := strconv.Atoi(data.Amount)

		txn := models.Transaction{
			Amount:              amount,
			Currency:            "cad",
			LastFour:            data.LastFour,
			ExpiryMonth:         data.ExpiryMonth,
			ExpiryYear:          data.ExpiryYear,
			TransactionStatusID: 2,
		}
		txnID, err := app.SaveTransaction(txn)
		if err != nil {
			app.errorLog.Println(err)
			return
		}
		// create order
		order := models.Order{
			WidgetID:      productID,
			TransactionID: txnID,
			CustomerID:    customerID,
			StatusID:      1,
			Quantity:      1,
			Amount:        amount,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		_, err = app.SaveOrder(order)
		if err != nil {
			app.errorLog.Println(err)
			return
		}

	}

	resp := jsonResponse{
		OK:      okay,
		Message: txnMsg,
	}

	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)

}

// SaveCustomer saves a customer and returns id
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

// SaveTransaction saves a txn to db and returns id
func (app *application) SaveTransaction(txn models.Transaction) (int, error) {

	id, err := app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// SaveOrder saves an order to db and returns id
func (app *application) SaveOrder(o models.Order) (int, error) {

	id, err := app.DB.InsertOrder(o)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// CreateAuthToken here getting  username & password as json
func (app *application) CreateAuthToken(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &userInput)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}
	// get the user from the database by email; send error if invalid email(invalid credentials)
	user, err := app.DB.GetUserByEmail(userInput.Email)
	if err != nil {
		app.invalidCredentials(w)
		return
	}

	// validate the password; send error if invalid password (compare what user enters against the hash stored in the db)
	validPassword, err := app.passwordMatches(user.Password, userInput.Password)
	if err != nil {
		app.invalidCredentials(w)
		return
	}

	if !validPassword {
		app.invalidCredentials(w)
		return
	}

	// generate the token to be sent back as part of the response
	token, err := models.GenerateToken(user.ID, 24*time.Hour, models.ScopeAuthentication)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// send response

	var payload struct {
		Error   bool          `json:"error"`
		Message string        `json:"message"`
		Token   *models.Token `json:"authentication_token"`
	}

	payload.Error = false
	payload.Message = fmt.Sprintf("token for %s created", userInput.Email)
	payload.Token = token

	_ = app.writeJSON(w, http.StatusOK, payload)

}
