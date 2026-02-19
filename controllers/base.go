// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package controllers

import (
	"encoding/json"

	"github.com/beego/beego/v2/server/web"
)

type BaseController struct {
	web.Controller
}

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data,omitempty"`
	Data2  interface{} `json:"data2,omitempty"`
}

func (c *BaseController) ResponseOk(data ...interface{}) {
	resp := Response{Status: "ok"}
	if len(data) > 0 {
		resp.Data = data[0]
	}
	if len(data) > 1 {
		resp.Data2 = data[1]
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *BaseController) ResponseError(msg string, data ...interface{}) {
	resp := Response{Status: "error", Msg: msg}
	if len(data) > 0 {
		resp.Data = data[0]
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *BaseController) GetRequestBody(v interface{}) error {
	return json.Unmarshal(c.Ctx.Input.RequestBody, v)
}
