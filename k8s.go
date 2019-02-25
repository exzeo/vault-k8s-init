package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

// GetSecret - retrieves secret from Kubernetes
func GetSecret() Secret {
	req, err := http.NewRequest("GET", GetSecretURL()+"/"+vaultSecretName, nil)
	if err != nil {
		log.Print(err)
	}

	req.Header.Add("Content-Type", "application/json")

	token := GetBearerToken()
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+string(token))
	}

	caCertPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		panic(err) // Can't find cert file
	}
	caCertPool.AppendCertsFromPEM(caCert)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	res, err := client.Do(req)
	if err != nil {
		log.Print(err)
		// panic(err)
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

// IsSecretExists - checks if secret exists already in Kubernetes
func IsSecretExists() (bool, string) {
	log.Print("Checking for tokens")
	token := GetSecret()
	if token.Data.RootToken == "" {
		return false, ""
	}
	return true, "Secret Exists!"
}

// SaveTokens - checks for tokens then formats to be saved
func SaveTokens(tokens VaultToken) {
	exists, err := IsSecretExists()

	if exists == true {
		log.Print(err)
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

// CreateSecret - creates the secret in Kubernetes
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

	req, err := http.NewRequest("POST", GetSecretURL(), &b)
	if err != nil {
		log.Print(err)
		// panic(err)
	}

	token := GetBearerToken()
	if token == "" {
		log.Printf("No kubernetes token exists!!!")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", string(token)))

	caCertPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		panic(err) // Can't find cert file
	}
	caCertPool.AppendCertsFromPEM(caCert)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	res, err := client.Do(req)
	if err != nil {
		log.Print(err)
		// panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		panic("init: non 201 status code: " + strconv.Itoa(res.StatusCode))
	}
}

// GetBearerToken - grabs the token from /var/run/secrets/kubernetes.io/serviceaccount/token - needs correct RBAC permissions
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

// GetSecretURL - formats the URL to access Kubernetes secrets
func GetSecretURL() string {

	var url string

	namespace := os.Getenv("KUBERNETES_NAMESPACE")

	if namespace == "" {
		namespace = "default"
	}

	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")
	if host != "" && port != "" {
		url = "https://" + host + ":" + port

	} else {
		url = "http://localhost:8001"
	}

	return url + "/api/v1/namespaces/" + namespace + "/secrets"
}
