package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

// import (
// 	"log"
// 	"testing"
// )

// TestSetSecrets : allows you to test result of setSecrets
// func TestSetSecrets(t *testing.T) {
// 	// expected := fmt.Sprintf("%s.us-east-1.s3.exzeo.com", bucket)
// 	result := setSecrets()

//     log.Println(result)

// 	// if result == "" {
// 	//     t.Fail()
// 	// }

// 	// if result != expected {
// 	//     t.Errorf("Got %s, Expected: %s", result, expected)
// 	// }
// }

func TestStruct(t *testing.T) {
	secret := Secret{
		Kind:       "Secret",
		APIVersion: "v1",
		Metadata: MetaData{
			Name: "Vault Tokens",
		},
		Data: VaultToken{
			Vault_root_token: toBase64("test root")+"=",
			Vault_token1:     toBase64("test token 1")+"=",
			Vault_token2:     toBase64("test token 2")+"=",
			Vault_token3:     toBase64("test token 3")+"=",
			Vault_token4:     toBase64("test token 4")+"=",
			Vault_token5:     toBase64("test token 5")+"=",
		},
	}

	b, err := json.Marshal(secret)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))

}
