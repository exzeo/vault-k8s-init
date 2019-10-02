package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

// Initialize - initialize vault
func Initialize() VaultToken {
	initRequest := InitRequest{
		SecretShares:    NumTokens,
		SecretThreshold: TokensRequired,
	}

	r := toJSON(initRequest)
	req, err := http.NewRequest("PUT", GetVaultURL("/v1/sys/init"), &r)
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

// Unseal - unseal vault
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

// UseKey - uses a key to unseal vault
func UseKey(key string) {
	unsealToken := UnsealToken{
		UnsealKey: key,
	}

	b := toJSON(unsealToken)

	req, err := http.NewRequest("PUT", GetVaultURL("/v1/sys/unseal"), &b)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	vaultResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}

	if res.StatusCode != 200 {
		log.Printf("Status Code: %d", res.StatusCode)
		log.Printf("Body: %+v", vaultResponse)

		panic("init: non 200 status code: " + strconv.Itoa(res.StatusCode))
	}

	target := VaultResponse{}
	fromJSON(vaultResponse, &target)
}

// GetVaultURL - crafts url for vault
func GetVaultURL(url string) string {
	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		return "http://127.0.0.1:8200" + url
	}
	return vaultAddr + url
}
