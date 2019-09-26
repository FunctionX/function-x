package client

import (
	"io"
	"net/http"

	"fx/chain/common/utils"
)

var (
	AddHead = "/api/v0/add"
)

type Response struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
	Size string `json:"size"`
}

type Client struct {
	client   *http.Client
	net      string
	method   string
	hostname string
	ip       string
	port     uint16
	path     string
	body     io.Reader
	header   map[string]string
}

func NewClient() (*Client, error) {
	client := &Client{client: http.DefaultClient}
	client.net = "http"
	client.ip = "127.0.0.1"
	return client, nil
}

func (c *Client) AddFile(ip string, path string) (string, error) {
	c.ip = ip
	c.port = 5001
	c.path = AddHead
	c.method = "POST"
	res := Response{}
	contentType, body, err := utils.MakeReaderFromPath(path, "ipfs_add_file", make(map[string]string))
	header := make(map[string]string)
	header["Content-Type"] = contentType

	c.body = body
	c.header = header

	err = c.do(&res)
	if err != nil {
		return "", err
	}
	return res.Hash, nil
}
