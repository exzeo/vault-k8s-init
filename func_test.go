package main

import (
	"testing"
)

func TestCreateSecret(t *testing.T) {
	request := Initialize()
	SaveTokens(request)
	Unseal()
}
