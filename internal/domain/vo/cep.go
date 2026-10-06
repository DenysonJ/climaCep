package cep

import (
	"errors"
	"regexp"
)

type Cep struct {
	Value string `json:"cep"`
}

func New(cep string) (*Cep, error) {
	if !isValid(cep) {
		return nil, errors.New("invalid cep")
	}
	return &Cep{Value: cep}, nil
}

func isValid(cep string) bool {
	if regexp.MustCompile("^\\d{5}-?\\d{3}$").MatchString(cep) {
		return true
	}

	return false
}
