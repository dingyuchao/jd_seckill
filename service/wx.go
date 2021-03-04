package service

import (
	"encoding/json"
	"log"
	"time"

	"github.com/giant-stone/go/ghttp"
	"github.com/giant-stone/go/gutil"
)

func NotifyUser(token, msg string) {
	if token == "" {
		return
	}

	fullurl := "https://sre24.com/api/v1/push"
	rqBody, _ := json.Marshal(&map[string]interface{}{
		"token": token,
		"msg": msg,
	})
	rq := ghttp.New().
		SetTimeout(time.Second * 3).
		SetRequestMethod("POST").
		SetUri(fullurl).
		SetPostBody(&rqBody)
	err := rq.Send()
	if gutil.CheckErr(err) {
		log.Println("notify user fail", string(rq.RespBody))
	}
}
