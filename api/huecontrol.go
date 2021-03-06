package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gopkg.in/validator.v2"

	"github.com/vincentcr/huecontrol/services"
)

func main() {
	svc, err := services.New("dev")
	if err != nil {
		panic(err.Error())
	}

	setupServer(svc)
}

func setupServer(svc *services.Services) {
	m := NewMux(svc)
	setupMiddlewares(m)
	routeUsers(m)
	m.Serve()
}

func setupMiddlewares(m *Mux) {
	m.Use(cors)
	m.Use(authenticate)
}

func cors(c *HCContext, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization,Accept,Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, HEAD")
	return nil
}

type UserRequest struct {
	Email    string `validate:"nonzero,regexp=^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[[:alnum:]]{2,}$"`
	Password string `validate:"nonzero,min=6"`
}

func routeUsers(m *Mux) {

	m.Post("/api/1.0.0/users", func(c *HCContext, w http.ResponseWriter, r *http.Request) error {
		var userReq UserRequest
		if err := parseAndValidate(r, &userReq); err != nil {
			return err
		}

		user, err := c.Services.Users.Create(userReq.Email, userReq.Password)
		if err == services.ErrUniqueViolation {
			return HttpError{StatusCode: 400, StatusText: "User already exists"}
		} else if err != nil {
			return err
		}

		token, err := c.Services.Users.CreateAccessToken(user)
		if err != nil {
			return err
		}

		res := map[string]interface{}{
			"user":  user,
			"token": token,
		}

		return jsonify(res, w)
	})

	m.Post("/api/1.0.0/users/tokens", mustAuthenticate(func(c *HCContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		token, err := c.Services.Users.CreateAccessToken(user)
		if err != nil {
			return err
		}

		res := map[string]interface{}{
			"user":  user,
			"token": token,
		}

		return jsonify(res, w)
	}))

	m.Get("/api/1.0.0/users/me", mustAuthenticate(func(c *HCContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		return jsonify(user, w)
	}))
}

func parseAndValidate(r *http.Request, result interface{}) error {
	if err := parseBody(r, result); err != nil {
		return NewHttpError(http.StatusBadRequest)
	}

	fmt.Printf("request: %#v\n", result)

	if err := validator.Validate(result); err != nil {
		return NewHttpError(http.StatusBadRequest)
	}

	return nil
}

func parseBody(r *http.Request, result interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(result)
}

func jsonify(result interface{}, w http.ResponseWriter) error {
	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return writeAs(w, "application/json", bytes)
}

func writeAs(w http.ResponseWriter, contentType string, bytes []byte) error {
	w.Header().Set("content-type", contentType)
	_, err := w.Write(bytes)

	return err
}
