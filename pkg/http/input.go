package http

type AppUIDInput struct {
	AppUID string `form:"appUid" binding:"required"`
}

type NonceInput struct {
	Nonce *string `form:"nonce" header:"X-Request-Nonce" binding:"omitempty,min=32,max=48"`
}

type TokenInput struct {
	Token string `form:"token" binding:"required"`
}
