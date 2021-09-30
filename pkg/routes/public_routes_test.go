package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestPublicRoutes(t *testing.T) {
	// Load .env.test file from the root folder
	if err := godotenv.Load("../../.env.test"); err != nil {
		panic(err)
	}

	// Define test bodies for JSON request.
	body := map[string]string{
		"empty":     `{"code": ""}`,
		"not-empty": `{"code": "123456"}`,
	}

	// Define a structure for specifying input and output data of a single test case.
	tests := []struct {
		description  string
		httpMethod   string
		route        string // input route
		body         io.Reader
		expectedCode int
	}{
		// Failed test cases:
		{
			"fail: apply activation code with no JSON body",
			"PATCH", "/v1/account/activate", nil,
			400, // unexpected end of JSON input
		},
		{
			"fail: apply activation code with empty code string in JSON body",
			"PATCH", "/v1/account/activate", bytes.NewBuffer([]byte(body["empty"])),
			404, // sql: no rows in result set
		},
		{
			"fail: apply activation code with JSON body, but user not found in DB",
			"PATCH", "/v1/account/activate", bytes.NewBuffer([]byte(body["not-empty"])),
			404, // sql: no rows in result set
		},
		{
			"fail: apply reset code without JSON body",
			"PATCH", "/v1/password/reset", nil,
			400, // unexpected end of JSON input
		},
		{
			"fail: apply reset code with empty code string in JSON body",
			"PATCH", "/v1/password/reset", bytes.NewBuffer([]byte(body["empty"])),
			404, // sql: no rows in result set
		},
		{
			"fail: apply reset code with JSON body, but code not found in DB",
			"PATCH", "/v1/password/reset", bytes.NewBuffer([]byte(body["not-empty"])),
			404, // sql: no rows in result set
		},
	}

	// Define Fiber app.
	app := fiber.New()

	// Define routes.
	PublicRoutes(app)

	// Iterate through test single test cases
	for index, test := range tests {
		// Create a new http request with the route from the test case.
		req := httptest.NewRequest(test.httpMethod, test.route, test.body)
		req.Header.Set("Content-Type", "application/json")

		// Perform the request plain with the app.
		resp, _ := app.Test(req, -1) // the -1 disables request latency

		// Parse the response body.
		body, errReadAll := ioutil.ReadAll(resp.Body)
		if errReadAll != nil {
			return
		}

		// Set the response body (JSON) to simple map.
		var result map[string]interface{}
		if errUnmarshal := json.Unmarshal(body, &result); errUnmarshal != nil {
			return
		}

		// Redefine index of the test case.
		readableIndex := index + 1

		// Define status & description from the response.
		status := int(result["status"].(float64))
		description := fmt.Sprintf(
			"[%d] need to %s\nreal error output: %s",
			readableIndex, test.description, result["msg"].(string),
		)

		// Checking, if the JSON field "status" from the response body has the expected status code.
		assert.Equalf(t, test.expectedCode, status, description)
	}
}
