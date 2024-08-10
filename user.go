package main

import (
	"fmt"
	"net/http"
)

type signUpReq struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ConfirmedPassword string `json:"confirmed_password"`
}

type commonResponse struct {
	Code    int
	Message string
	Data    interface{}
}

func SignUp(ctx *Context) {
	req := &signUpReq{}
	resp := &commonResponse{Code: 0, Message: "suc"}
	err := ctx.ReadJson(req)
	if err != nil {
		ctx.BadRequestJson(resp)
		return
	}

	err = ctx.WriteJson(http.StatusOK, resp)
	if err != nil {
		fmt.Printf("写入响应失败:%v", err)
	}
}
