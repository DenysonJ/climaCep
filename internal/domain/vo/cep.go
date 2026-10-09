package cep

import (
	"errors"
	"regexp"
)

var ErrorInvalidCep = errors.New("invalid zipcode")

type Cep struct {
	Value string `json:"cep"`
}

func New(cep string) (*Cep, error) {
	if !isValid(cep) {
		return nil, ErrorInvalidCep
	}
	return &Cep{Value: cep}, nil
}

func isValid(cep string) bool {
	if regexp.MustCompile("^\\d{5}-?\\d{3}$").MatchString(cep) {
		return true
	}

	return false
}
