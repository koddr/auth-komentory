package routes

import (
	"Komentory/auth/pkg/helpers"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/Komentory/utilities"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestPrivateRoutes(t *testing.T) {
	// Load .env.test file from the root folder
	if err := godotenv.Load("../../.env.test"); err != nil {
		panic(err)
	}

	// Define test variables.
	tokens, _ := helpers.GenerateNewTokens(uuid.New().String(), utilities.RoleNameUser)
	body := map[string]string{
		"empty":     `{}`,
		"non-empty": `{"first_name": "Bob"}`,
	}

	// Define a structure for specifying input and output data of a single test case.
	tests := []struct {
		description  string
		httpMethod   string
		route        string // input route
		tokenString  string
		body         io.Reader
		expectedCode int
	}{
		// Failed test cases:
		{
			"fail: update user attrs without JWT",
			"PATCH", "/v1/user/update/attrs", "", nil,
			400, // Missing or malformed JWT
		},
		{
			"fail: update user attrs without JSON body",
			"PATCH", "/v1/user/update/attrs", tokens.Access, nil,
			400, // unexpected end of JSON input
		},
		{
			"fail: update user attrs with empty JSON body",
			"PATCH", "/v1/user/update/attrs", tokens.Access, bytes.NewBuffer([]byte(body["empty"])),
			400, // validation errors
		},
		{
			"fail: update user attrs with JSON body, but user not found in DB",
			"PATCH", "/v1/user/update/attrs", tokens.Access, bytes.NewBuffer([]byte(body["non-empty"])),
			404, // sql: no rows in result set
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
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", test.tokenString))
		req.Header.Set("Content-Type", "application/json")

		// Perform the request plain with the app.
		resp, _ := app.Test(req, -1) // the -1 disables request latency

		// Parse the response body.
		body, errReadAll := io.ReadAll(resp.Body)
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
