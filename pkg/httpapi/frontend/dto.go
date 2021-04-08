package frontend

type AllowDTO struct {
	FQDN string `json:"fqdn"`
}

type ResponseDTO struct {
	Message string `json:"msg"`
}
