package server

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"VladBag2022/gophermart/internal/storage"
	"VladBag2022/gophermart/mocks"
)

func makeTestRequest(
	t *testing.T,
	ts *httptest.Server,
	method, path, contentType, authorization string,
	body io.Reader,
) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", contentType)

	if len(authorization) > 0 {
		req.Header.Set(authorizationHeader, authorization)
	}

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

func getTestEntities(
	addExpectationsFunc func(repository *mocks.Repository),
) (*Server, *httptest.Server) {
	config, err := NewConfig()
	if err != nil {
		return nil, nil
	}
	repository := new(mocks.Repository)
	addExpectationsFunc(repository)
	server := NewServer(repository, config)
	router := rootRouter(server)
	return &server, httptest.NewServer(router)
}

func TestServer_register(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name        string
		logins      []string
		contentType string
		content     string
		want        want
	}{
		{
			name:        "positive test - some logins",
			logins:      []string{"a", "b"},
			contentType: contentTypeJSON,
			content:     "{\"login\": \"c\",\"password\": \"123\"}",
			want: want{
				statusCode: 200,
			},
		},
		{
			name:        "positive test - no logins",
			logins:      []string{},
			contentType: contentTypeJSON,
			content:     "{\"login\": \"c\",\"password\": \"123\"}",
			want: want{
				statusCode: 200,
			},
		},
		{
			name:        "negative test - empty login",
			logins:      []string{},
			contentType: contentTypeJSON,
			content:     "{\"login\": \"\",\"password\": \"123\"}",
			want: want{
				statusCode: 400,
			},
		},
		{
			name:        "negative test - empty password",
			logins:      []string{},
			contentType: contentTypeJSON,
			content:     "{\"login\": \"c\",\"password\": \"\"}",
			want: want{
				statusCode: 400,
			},
		},
		{
			name:        "negative test - empty login and password",
			logins:      []string{},
			contentType: contentTypeJSON,
			content:     "{\"login\": \"\",\"password\": \"\"}",
			want: want{
				statusCode: 400,
			},
		},
		{
			name:        "negative test - duplicate login",
			logins:      []string{"a", "b"},
			contentType: contentTypeJSON,
			content:     "{\"login\": \"a\",\"password\": \"123\"}",
			want: want{
				statusCode: 409,
			},
		},
		{
			name:        "negative test - wrong content type",
			logins:      []string{},
			contentType: "text/plain; charset=utf-8",
			content:     "{\"login\": \"\",\"password\": \"\"}",
			want: want{
				statusCode: 400,
			},
		},
		{
			name:        "negative test - malformed content",
			logins:      []string{"a", "b"},
			contentType: contentTypeJSON,
			content:     "{\"login\": \"a\",\"password\": \"12",
			want: want{
				statusCode: 400,
			},
		},
		{
			name:        "negative test - wrong content",
			logins:      []string{"a", "b"},
			contentType: contentTypeJSON,
			content:     "{\"user\": \"a\",\"pass\": \"12\"}",
			want: want{
				statusCode: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ts := getTestEntities(func(repository *mocks.Repository) {
				for _, login := range tt.logins {
					repository.On("IsLoginAvailable",
						mock.Anything, login).Return(false, nil)
				}
				repository.On("IsLoginAvailable",
					mock.Anything, mock.Anything).Return(true, nil)
				repository.On("Register",
					mock.Anything, mock.Anything, mock.Anything).Return(nil)
			})
			require.NotNil(t, ts)
			defer ts.Close()

			response, _ := makeTestRequest(t, ts, http.MethodPost, "/api/user/register", tt.contentType, "",
				strings.NewReader(tt.content))
			err := response.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, response.StatusCode)
		})
	}
}

func TestServer_login(t *testing.T) {
	type want struct {
		statusCode int
	}
	type user struct {
		login    string
		password string
	}
	tests := []struct {
		name        string
		users       []user
		contentType string
		content     string
		want        want
	}{
		{
			name: "positive test",
			users: []user{
				{"a", "123"},
				{"b", "123"},
				{"c", "123"},
			},
			contentType: contentTypeJSON,
			content:     "{\"login\": \"c\",\"password\": \"123\"}",
			want: want{
				statusCode: 200,
			},
		},
		{
			name: "negative test - wrong password",
			users: []user{
				{"a", "123"},
				{"b", "123"},
				{"c", "123"},
			},
			contentType: contentTypeJSON,
			content:     "{\"login\": \"c\",\"password\": \"1234\"}",
			want: want{
				statusCode: 401,
			},
		},
		{
			name: "negative test - wrong login",
			users: []user{
				{"a", "123"},
				{"b", "123"},
				{"c", "123"},
			},
			contentType: contentTypeJSON,
			content:     "{\"login\": \"cd\",\"password\": \"1234\"}",
			want: want{
				statusCode: 401,
			},
		},
		{
			name:        "negative test - malformed content",
			users:       []user{},
			contentType: contentTypeJSON,
			content:     "{\"login\": \"cd\",\"password\": \"12",
			want: want{
				statusCode: 400,
			},
		},
		{
			name:        "negative test - wrong content type",
			users:       []user{},
			contentType: "text",
			content:     "{\"login\": \"cd\",\"password\": \"1234\"}",
			want: want{
				statusCode: 400,
			},
		},
		{
			name:        "negative test - wrong content",
			users:       []user{},
			contentType: contentTypeJSON,
			content:     "{\"user\": \"cd\",\"pass\": \"1234\"}",
			want: want{
				statusCode: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ts := getTestEntities(func(repository *mocks.Repository) {
				for _, tu := range tt.users {
					repository.On("Login",
						mock.Anything, tu.login, tu.password).Return(true, nil)
				}
				repository.On("Login",
					mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
			})
			require.NotNil(t, ts)
			defer ts.Close()

			response, _ := makeTestRequest(t, ts, http.MethodPost, "/api/user/login", tt.contentType, "",
				strings.NewReader(tt.content))
			err := response.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, response.StatusCode)
		})
	}
}

func TestServer_upload(t *testing.T) {
	type want struct {
		statusCode int
	}

	tests := []struct {
		name        string
		userOrders  map[string][]int64
		user        string
		contentType string
		content     string
		want        want
	}{
		{
			name: "positive test - order already uploaded",
			userOrders: map[string][]int64{
				"a": {123456789031, 566165445},
				"b": {9579343, 58568287791534},
			},
			user:        "a",
			contentType: "text/plain",
			content:     "123456789031",
			want: want{
				statusCode: 200,
			},
		},
		{
			name: "positive test - new order",
			userOrders: map[string][]int64{
				"a": {123456789031, 566165445},
				"b": {9579343, 58568287791534},
			},
			user:        "a",
			contentType: "text/plain",
			content:     "12345678903",
			want: want{
				statusCode: 202,
			},
		},
		{
			name: "negative test - wrong content type",
			userOrders: map[string][]int64{
				"a": {123456789031, 566165445},
				"b": {9579343, 58568287791534},
			},
			user:        "a",
			contentType: contentTypeJSON,
			content:     "123456789032",
			want: want{
				statusCode: 400,
			},
		},
		{
			name: "negative test - wrong content",
			userOrders: map[string][]int64{
				"a": {123456789031, 566165445},
				"b": {9579343, 58568287791534},
			},
			user:        "a",
			contentType: "text/plain",
			content:     "gg",
			want: want{
				statusCode: 422,
			},
		},
		{
			name: "negative test - unauthorized",
			userOrders: map[string][]int64{
				"a": {123456789031, 566165445},
				"b": {9579343, 58568287791534},
			},
			user:        "",
			contentType: "text/plain",
			content:     "123",
			want: want{
				statusCode: 401,
			},
		},
		{
			name: "negative test - order already uploaded by another user",
			userOrders: map[string][]int64{
				"a": {123456789031, 566165445},
				"b": {9579343, 58568287791534},
			},
			user:        "b",
			contentType: "text/plain",
			content:     "123456789031",
			want: want{
				statusCode: 409,
			},
		},
		{
			name: "negative test - bad Luhn check",
			userOrders: map[string][]int64{
				"a": {123456789031, 566165445},
				"b": {9579343, 58568287791534},
			},
			user:        "b",
			contentType: "text/plain",
			content:     "123456789035",
			want: want{
				statusCode: 422,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderUploaded := false
			userRegistered := false
			s, ts := getTestEntities(func(repository *mocks.Repository) {
				for tUser, tOrders := range tt.userOrders {
					for _, tOrder := range tOrders {
						repository.On("OrderOwner", mock.Anything, tOrder).Return(tUser, nil)
						if strconv.FormatInt(tOrder, 10) == tt.content {
							orderUploaded = true
						}
					}
					if tUser == tt.user {
						userRegistered = true
					}
				}
				if !orderUploaded {
					order, err := strconv.ParseInt(tt.content, 10, 64)
					if err == nil {
						repository.On("OrderOwner", mock.Anything, order).Return("", nil)
					}
				}
				repository.On("UploadOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			})
			require.NotNil(t, ts)
			defer ts.Close()

			h := ""
			if userRegistered {
				nh, err := getAuthHeader(*s, tt.user)
				require.NoError(t, err)
				h = nh
			}

			response, _ := makeTestRequest(t, ts, http.MethodPost, "/api/user/orders", tt.contentType,
				h, strings.NewReader(tt.content))
			err := response.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, response.StatusCode)
		})
	}
}

func TestServer_list(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		content     bool
	}

	tests := []struct {
		name       string
		userOrders map[string]bool
		user       string
		want       want
	}{
		{
			name: "positive test - data is presented",
			userOrders: map[string]bool{
				"a": true,
				"b": false,
			},
			user: "a",
			want: want{
				statusCode:  200,
				contentType: contentTypeJSON,
				content:     true,
			},
		},
		{
			name: "positive test - no data",
			userOrders: map[string]bool{
				"a": true,
				"b": false,
			},
			user: "b",
			want: want{
				statusCode:  204,
				contentType: contentTypeJSON,
				content:     true,
			},
		},
		{
			name: "negative test - unauthorized",
			userOrders: map[string]bool{
				"a": true,
				"b": false,
			},
			user: "c",
			want: want{
				statusCode:  401,
				contentType: "",
				content:     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRegistered := false
			s, ts := getTestEntities(func(repository *mocks.Repository) {
				for tUser, tOrder := range tt.userOrders {
					if tOrder {
						repository.On("Orders", mock.Anything, tUser).Return([]storage.OrderInfo{
							{
								Number:  123,
								Accrual: 0.0,
							},
						}, nil)
					} else {
						repository.On("Orders", mock.Anything, tUser).Return([]storage.OrderInfo{}, nil)
					}
					if tUser == tt.user {
						userRegistered = true
					}
				}
			})
			require.NotNil(t, ts)
			defer ts.Close()

			h := ""
			if userRegistered {
				nh, err := getAuthHeader(*s, tt.user)
				require.NoError(t, err)
				h = nh
			}

			response, content := makeTestRequest(t, ts, http.MethodGet, "/api/user/orders", "", h, nil)
			err := response.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, response.StatusCode)
			assert.Equal(t, tt.want.contentType, response.Header.Get("Content-Type"))
			t.Log(content)
			assert.Equal(t, tt.want.content, len(content) > 0)
		})
	}
}

func TestServer_balance(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		content     bool
	}

	tests := []struct {
		name  string
		users []string
		user  string
		want  want
	}{
		{
			name:  "positive test",
			users: []string{"a", "b"},
			user:  "a",
			want: want{
				statusCode:  200,
				contentType: contentTypeJSON,
				content:     true,
			},
		},
		{
			name:  "negative test - unauthorized",
			users: []string{"a", "b"},
			user:  "c",
			want: want{
				statusCode:  401,
				contentType: "",
				content:     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRegistered := false
			s, ts := getTestEntities(func(repository *mocks.Repository) {
				for _, tUser := range tt.users {
					repository.On("Balance", mock.Anything, tUser).Return(storage.BalanceInfo{
						Current:   0.0,
						Withdrawn: 0.0,
					}, nil)
					if tUser == tt.user {
						userRegistered = true
					}
				}
			})
			require.NotNil(t, ts)
			defer ts.Close()

			h := ""
			if userRegistered {
				nh, err := getAuthHeader(*s, tt.user)
				require.NoError(t, err)
				h = nh
			}

			response, content := makeTestRequest(t, ts, http.MethodGet, "/api/user/balance", "", h, nil)
			err := response.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, response.StatusCode)
			assert.Equal(t, tt.want.contentType, response.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.content, len(content) > 0)
		})
	}
}

func TestServer_withdraw(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name         string
		userBalances map[string]float64
		user         string
		contentType  string
		content      string
		want         want
	}{
		{
			name: "positive test",
			userBalances: map[string]float64{
				"a": 100.0,
				"b": 200.0,
			},
			user:        "a",
			contentType: contentTypeJSON,
			content:     "{\"order\": \"2377225624\",\"sum\": 10}",
			want: want{
				statusCode: 200,
			},
		},
		{
			name: "negative test - malformed content",
			userBalances: map[string]float64{
				"a": 100.0,
				"b": 200.0,
			},
			user:        "a",
			contentType: contentTypeJSON,
			content:     "{\"login\": \"cd\",\"password\": \"12",
			want: want{
				statusCode: 400,
			},
		},
		{
			name: "negative test - wrong content type",
			userBalances: map[string]float64{
				"a": 100.0,
				"b": 200.0,
			},
			user:        "a",
			contentType: "text",
			content:     "{\"order\": \"cd\",\"sum\": 1234}",
			want: want{
				statusCode: 400,
			},
		},
		{
			name: "negative test - wrong content",
			userBalances: map[string]float64{
				"a": 100.0,
				"b": 200.0,
			},
			user:        "a",
			contentType: contentTypeJSON,
			content:     "{\"user\": \"cd\",\"pass\": 1234}",
			want: want{
				statusCode: 400,
			},
		},
		{
			name: "negative test - unauthorized",
			userBalances: map[string]float64{
				"a": 100,
				"b": 200,
			},
			user:        "c",
			contentType: contentTypeJSON,
			content:     "{\"order\": \"2377225624\",\"sum\": 10}",
			want: want{
				statusCode: 401,
			},
		},
		{
			name: "negative test - low balance",
			userBalances: map[string]float64{
				"a": 100,
				"b": 200,
			},
			user:        "a",
			contentType: contentTypeJSON,
			content:     "{\"order\": \"2377225624\",\"sum\": 1000}",
			want: want{
				statusCode: 402,
			},
		},
		{
			name: "negative test - wrong order number",
			userBalances: map[string]float64{
				"a": 100,
				"b": 200,
			},
			user:        "a",
			contentType: contentTypeJSON,
			content:     "{\"order\": \"2377225626\",\"sum\": 10}",
			want: want{
				statusCode: 422,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRegistered := false
			s, ts := getTestEntities(func(repository *mocks.Repository) {
				for tUser, tBalance := range tt.userBalances {
					repository.On("Balance", mock.Anything, tUser).Return(storage.BalanceInfo{
						Current:   tBalance,
						Withdrawn: 0.0,
					}, nil)
					if tUser == tt.user {
						userRegistered = true
					}
				}
				repository.On("Withdraw", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			})
			require.NotNil(t, ts)
			defer ts.Close()

			h := ""
			if userRegistered {
				nh, err := getAuthHeader(*s, tt.user)
				require.NoError(t, err)
				h = nh
			}

			response, _ := makeTestRequest(t, ts, http.MethodPost, "/api/user/balance/withdraw", tt.contentType, h,
				strings.NewReader(tt.content))
			err := response.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, response.StatusCode)
		})
	}
}

func TestServer_withdrawals(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		content     bool
	}
	tests := []struct {
		name            string
		userWithdrawals map[string]bool
		user            string
		want            want
	}{
		{
			name: "positive test - withdrawals found",
			userWithdrawals: map[string]bool{
				"a": true,
				"b": false,
			},
			user: "a",
			want: want{
				statusCode:  200,
				contentType: contentTypeJSON,
				content:     true,
			},
		},
		{
			name: "positive test - no withdrawals",
			userWithdrawals: map[string]bool{
				"a": true,
				"b": false,
			},
			user: "b",
			want: want{
				statusCode:  204,
				contentType: contentTypeJSON,
				content:     true,
			},
		},
		{
			name: "negative test - unauthorized",
			userWithdrawals: map[string]bool{
				"a": true,
				"b": false,
			},
			user: "c",
			want: want{
				statusCode:  401,
				contentType: "",
				content:     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRegistered := false
			s, ts := getTestEntities(func(repository *mocks.Repository) {
				for tUser, tWithdrawal := range tt.userWithdrawals {
					if tWithdrawal {
						repository.On("Withdrawals", mock.Anything, tUser).Return([]storage.WithdrawalInfo{
							{
								Order: 123,
								Sum:   10,
							},
						}, nil)
					} else {
						repository.On("Withdrawals", mock.Anything, tUser).Return([]storage.WithdrawalInfo{}, nil)
					}
					if tUser == tt.user {
						userRegistered = true
					}
				}
				repository.On("Withdraw", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			})
			require.NotNil(t, ts)
			defer ts.Close()

			h := ""
			if userRegistered {
				nh, err := getAuthHeader(*s, tt.user)
				require.NoError(t, err)
				h = nh
			}

			response, content := makeTestRequest(t, ts, http.MethodGet, "/api/user/withdrawals", "", h, nil)
			err := response.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, response.StatusCode)
			assert.Equal(t, tt.want.contentType, response.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.content, len(content) > 0)
		})
	}
}
