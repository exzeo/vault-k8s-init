package main

import (
	"testing"
)

// func TestCreateSecret(t *testing.T) {
// 	// secret := Secret{
// 	// 	Kind:       "Secret",
// 	// 	APIVersion: "v1",
// 	// 	Metadata: MetaData{
// 	// 		Name: "Vault Tokens",
// 	// 	},
// 	// 	Data: VaultToken{
// 	// 		Vault_root_token: toBase64("test root"),
// 	// 		Vault_token1:     toBase64("test token 1"),
// 	// 		Vault_token2:     toBase64("test token 2"),
// 	// 		Vault_token3:     toBase64("test token 3"),
// 	// 		Vault_token4:     toBase64("test token 4"),
// 	// 		Vault_token5:     toBase64("test token 5"),
// 	// 	},
// 	// }

// 	// b, err := json.Marshal(secret)
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// 	return
// 	// }
// 	// fmt.Println(string(b))

// 	tokens := []string{}

// 	for index := 0; index < NumTokens; index++ {
// 		tokens = append(tokens, "token")
// 	}

// 	vault := VaultToken{
// 		RootToken: "test-root",
// 		Tokens:    tokens,
// 	}

// 	k8s.CreateSecret(value)
// }
