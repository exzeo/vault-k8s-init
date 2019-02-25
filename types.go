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
	Data       K8sSecrets `json:"data"`
}

// MetaData holds name for secret.
type MetaData struct {
	Name string `json:"name"`
}

// VaultToken holds root token and tokens to be added to secret.
type VaultToken struct {
	RootToken string   `json:"root_token"`
	Tokens []string `json:"keys"`
}

// K8sSecrets holds root token and tokens to be added to secret.
type K8sSecrets struct {
	RootToken string  `json:"root-token"`
	Token1    string  `json:"key1"`
	Token2    string  `json:"key2"`
	Token3    string  `json:"key3"`
	Token4    string  `json:"key4"`
	Token5    string  `json:"key5"`
}

// UnsealToken holds one token used to unseal vault.
type UnsealToken struct {
	UnsealKey string   `json:"key"`
}

// VaultResponse holds staus of vault.
type VaultResponse struct {
	Sealed bool   `json:"sealed"`
	Progress int   `json:"progress"`
}