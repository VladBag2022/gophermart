package accrual

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type orderInfoResponse struct {
	Order     int     `json:"order"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual"`
}

func (d Daemon) orderInfo(order int64) (info *orderInfoResponse, retryAfter int, err error) {
	response, err := http.Get(fmt.Sprintf("%s/api/orders/%d", d.accrualAddress, order))
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, 0, err
		}
		if err = json.Unmarshal(content, &info); err != nil {
			return nil, 0, err
		}
		return info, 0, nil
	} else if response.StatusCode == 429 {
		header := response.Header.Get("Retry-After")
		retryAfter, err = strconv.Atoi(header)
		if err != nil {
			return nil, 0, err
		}
		return nil, retryAfter, nil
	} else if response.StatusCode == 204 {
		return nil, 1, nil
	}
	return nil, 0, fmt.Errorf("unknown response status code: %d", response.StatusCode)
}
