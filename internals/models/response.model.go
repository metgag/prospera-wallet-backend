package models

type Response struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Request processed successfully"`
	Data    interface{} `json:"data"`
}

type ResponseLogin struct {
	Success    bool   `json:"success" example:"true"`
	Message    string `json:"message" example:"Login successful"`
	Token      string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	IsPinExist bool   `json:"isPinExist" example:"false"`
	Email      string `json:"email" example:"user1@mail.com"`
}

type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	TotalPages int         `json:"total_pages"`
}
