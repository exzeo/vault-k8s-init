package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func Initialize() VaultToken{

	initRequest := InitRequest{
		SecretShares:    NumTokens,
		SecretThreshold: TokensRequired,
	}

	r := toJSON(initRequest)
	req, err := http.NewRequest("PUT", GetVaultUrl("/v1/sys/init"), &r)
	if err != nil {
		panic(err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}

	target := VaultToken{}

	json.NewDecoder(res.Body).Decode(target)

	return target
}

func SaveTokens(tokens VaultToken) {
	exists,err := IsSecretExists()
	if exists == true {
		panic(err)
	}
	
	log.Print(tokens)
	CreateSecret(tokens)
}

func Unseal() {

}

func Verify() error {
	status, err := GetStatus()
	if err != nil {
		return err
	}

	switch status {
	case 200:
		log.Println("Vault is initialized and unsealed.")
	case 429:
		log.Println("Vault is unsealed and in standby mode.")
	case 501:
		log.Println("Vault is not initialized. Initializing and unsealing...")
		request := Initialize()
		SaveTokens(request)
		// Unseal()
	case 503:
		log.Println("Vault is sealed. Unsealing...")
		Unseal()
	default:
		log.Printf("Vault is in an unknown state. Status code: %d", status)
	}

	return nil
}

func GetStatus() (int, error) {
	res, err := httpClient.Head(GetVaultUrl("/v1/sys/health"))

	return res.StatusCode, err
}

func GetVaultUrl(url string) string {
	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		// return "https://127.0.0.1:8200"
		return "http://vault-dev.exzeo.io:8200" + url
	}

	return vaultAddr + url
}
