package p2p

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"

	"fx/chain/common/crypto"
	"fx/chain/logger"
)

const Version = "1.0.5"

type EventCallback interface {
	OnMessage(res *Result)
}

type Result struct {
	success     bool
	code        int
	msg         string
	hash        string
	blockNumber uint64
	data        string
}

func (res *Result) Success() bool    { return res.success }
func (res *Result) Code() int        { return res.code }
func (res *Result) Msg() string      { return res.msg }
func (res *Result) GetData() string  { return res.data }
func (res *Result) GetHash() string  { return res.hash }
func (res *Result) BlockNumber() int { return int(res.blockNumber) }

func (res *Result) String() string {
	return fmt.Sprintf("success:%v, msg:%s, data:%s, hash:%s", res.success, res.msg, res.data, res.hash)
}

func failed(msg string, code ...int) *Result {
	r := &Result{success: false, msg: msg}
	if len(code) > 0 {
		r.code = code[0]
	}
	return r
}

func success(msg string, code ...int) *Result {
	r := &Result{success: true, msg: msg}
	if len(code) > 0 {
		r.code = code[0]
	}
	return r
}

func successData(msg, data, hash string, blockNumber uint64, code ...int) *Result {
	r := &Result{success: true, msg: msg, data: data, hash: hash, blockNumber: blockNumber}
	if len(code) > 0 {
		r.code = code[0]
	}
	return r
}

type Location struct {
	Longitude string
	Latitude  string
}

func GetLocationByIP(ip string) *Location {
	lat, lon, err := getLocationByHttp(ip)
	if err != nil {
		logger.Error("location http", "err", err.Error())
		return &Location{}
	}
	return &Location{floatToString(lon), floatToString(lat)}
}

func getLocationByHttp(ip string) (Lat, Lon float64, err error) {
	resp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s?fields=520191&lang=en", ip))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var location struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}
	err = json.NewDecoder(resp.Body).Decode(&location)
	return location.Lat, location.Lon, err
}

func floatToString(number float64) string {
	return fmt.Sprintf("%.4f", number)
}

func parsePriKey(priKey string) *ecdsa.PrivateKey {
	key, err := crypto.PriKeyDecode(priKey)
	if err != nil {
		panic("parse private key err:" + err.Error())
	}
	logger.Debug("============", "pub key", crypto.PubKeyEncode(&key.PublicKey))
	logger.Debug("============", "address", crypto.Address(key.PublicKey))
	return key
}

