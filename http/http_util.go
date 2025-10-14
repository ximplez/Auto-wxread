package http

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

var client = resty.New()

func Get(url string, params, headers map[string]string) (http.Header, []byte, error) {
	resp, err := client.R().
		SetQueryParams(params).
		SetHeaders(headers).
		Get(url)
	if err != nil {
		return nil, nil, err
	}
	return resp.Header(), resp.Body(), nil
}

func Post(url string, data string, headers map[string]string) (http.Header, string, error) {
	resp, err := client.R().
		SetBody(data).
		SetHeaders(headers).
		Post(url)
	if err != nil {
		return nil, "", err
	}
	return resp.Header(), resp.String(), nil
}
