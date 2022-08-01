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

func getTestServer(logins []string) *httptest.Server {
	return &httptest.Server{}
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
			name:    		"positive test with some logins",
			logins: 		[]string{"a", "b"},
			contentType: 	"application/json",
			content: 		"{\"login\": \"c\",\"password\": \"123\"}",
			want: want{
				statusCode:  200,
			},
		},
		{
			name:    		"positive test without any logins",
			logins: 		[]string{},
			contentType: 	"application/json",
			content: 		"{\"login\": \"c\",\"password\": \"123\"}",
			want: want{
				statusCode:  200,
			},
		},
		{
			name:    		"negative test with empty login",
			logins: 		[]string{},
			contentType: 	"application/json",
			content: 		"{\"login\": \"\",\"password\": \"123\"}",
			want: want{
				statusCode:  400,
			},
		},
		{
			name:    		"negative test with empty password",
			logins: 		[]string{},
			contentType: 	"application/json",
			content: 		"{\"login\": \"c\",\"password\": \"\"}",
			want: want{
				statusCode:  400,
			},
		},
		{
			name:    		"negative test with empty login and password",
			logins: 		[]string{},
			contentType: 	"application/json",
			content: 		"{\"login\": \"\",\"password\": \"\"}",
			want: want{
				statusCode:  400,
			},
		},
		{
			name:    		"negative test with duplicate login",
			logins: 		[]string{"a", "b"},
			contentType: 	"application/json",
			content: 		"{\"login\": \"a\",\"password\": \"123\"}",
			want: want{
				statusCode:  409,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := getTestServer(tt.logins)
			defer ts.Close()

			response, _ := makeTestRequest(t, ts, http.MethodPost, "/api/user/register",
				strings.NewReader(tt.content))
			err := response.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, response.StatusCode)
		})
	}
}
