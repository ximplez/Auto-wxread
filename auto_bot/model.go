package main

type ErrorResp struct {
	Errcode int    `json:"errcode"`
	Errlog  string `json:"errlog"`
	Errmsg  string `json:"errmsg"`
}
