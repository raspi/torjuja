package iface

type Allowed interface {
	AllowedA(name string) (bool, error)
	AllowedAAAA(name string) (bool, error)
	AllowedPTR(name string) (bool, error)
}

type AllowAPI interface {
	AllowA(name string) error
	AllowAAAA(name string) error
	AllowPTR(name string) error
}

type Database interface {
	Allowed
	AllowAPI
}
