package iface

type Allowed interface {
	AllowedA(name string) (bool, error)
	AllowedAAAA(name string) (bool, error)
	AllowedPTR(name string) (bool, error)
}

type AllowAPI interface {
	AllowA(name string) error    // IPv4
	AllowAAAA(name string) error // IPv6
	AllowPTR(name string) error  // Reverse
}

type Database interface {
	Allowed
	AllowAPI
}
