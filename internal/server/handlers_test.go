package server

import (
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

func getTestServer() *httptest.Server {
	return &httptest.Server{}
}

func TestServer_shorten(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		response    bool
	}
	tests := []struct {
		name    string
		content string
		want    want
	}{
		{
			name:    "positive test",
			content: "https://example.com",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  201,
				response:    true,
			},
		},
		{
			name:    "negative test",
			content: "",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				response:    true,
			},
		},
	}

	ts := getTestServer()
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, _ := makeTestRequest(t, ts, http.MethodPost, "/", strings.NewReader(tt.content))
			err := response.Body.Close()
			require.NoError(t, err)
		})
	}
}
