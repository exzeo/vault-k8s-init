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

// Secret holds a kubernetes secret
type Secret struct {
	Kind       string     `json:"kind"`
	APIVersion string     `json:"apiVersion"`
	Metadata   MetaData   `json:"metadata"`
	Data       VaultToken `json:"data"`
}

// MetaData holds name for secret.
type MetaData struct {
	Name string `json:"name"`
}

// VaultToken holds root token and tokens to be added to secret.
type VaultToken struct {
	RootToken string   `json:"root_token"`
	Tokens    []string `json:"tokens"`
}
