package client

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/hashicorp/boundary/api/authtokens"
	nkeyring "github.com/jefferai/keyring"
	zkeyring "github.com/zalando/go-keyring"
)

const (
	NoneKeyring          = "none"
	AutoKeyring          = "auto"
	WincredKeyring       = "wincred"
	PassKeyring          = "pass"
	KeychainKeyring      = "keychain"
	SecretServiceKeyring = "secret-service"

	DefaultTokenName = "default"
	LoginCollection  = "login"
	PassPrefix       = "HashiCorp_Boundary"

	EnvToken        = "BOUNDARY_TOKEN"
	EnvTokenName    = "BOUNDARY_TOKEN_NAME"
	EnvKeyringType  = "BOUNDARY_KEYRING_TYPE"
	StoredTokenName = "HashiCorp Boundary Auth Token"
)

func getBoundaryToken() (string, error) {
	token := os.Getenv(EnvToken)
	if len(token) == 0 {
		keyringType, tokenName, err := discoverKeyringTokenInfo()
		if err != nil {
			return "", err
		}
		authToken, err := readTokenFromKeyring(keyringType, tokenName)
		if err != nil {
			return "", err
		}
		return authToken.Token, nil
	} else {
		return token, nil
	}
}

func discoverKeyringTokenInfo() (string, string, error) {
	// Stops the underlying library from invoking a dbus call that ends up
	// freezing some systems
	os.Setenv("DISABLE_KWALLET", "1")

	tokenName := DefaultTokenName

	// Set so we can look it up later when printing out curl strings
	os.Setenv(EnvTokenName, tokenName)

	var foundKeyringType bool
	keyringType := os.Getenv(EnvKeyringType)
	switch runtime.GOOS {
	case "windows":
		switch keyringType {
		case AutoKeyring, WincredKeyring, PassKeyring:
			foundKeyringType = true
			if keyringType == AutoKeyring {
				keyringType = WincredKeyring
			}
		}
	case "darwin":
		switch keyringType {
		case AutoKeyring, KeychainKeyring, PassKeyring:
			foundKeyringType = true
			if keyringType == AutoKeyring {
				keyringType = KeychainKeyring
			}
		}
	default:
		switch keyringType {
		case AutoKeyring, SecretServiceKeyring, PassKeyring:
			foundKeyringType = true
			if keyringType == AutoKeyring {
				keyringType = PassKeyring
			}
		}
	}

	if !foundKeyringType {
		return "", "", fmt.Errorf("Given keyring type %q is not valid, or not valid for this platform", EnvKeyringType)
	}

	var available bool
	switch keyringType {
	case WincredKeyring, KeychainKeyring:
		available = true
	case PassKeyring, SecretServiceKeyring:
		avail := nkeyring.AvailableBackends()
		for _, a := range avail {
			if keyringType == string(a) {
				available = true
			}
		}
	}

	if !available {
		return "", "", fmt.Errorf("Keyring type %q is not available on this machine. For help with setting up keyrings, see https://www.boundaryproject.io/docs/api-clients/cli.", keyringType)
	}

	os.Setenv(EnvKeyringType, keyringType)

	return keyringType, tokenName, nil
}

func readTokenFromKeyring(keyringType, tokenName string) (*authtokens.AuthToken, error) {
	var token string
	var err error

	switch keyringType {
	case NoneKeyring:
		return nil, nil

	case WincredKeyring, KeychainKeyring:
		token, err = zkeyring.Get(StoredTokenName, tokenName)
		if err != nil {
			if err == zkeyring.ErrNotFound {
				fmt.Println("No saved credential found, continuing without")
			} else {
				fmt.Printf(fmt.Sprintf("Error reading auth token from keyring: %s\n", err))
				fmt.Printf("Token must be provided via BOUNDARY_TOKEN env var or -token flag. Reading the token can also be disabled via -keyring-type=none.\n")
			}
			token = ""
		}

	default:
		krConfig := nkeyring.Config{
			LibSecretCollectionName: LoginCollection,
			PassPrefix:              PassPrefix,
			AllowedBackends:         []nkeyring.BackendType{nkeyring.BackendType(keyringType)},
		}

		kr, err := nkeyring.Open(krConfig)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error opening keyring: %s", err))
			fmt.Println("Token must be provided via BOUNDARY_TOKEN env var or -token flag. Reading the token can also be disabled via -keyring-type=none.")
			break
		}

		item, err := kr.Get(tokenName)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error fetching token from keyring: %s", err))
			fmt.Println("Token must be provided via BOUNDARY_TOKEN env var or -token flag. Reading the token can also be disabled via -keyring-type=none.")
			break
		}

		token = string(item.Data)
	}

	if token != "" {
		tokenBytes, err := base64.RawStdEncoding.DecodeString(token)
		switch {
		case err != nil:
			fmt.Println(fmt.Errorf("Error base64-unmarshaling stored token from system credential store: %w", err).Error())
		case len(tokenBytes) == 0:
			fmt.Println("Zero length token after decoding stored token from system credential store")
		default:
			var authToken authtokens.AuthToken
			if err := json.Unmarshal(tokenBytes, &authToken); err != nil {
				fmt.Println(fmt.Sprintf("Error unmarshaling stored token information after reading from system credential store: %s", err))
			} else {
				return &authToken, nil
			}
		}
	}
	return nil, err
}

func TokenIdFromToken(token string) (string, error) {
	split := strings.Split(token, "_")
	if len(split) < 3 {
		return "", errors.New("Unexpected stored token format")
	}
	return strings.Join(split[0:2], "_"), nil
}
