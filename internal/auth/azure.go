package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Auth represents authentication credentials for Azure DevOps.
type Auth struct {
	Scheme string // "Bearer" or "Basic"
	Token  string
}

// AuthorizationHeader returns the formatted Authorization header value.
func (a *Auth) AuthorizationHeader() string {
	if a.Scheme == "Basic" {
		encoded := base64.StdEncoding.EncodeToString([]byte(":" + a.Token))
		return "Basic " + encoded
	}
	return "Bearer " + a.Token
}

// GetAzureAuth retrieves Azure DevOps authentication.
// It first checks for PAT tokens in environment variables (AZDO_PAT, ADO_PAT),
// then falls back to Azure CLI if available.
func GetAzureAuth() (*Auth, error) {
	// Check for PAT in environment
	if pat := os.Getenv("AZDO_PAT"); pat != "" {
		return &Auth{Scheme: "Basic", Token: pat}, nil
	}
	if pat := os.Getenv("ADO_PAT"); pat != "" {
		return &Auth{Scheme: "Basic", Token: pat}, nil
	}

	// Try Azure CLI
	auth, err := getAzCLIToken()
	if err == nil {
		return auth, nil
	}

	return nil, errors.New("missing auth: set AZDO_PAT or ADO_PAT, or sign in with `az login`")
}

// getAzCLIToken retrieves an access token from Azure CLI.
func getAzCLIToken() (*Auth, error) {
	// Azure DevOps resource ID
	const azureDevOpsResource = "499b84ac-1321-427f-aa17-267ca6975798"

	cmd := exec.Command("az", "account", "get-access-token",
		"--resource", azureDevOpsResource,
		"--query", "accessToken",
		"-o", "tsv",
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("az cli failed: %w", err)
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return nil, errors.New("az cli returned empty token")
	}

	return &Auth{Scheme: "Bearer", Token: token}, nil
}
