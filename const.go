package main

import "net/http"

var (
	// NumTokens is the number of tokens created during vault init
	NumTokens = 5

	// TokensRequired is how many tokens required to unseal
	TokensRequired = 3

	// vaultSecretName is name of secret in Kubernetes
	vaultSecretName = "vault-tokens"

	httpClient http.Client
)
