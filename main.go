// Copyright 2018 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package main

import (
	"bytes"
	// "context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

var (
	vaultAddr     string
	k8sHost       string
	k8sPort       string
	k8sAddr       string
	k8sToken      string
	k8sNamespace  string
	checkInterval string
	gcsBucketName string
	httpClient    http.Client
	unencoded     string

	userAgent = fmt.Sprintf("vault-init/0.1.0 (%s)", runtime.Version())
)

// InitRequest holds a Vault init request.
type InitRequest struct {
	SecretShares    int `json:"secret_shares"`
	SecretThreshold int `json:"secret_threshold"`
}

// type secretMetadata struct {
// 	name string
// }

// type secretData struct {
// 	testSecret  string
// 	testSecret2 string
// }

// type k8sSecretsRequest struct {
// 	kind       string
// 	apiVersion string
// 	metadata   secretMetadata
// 	data       secretData
// }

// InitResponse holds a Vault init response.
type InitResponse struct {
	Keys       []string `json:"keys"`
	KeysBase64 []string `json:"keys_base64"`
	RootToken  string   `json:"root_token"`
}

// UnsealRequest holds a Vault unseal request.
type UnsealRequest struct {
	Key   string `json:"key"`
	Reset bool   `json:"reset"`
}

// UnsealResponse holds a Vault unseal response.
type UnsealResponse struct {
	Sealed   bool `json:"sealed"`
	T        int  `json:"t"`
	N        int  `json:"n"`
	Progress int  `json:"progress"`
}

func main() {
	log.Println("Starting the vault-init service...")

	// Set kubernetes token as ENV variable
	kubeToken, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	os.Setenv("KUBE_TOKEN", string(kubeToken))

	vaultAddr = os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		// vaultAddr = "https://127.0.0.1:8200"
		vaultAddr = "http://vault-dev.exzeo.io:8200"
	}

	k8sHost = os.Getenv("KUBERNETES_SERVICE_HOST")
	k8sPort = os.Getenv("KUBERNETES_PORT_443_TCP_PORT")
	if k8sAddr == "" {
		k8sAddr = "https://" + k8sHost + ":" + k8sPort
	}

	checkInterval = os.Getenv("CHECK_INTERVAL")
	if checkInterval == "" {
		checkInterval = "10"
	}

	i, err := strconv.Atoi(checkInterval)
	if err != nil {
		log.Fatalf("CHECK_INTERVAL is invalid: %s", err)
	}

	checkIntervalDuration := time.Duration(i) * time.Second

	// gcsBucketName = os.Getenv("GCS_BUCKET_NAME")
	// if gcsBucketName == "" {
	// 	log.Fatal("GCS_BUCKET_NAME must be set and not empty")
	// }

	// kmsKeyId = os.Getenv("KMS_KEY_ID")
	// if kmsKeyId == "" {
	// 	log.Fatal("KMS_KEY_ID must be set and not empty")
	// }

	// kmsCtx, kmsCtxCancel := context.WithCancel(context.Background())
	// defer kmsCtxCancel()
	// kmsClient, err := google.DefaultClient(kmsCtx, "https://www.googleapis.com/auth/cloudkms")
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// kmsService, err = cloudkms.New(kmsClient)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// kmsService.UserAgent = userAgent

	// storageCtx, storageCtxCancel := context.WithCancel(context.Background())
	// defer storageCtxCancel()
	// storageClient, err = storage.NewClient(storageCtx,
	// 	option.WithUserAgent(userAgent),
	// 	option.WithScopes(storage.ScopeReadWrite),
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }

	httpClient = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Allow CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
	)

	stop := func() {
		log.Printf("Shutting down")
		// kmsCtxCancel()
		// storageCtxCancel()
		os.Exit(0)
	}

	for {
		select {
		case <-signalCh:
			stop()
		case <-c:
			stop()
		default:
		}
		response, err := httpClient.Head(vaultAddr + "/v1/sys/health")

		if response != nil && response.Body != nil {
			response.Body.Close()
		}

		if err != nil {
			log.Println(err)
			time.Sleep(checkIntervalDuration)
			continue
		}

		switch response.StatusCode {
		case 200:
			log.Println("Vault is initialized and unsealed.")
		case 429:
			log.Println("Vault is unsealed and in standby mode.")
		case 501:
			log.Println("Vault is not initialized. Initializing and unsealing...")
			initialize()
			unseal()
		case 503:
			log.Println("Vault is sealed. Unsealing...")
			unseal()
		default:
			log.Printf("Vault is in an unknown state. Status code: %d", response.StatusCode)
		}

		log.Printf("Next check in %s", checkIntervalDuration)

		select {
		case <-signalCh:
			stop()
		case <-time.After(checkIntervalDuration):
		}
	}
}

func initialize() {
	initRequest := InitRequest{
		SecretShares:    5,
		SecretThreshold: 3,
	}

	initRequestData, err := json.Marshal(&initRequest)
	if err != nil {
		log.Println(err)
		return
	}

	r := bytes.NewReader(initRequestData)
	request, err := http.NewRequest("PUT", vaultAddr+"/v1/sys/init", r)
	if err != nil {
		log.Println(err)
		return
	}

	response, err := httpClient.Do(request)
	if err != nil {
		log.Println(err)
		return
	}
	defer response.Body.Close()

	initRequestResponseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return
	}

	if response.StatusCode != 200 {
		log.Printf("init: non 200 status code: %d", response.StatusCode)
		return
	}

	var initResponse InitResponse

	if err := json.Unmarshal(initRequestResponseBody, &initResponse); err != nil {
		log.Println(err)
		return
	}

	log.Println("Encrypting unseal keys and the root token...")

	setSecrets(initResponse)

	// rootTokenEncryptRequest := &cloudkms.EncryptRequest{
	// 	Plaintext: base64.StdEncoding.EncodeToString([]byte(initResponse.RootToken)),
	// }

	// rootTokenEncryptResponse, err := kmsService.Projects.Locations.KeyRings.CryptoKeys.Encrypt(kmsKeyId, rootTokenEncryptRequest).Do()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// unsealKeysEncryptRequest := &cloudkms.EncryptRequest{
	// 	Plaintext: base64.StdEncoding.EncodeToString(initRequestResponseBody),
	// }

	// unsealKeysEncryptResponse, err := kmsService.Projects.Locations.KeyRings.CryptoKeys.Encrypt(kmsKeyId, unsealKeysEncryptRequest).Do()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// bucket := storageClient.Bucket(gcsBucketName)

	// Save the encrypted unseal keys.
	// ctx := context.Background()
	// unsealKeysObject := bucket.Object("unseal-keys.json.enc").NewWriter(ctx)
	// defer unsealKeysObject.Close()

	// _, err = unsealKeysObject.Write([]byte(unsealKeysEncryptResponse.Ciphertext))
	// if err != nil {
	// 	log.Println(err)
	// }

	// log.Printf("Unseal keys written to gs://%s/%s", gcsBucketName, "unseal-keys.json.enc")

	// // Save the encrypted root token.
	// rootTokenObject := bucket.Object("root-token.enc").NewWriter(ctx)
	// defer rootTokenObject.Close()

	// _, err = rootTokenObject.Write([]byte(rootTokenEncryptResponse.Ciphertext))
	// if err != nil {
	// 	log.Println(err)
	// }

	// log.Printf("Root token written to gs://%s/%s", gcsBucketName, "root-token.enc")

	// log.Println("Initialization complete.")
}

func unseal() {
	// bucket := storageClient.Bucket(gcsBucketName)

	// need to read keys in from kubernetes
	// ctx := context.Background()
	// unsealKeysObject, err := bucket.Object("unseal-keys.json.enc").NewReader(ctx)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// defer unsealKeysObject.Close()

	// unsealKeysData, err := ioutil.ReadAll(unsealKeysObject)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// unsealKeysDecryptRequest := &cloudkms.DecryptRequest{
	// 	Ciphertext: string(unsealKeysData),
	// }

	// unsealKeysDecryptResponse, err := kmsService.Projects.Locations.KeyRings.CryptoKeys.Decrypt(kmsKeyId, unsealKeysDecryptRequest).Do()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// var initResponse InitResponse

	// unsealKeysPlaintext, err := base64.StdEncoding.DecodeString(unsealKeysDecryptResponse.Plaintext)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// if err := json.Unmarshal(unsealKeysPlaintext, &initResponse); err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// for _, key := range initResponse.KeysBase64 {
	// 	done, err := unsealOne(key)
	// 	if done {
	// 		return
	// 	}

	// 	if err != nil {
	// 		log.Println(err)
	// 		return
	// 	}
	// }
}

func unsealOne(key string) (bool, error) {
	// unsealRequest := UnsealRequest{
	// 	Key: key,
	// }

	// unsealRequestData, err := json.Marshal(&unsealRequest)
	// if err != nil {
	// 	return false, err
	// }

	// r := bytes.NewReader(unsealRequestData)
	// request, err := http.NewRequest(http.MethodPut, vaultAddr+"/v1/sys/unseal", r)
	// if err != nil {
	// 	return false, err
	// }

	// response, err := httpClient.Do(request)
	// if err != nil {
	// 	return false, err
	// }
	// defer response.Body.Close()

	// if response.StatusCode != 200 {
	// 	return false, fmt.Errorf("unseal: non-200 status code: %d", response.StatusCode)
	// }

	// unsealRequestResponseBody, err := ioutil.ReadAll(response.Body)
	// if err != nil {
	// 	return false, err
	// }

	// var unsealResponse UnsealResponse
	// if err := json.Unmarshal(unsealRequestResponseBody, &unsealResponse); err != nil {
	// 	return false, err
	// }

	// if !unsealResponse.Sealed {
	// 	return true, nil
	// }

	return false, nil
}

func toBase64(key string) string {
	return base64.StdEncoding.EncodeToString([]byte(key))
}

func setSecrets(initResponse InitResponse) *http.Response {

	rootTokenEncrypted := toBase64(initResponse.RootToken)
	log.Println("rootTokenEncrypted\n\n" + rootTokenEncrypted + "\n\n")

	k8sSecretsRequest := map[string]interface{}{
		"kind":       "Secret",
		"apiVersion": "v1",
		"metadata": map[string]string{
			"name": "Vault Tokens",
		},
		"data": map[string]interface{}{
			"vault_root_token":  toBase64(initResponse.RootToken),
			"vault_token1": toBase64(initResponse.Keys[0]),
			"vault_token2": toBase64(initResponse.Keys[1]),
			"vault_token3": toBase64(initResponse.Keys[2]),
			"vault_token4": toBase64(initResponse.Keys[3]),
			"vault_token5": toBase64(initResponse.Keys[4]),
		},
	}

	// parse JSON Data
	k8sSecretRequestData, err := json.Marshal(&k8sSecretsRequest)
	if err != nil {
		log.Println(err)
		// return
	} else {
		log.Println("k8sSecretRequestData PASSED")
		log.Println(k8sSecretRequestData)
	}

	// POST to k8sAddr+`/api/v1/namespaces/vault-dev/secrets
	k8sNamespace = os.Getenv("KUBERNETES_NAMESPACE")
	k8sToken = os.Getenv("KUBE_TOKEN")
	k8sR := bytes.NewReader(k8sSecretRequestData)
	k8sRequest, err := http.NewRequest("POST", k8sAddr+"/api/v1/namespaces/"+k8sNamespace+"/secrets", k8sR)
	k8sRequest.Header.Add("Accept", "application/json")
	k8sRequest.Header.Add("Content-Type", "application/json")
	k8sRequest.Header.Add("Authorization", "Bearer "+k8sToken)
	if err != nil {
		log.Println(err)
		// return
	} else {
		log.Println("k8sRequest PASSED")
		log.Print(k8sRequest)
	}

	// Response Received
	k8sResponse, err := httpClient.Do(k8sRequest)
	if err != nil {
		log.Println(err)
		// return
	} else {
		log.Println("k8sResponse PASSED")
		log.Println(k8sResponse)
	}
	defer k8sResponse.Body.Close()

	return k8sResponse
}
