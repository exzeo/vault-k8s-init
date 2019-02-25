package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	log.Println("Starting the vault-init service...")

	checkInterval := os.Getenv("CHECK_INTERVAL")
	if checkInterval == "" {
		checkInterval = "10"
	}

	i, err := strconv.Atoi(checkInterval)
	if err != nil {
		log.Fatalf("CHECK_INTERVAL is invalid: %s", err)
	}

	checkIntervalDuration := time.Duration(i) * time.Second

	//Allow CTRL+C
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

		response, err := httpClient.Head(GetVaultURL("/v1/sys/health"))

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
			vaultResponse := Initialize()
			log.Print("Initialized!! Saving Tokens")
			SaveTokens(vaultResponse)
			Unseal()
		case 503:
			log.Println("Vault is sealed. Unsealing...")
			Unseal()
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