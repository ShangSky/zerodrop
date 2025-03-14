package model

import (
	"encoding/json"
	"net/http"
)

const (
	MethodFreshUsers         = "FRESH_USERS"
	MethodSendMsg            = "SEND_MSG"
	MethodReceiveMsg         = "RECEIVE_MSG"
	MethodSendFileStart      = "SEND_FILE_START"
	MethodSendFileCancel     = "SEND_FILE_CANCEL"
	MethodReceiveFileConfirm = "RECEIVE_FILE_CONFIRM"
	MethodReceiveFileREFUSED = "RECEIVE_FILE_REFUSED"
)

type Req struct {
	Method string          `json:"method"`
	Data   json.RawMessage `json:"data"`
}

type Resp struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Method string `json:"method"`
	Data   any    `json:"data"`
}

func RespOK(Method string, data any) Resp {
	return Resp{Status: http.StatusOK, Method: Method, Msg: "success", Data: data}
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Device   string `json:"device"`
	IsMobile bool   `json:"is_mobile"`
}

type RegisterResp struct {
	Me    User   `json:"me"`
	Users []User `json:"users"`
}

type SendMsg struct {
	To  string `json:"to"`
	Msg string `json:"msg"`
}

type ReceiveMsg struct {
	From string `json:"from"`
	Msg  string `json:"msg"`
}

type ReceiveFileConfirm struct {
	From       string `json:"from"`
	DownloadID string `json:"download_id"`
	Filename   string `json:"filename"`
	FileSize   int64  `json:"file_size"`
}

type SendFileStartResp struct {
	DownloadID string `json:"download_id"`
}

type TransferFileCancel struct {
	DownloadID string `json:"download_id"`
}
