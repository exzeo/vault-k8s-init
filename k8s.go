package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

func GetSecret() Secret {
	req, err := http.NewRequest("GET", GetSecretUrl()+"/"+vaultSecretName, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", "application/json")

	token := GetBearerToken()
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+string(token))
	}

	res, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	k8sResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}

	target := Secret{}
	fromJSON(k8sResponse, &target)
	return target
}

func IsSecretExists() (bool, string) {
	token := GetSecret()
	if token.Data.RootToken == "" {
		return false, ""
	}
	return true, "Secret Exists!"
}

func SaveTokens(tokens VaultToken) {
	exists, err := IsSecretExists()
	if exists == true {
		panic(err)
	}

	secret := K8sSecrets{
		RootToken: base64.StdEncoding.EncodeToString([]byte(tokens.RootToken)),
		Token1:    base64.StdEncoding.EncodeToString([]byte(tokens.Tokens[0])),
		Token2:    base64.StdEncoding.EncodeToString([]byte(tokens.Tokens[1])),
		Token3:    base64.StdEncoding.EncodeToString([]byte(tokens.Tokens[2])),
		Token4:    base64.StdEncoding.EncodeToString([]byte(tokens.Tokens[3])),
		Token5:    base64.StdEncoding.EncodeToString([]byte(tokens.Tokens[4])),
	}

	CreateSecret(secret)
}

func CreateSecret(vault K8sSecrets) {
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

	res, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		panic("init: non 201 status code: " + strconv.Itoa(res.StatusCode))
	}
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
