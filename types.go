package main

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

// InitRequest holds a Vault init request.
type InitRequest struct {
	SecretShares    int `json:"secret_shares"`
	SecretThreshold int `json:"secret_threshold"`
}

type Secret struct {
	Kind       string     `json:"kind"`
	APIVersion string     `json:"apiVersion"`
	Metadata   MetaData   `json:"metadata"`
	Data       VaultToken `json:"data"`
}

type MetaData struct {
	Name string `json:"name"`
}

type VaultToken struct {
	Vault_root_token string `json:"root_token"`
	Vault_token1     string `json:"token1"`
	Vault_token2     string `json:"token2"`
	Vault_token3     string `json:"token3"`
	Vault_token4     string `json:"token4"`
	Vault_token5     string `json:"token5"`
}
