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
	bodyCodeEmpty := bytes.NewBuffer([]byte(`{"code": ""}`))
	bodyCodeWrong := bytes.NewBuffer([]byte(`{"code": "123456"}`))

	// Define a structure for specifying input and output data of a single test case.
	tests := []struct {
		description   string
		route         string // input route
		httpMethod    string
		body          io.Reader
		expectedError bool
		expectedCode  int
	}{
		// Failed test cases:
		{
			description:   "fail: apply activation code with no request JSON body",
			route:         "/v1/account/activate",
			httpMethod:    "PATCH",
			body:          nil,
			expectedError: false,
			expectedCode:  400,
		},
		{
			description:   "fail: apply activation code with empty code string (not found) in request JSON body",
			route:         "/v1/account/activate",
			httpMethod:    "PATCH",
			body:          bodyCodeEmpty,
			expectedError: false,
			expectedCode:  404,
		},
		{
			description:   "fail: apply activation code with wrong code string in request JSON body",
			route:         "/v1/account/activate",
			httpMethod:    "PATCH",
			body:          bodyCodeWrong,
			expectedError: false,
			expectedCode:  404,
		},
		{
			description:   "fail: apply reset code without request JSON body",
			route:         "/v1/account/activate",
			httpMethod:    "PATCH",
			body:          nil,
			expectedError: false,
			expectedCode:  400,
		},
		{
			description:   "fail: apply reset code with empty code string (not found) in request JSON body",
			route:         "/v1/password/reset",
			httpMethod:    "PATCH",
			body:          bodyCodeEmpty,
			expectedError: false,
			expectedCode:  400,
		},
		{
			description:   "fail: apply reset code with wrong code string in request JSON body",
			route:         "/v1/password/reset",
			httpMethod:    "PATCH",
			body:          bodyCodeWrong,
			expectedError: false,
			expectedCode:  400,
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
		resp, err := app.Test(req, -1) // the -1 disables request latency

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

		// Verify, that no error occurred, that is not expected
		assert.Equalf(t, test.expectedError, err != nil, test.description)

		// As expected errors lead to broken responses,
		// the next test case needs to be processed.
		if test.expectedError {
			continue
		}

		// Checking, if the JSON field "status" from the response body has the expected status code.
		assert.Equalf(t, test.expectedCode, status, description)
	}
}
