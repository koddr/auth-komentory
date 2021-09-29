package routes

import (
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

func TestPrivateRoutes(t *testing.T) {
	// Load .env.test file from the root folder
	if err := godotenv.Load("../../.env.test"); err != nil {
		panic(err)
	}

	// Define a structure for specifying input and output data of a single test case.
	tests := []struct {
		description   string
		route         string // input route
		httpMethod    string
		tokenString   string
		body          io.Reader
		expectedError bool
		expectedCode  int
	}{
		{
			description:   "fail: update password without JWT",
			route:         "/v1/user/update/attrs",
			httpMethod:    "PATCH",
			tokenString:   "",
			body:          nil,
			expectedError: false,
			expectedCode:  400, // "Missing or malformed JWT"
		},
	}

	// Define Fiber app.
	app := fiber.New()

	// Define routes.
	PrivateRoutes(app)

	// Iterate through test single test cases
	for index, test := range tests {
		// Create a new http request with the route from the test case.
		req := httptest.NewRequest(test.httpMethod, test.route, test.body)
		req.Header.Set("Authorization", test.tokenString)
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
