package accrual

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type orderInfoResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func (d Daemon) orderInfo(order int64) (info *orderInfoResponse, retryAfter int, err error) {
	var response *http.Response
	response, err = http.Get(fmt.Sprintf("%s/api/orders/%d", d.accrualAddress, order))
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		var content []byte
		content, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, 0, err
		}
		if err = json.Unmarshal(content, &info); err != nil {
			return nil, 0, err
		}
		return info, 0, nil
	case http.StatusTooManyRequests:
		header := response.Header.Get("Retry-After")
		retryAfter, err = strconv.Atoi(header)
		if err != nil {
			return nil, 0, err
		}
		return nil, retryAfter, nil
	case http.StatusNoContent:
		return nil, 1, nil
	default:
		return nil, 0, fmt.Errorf("unknown response status code: %d", response.StatusCode)
	}
}
