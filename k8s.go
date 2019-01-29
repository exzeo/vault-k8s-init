package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	vaultSecretName = "vault-tokens"
)

func GetSecret() string {
	req, err := http.NewRequest("GET", GetSecretUrl()+"/"+vaultSecretName, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", "application/json")

	token := GetBearerToken()
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+string(token))
	}

	// Response Received
	res, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}

	if res.StatusCode != 200 {
		return ""
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	return buf.String()
}

func IsSecretExists() (bool, string) {
	token := GetSecret()
	if token == "" {
		return false, ""
	}
	return true, "Secret Exists!"
}

func CreateSecret(vault VaultToken) {

	secret := Secret{
		Kind:       "Secret",
		APIVersion: "v1",
		Metadata: MetaData{
			Name: vaultSecretName,
		},
		Data: vault,
	}

	b := toJSON(secret)

	req, err := http.NewRequest("POST", GetSecretUrl(), &b)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", "application/json")

	token := GetBearerToken()
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+string(token))
	}

	// Response Received
	res, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	fmt.Println(res)
}

func GetBearerToken() string {

	location := "/var/run/secrets/kubernetes.io/serviceaccount/token"
	_, er := os.Stat(location)
	if er == nil {
		token, err := ioutil.ReadFile(location)
		if err != nil {
			panic(err)
		}

		return string(token)
	}

	return ""
}

func GetSecretUrl() string {

	var url string

	namespace := os.Getenv("KUBERNETES_NAMESPACE")

	if namespace == "" {
		namespace = "default"
	}

	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_PORT_443_TCP_PORT")
	if host != "" && port != "" {
		url = "https://" + host + ":" + port

	} else {
		url = "http://localhost:8001"
	}

	return url + "/api/v1/namespaces/" + namespace + "/secrets"
}
