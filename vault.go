package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func Initialize() VaultToken {
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
	defer res.Body.Close()

	if res.StatusCode != 200 {
		panic("init: non 200 status code: " + strconv.Itoa(res.StatusCode))
	}

	vaultResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}

	target := VaultToken{}
	fromJSON(vaultResponse, &target)
	return target
}

func Unseal() {
	exists, err := IsSecretExists()
	if exists == false {
		log.Print("Secrets do not exist!")
		panic(err)
	}

	secrets := GetSecret()
	token1, error := base64.StdEncoding.DecodeString(secrets.Data.Token1)
	if error != nil {
		log.Print("Could not decode token1")
		panic(error)
	}
	token2, error := base64.StdEncoding.DecodeString(secrets.Data.Token2)
	if error != nil {
		log.Print("Could not decode token2")
		panic(error)
	}
	token3, error := base64.StdEncoding.DecodeString(secrets.Data.Token3)
	if error != nil {
		log.Print("Could not decode token3")
		panic(error)
	}

	UseKey(string(token1[:]))
	UseKey(string(token2[:]))
	UseKey(string(token3[:]))
}

func UseKey(key string) {
	unsealToken := UnsealToken{
		UnsealKey: key,
	}

	b := toJSON(unsealToken)

	req, err := http.NewRequest("POST", GetVaultUrl("/v1/sys/unseal"), &b)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("Status Code: %d", res.StatusCode)
		log.Printf("Body: %+v", res.Body)

		panic("init: non 200 status code: " + strconv.Itoa(res.StatusCode))
	}

	vaultResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}

	target := VaultResponse{}
	fromJSON(vaultResponse, &target)
}

func Verify() error {
	status := GetStatus()

	switch status {
	case 200:
		log.Println("Vault is initialized and unsealed.")
	case 429:
		log.Println("Vault is unsealed and in standby mode.")
	case 501:
		log.Println("Vault is not initialized. Initializing and unsealing...")
		vaultResponse := Initialize()
		SaveTokens(vaultResponse)
		Unseal()
	case 503:
		log.Println("Vault is sealed. Unsealing...")
		Unseal()
	default:
		log.Printf("Vault is in an unknown state. Status code: %d", status)
	}

	return nil
}

func GetStatus() int {
	// res, err := httpClient.Head(GetVaultUrl("/v1/sys/health"))
	res, err := httpClient.Head(GetVaultUrl("/v1/sys/health"))
	if err != nil {
		fmt.Print(err)
		log.Printf("Sleeping 10 seconds")
		time.Sleep(10 * time.Second)
		GetStatus()
	}

	return res.StatusCode
}

func GetVaultUrl(url string) string {
	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		return "http://127.0.0.1:8200" + url
		// return "http://vault-dev.exzeo.io:8200" + url
	}

	return vaultAddr + url
}
