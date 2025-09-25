package models

type Response[T any] struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Request processed successfully"`
	Data    T      `json:"data"`
}

type ResponseLogin struct {
	Success    bool   `json:"success" example:"true"`
	Message    string `json:"message" example:"Request processed successfully"`
	Token      string `json:"token"`
	IsPinExist bool   `json:"isPinExist" example:"false"`
}
