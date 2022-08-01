package server

import (
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func makeTestRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	response, err := client.Do(req)
	require.NoError(t, err)

	responseBody, err := ioutil.ReadAll(response.Body)
	require.NoError(t, err)

	err = response.Body.Close()
	require.NoError(t, err)

	return response, string(responseBody)
}

func TestServer_register(t *testing.T) {
	type want struct {
		statusCode  int
	}
	tests := []struct {
		name    	string
		logins		[]string
		contentType string
		content 	string
		want    	want
	}{
		{
			name:    		"positive test - some logins",
			logins: 		[]string{"a", "b"},
			contentType: 	"application/json",
			content: 		"{\"login\": \"c\",\"password\": \"123\"}",
			want: want{
				statusCode:  200,
			},
		},
		{
			name:    		"positive test - no logins",
			logins: 		[]string{},
			contentType: 	"application/json",
			content: 		"{\"login\": \"c\",\"password\": \"123\"}",
			want: want{
				statusCode:  200,
			},
		},
		{
			name:    		"negative test - empty login",
			logins: 		[]string{},
			contentType: 	"application/json",
			content: 		"{\"login\": \"\",\"password\": \"123\"}",
			want: want{
				statusCode:  400,
			},
		},
		{
			name:    		"negative test - empty password",
			logins: 		[]string{},
			contentType: 	"application/json",
			content: 		"{\"login\": \"c\",\"password\": \"\"}",
			want: want{
				statusCode:  400,
			},
		},
		{
			name:    		"negative test - empty login and password",
			logins: 		[]string{},
			contentType: 	"application/json",
			content: 		"{\"login\": \"\",\"password\": \"\"}",
			want: want{
				statusCode:  400,
			},
		},
		{
			name:    		"negative test - duplicate login",
			logins: 		[]string{"a", "b"},
			contentType: 	"application/json",
			content: 		"{\"login\": \"a\",\"password\": \"123\"}",
			want: want{
				statusCode:  409,
			},
		},
		{
			name:    		"negative test - wrong content type",
			logins: 		[]string{},
			contentType: 	"text/plain; charset=utf-8",
			content: 		"{\"login\": \"\",\"password\": \"\"}",
			want: want{
				statusCode:  400,
			},
		},
		{
			name:    		"negative test - malformed content",
			logins: 		[]string{"a", "b"},
			contentType: 	"application/json",
			content: 		"{\"login\": \"a\",\"password\": \"12",
			want: want{
				statusCode:  400,
			},
		},
		{
			name:    		"negative test - wrong content",
			logins: 		[]string{"a", "b"},
			contentType: 	"application/json",
			content: 		"{\"user\": \"a\",\"pass\": \"12\"}",
			want: want{
				statusCode:  400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &httptest.Server{}
			defer ts.Close()

			response, _ := makeTestRequest(t, ts, http.MethodPost, "/api/user/register",
				strings.NewReader(tt.content))
			err := response.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, response.StatusCode)
		})
	}
}

func TestServer_login(t *testing.T) {
	type want struct {
		statusCode  int
	}
	type user struct {
		login 	 	string
		password 	string
	}
	tests := []struct {
		name    	string
		users 		[]user
		contentType string
		content 	string
		want    	want
	}{
		{
			name:    		"positive test",
			users: 			[]user{
				{"a", "123"},
				{"b", "123"},
				{"c", "123"},
			},
			contentType: 	"application/json",
			content: 		"{\"login\": \"c\",\"password\": \"123\"}",
			want: want{
				statusCode:  200,
			},
		},
		{
			name:    		"negative test - wrong password",
			users: 			[]user{
				{"a", "123"},
				{"b", "123"},
				{"c", "123"},
			},
			contentType: 	"application/json",
			content: 		"{\"login\": \"c\",\"password\": \"1234\"}",
			want: want{
				statusCode:  401,
			},
		},
		{
			name:    		"negative test - wrong login",
			users: 			[]user{
				{"a", "123"},
				{"b", "123"},
				{"c", "123"},
			},
			contentType: 	"application/json",
			content: 		"{\"login\": \"cd\",\"password\": \"1234\"}",
			want: want{
				statusCode:  401,
			},
		},
		{
			name:    		"negative test - malformed content",
			users: 			[]user{},
			contentType: 	"application/json",
			content: 		"{\"login\": \"cd\",\"password\": \"12",
			want: want{
				statusCode:  400,
			},
		},
		{
			name:    		"negative test - wrong content type",
			users: 			[]user{},
			contentType: 	"text",
			content: 		"{\"login\": \"cd\",\"password\": \"1234\"}",
			want: want{
				statusCode:  400,
			},
		},
		{
			name:    		"negative test - wrong content",
			users: 			[]user{},
			contentType: 	"application/json",
			content: 		"{\"user\": \"cd\",\"pass\": \"1234\"}",
			want: want{
				statusCode:  400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &httptest.Server{}
			defer ts.Close()

			response, _ := makeTestRequest(t, ts, http.MethodPost, "/api/user/login",
				strings.NewReader(tt.content))
			err := response.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, response.StatusCode)
		})
	}
}

func TestServer_orders(t *testing.T) {
	type want struct {
		statusCode  int
	}

	tests := []struct {
		name    	string
		userOrders 	map[string][]int
		user 		string
		contentType string
		content 	string
		want    	want
	}{
		{
			name:    		"positive test - order already uploaded",
			userOrders:		map[string][]int{
				"a": {123456789031, 566165445},
				"b": {9579343, 58568287791534},
			},
			user:			"a",
			contentType: 	"text/plain",
			content: 		"123456789031",
			want: want{
				statusCode:  200,
			},
		},
		{
			name:    		"positive test - new order",
			userOrders:		map[string][]int{
				"a": {123456789031, 566165445},
				"b": {9579343, 58568287791534},
			},
			user:			"a",
			contentType: 	"text/plain",
			content: 		"123456789032",
			want: want{
				statusCode:  200,
			},
		},
		{
			name:    		"negative test - wrong content type",
			userOrders:		map[string][]int{
				"a": {123456789031, 566165445},
				"b": {9579343, 58568287791534},
			},
			user:			"a",
			contentType: 	"application/json",
			content: 		"123456789032",
			want: want{
				statusCode:  400,
			},
		},
		{
			name:    		"negative test - wrong content",
			userOrders:		map[string][]int{
				"a": {123456789031, 566165445},
				"b": {9579343, 58568287791534},
			},
			user:			"a",
			contentType: 	"text/plain",
			content: 		"gg",
			want: want{
				statusCode:  400,
			},
		},
		{
			name:    		"negative test - unauthorized",
			userOrders:		map[string][]int{
				"a": {123456789031, 566165445},
				"b": {9579343, 58568287791534},
			},
			user:			"",
			contentType: 	"text/plain",
			content: 		"123",
			want: want{
				statusCode:  401,
			},
		},
		{
			name:    		"negative test - order already uploaded by another user",
			userOrders:		map[string][]int{
				"a": {123456789031, 566165445},
				"b": {9579343, 58568287791534},
			},
			user:			"b",
			contentType: 	"text/plain",
			content: 		"123456789031",
			want: want{
				statusCode:  409,
			},
		},
		{
			name:    		"negative test - bad Luhn check",
			userOrders:		map[string][]int{
				"a": {123456789031, 566165445},
				"b": {9579343, 58568287791534},
			},
			user:			"b",
			contentType: 	"text/plain",
			content: 		"123456789035",
			want: want{
				statusCode:  422,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &httptest.Server{}
			defer ts.Close()

			response, _ := makeTestRequest(t, ts, http.MethodPost, "/api/user/orders",
				strings.NewReader(tt.content))
			err := response.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, response.StatusCode)
		})
	}
}
