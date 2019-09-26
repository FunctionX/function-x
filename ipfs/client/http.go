package client

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func (c *Client) do(obj interface{}) error {
	resp, err := c.request()
	if err != nil {
		return err
	}
	return c.handlerResponse(resp, obj)
}

func (c *Client) request() (*http.Response, error) {
	port, err := c.checkPort()
	if err != nil {
		return nil, err
	}
	urlPath := c.net + "://" + c.ip + ":" + port + "/" + strings.TrimPrefix(c.path, "/")
	r, err := http.NewRequest(c.method, urlPath, c.body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	for k, v := range c.header {
		r.Header.Set(k, v)
	}
	return c.client.Do(r)
}

func (c *Client) checkPort() (string, error) {
	if c.port == 0 || c.port > 65535 {
		return "", errors.New("http port error")
	}
	port := strconv.FormatUint(uint64(c.port), 10)
	return port, nil
}

func (c *Client) handlerResponse(resp *http.Response, obj interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	switch {
	case resp.StatusCode == http.StatusAccepted:
		//log.Printf("request accepted")
	case resp.StatusCode == http.StatusNoContent:
		//log.Printf("request suceeded,but no content")
	default:
		if resp.StatusCode > 399 {
			err = json.Unmarshal(body, err)
			if err != nil {
				return err
			}
			return err
			//return errors.New("error:" + strconv.Itoa(resp.StatusCode))
		}
		err = json.Unmarshal(body, obj)
		if err != nil {
			return err
		}
	}
	return nil
}
