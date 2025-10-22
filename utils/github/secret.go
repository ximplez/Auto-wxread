package github

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/ximplez/wxread/utils/http"
	"github.com/ximplez/wxread/utils/json_tool"
	"golang.org/x/crypto/nacl/box"
)

type ghSecretBody struct {
	EncryptedValue string `json:"encrypted_value"`
	KeyId          string `json:"key_id"`
}

func CreateOrUpdateGithubSecret(githubToken, repo, secretName, secretValue string) error {
	if githubToken == "" || repo == "" || secretName == "" || secretValue == "" {
		return nil
	}
	keyId, key, err := getGithubRepoPubKey(githubToken, repo)
	if err != nil {
		return err
	}

	secretWithPublicKey, err := encryptSecretWithPublicKey(secretValue, key, "")
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/actions/secrets/%s", repo, secretName)
	if _, r, err := http.Put(url, json_tool.ToJson(&ghSecretBody{
		EncryptedValue: secretWithPublicKey,
		KeyId:          keyId,
	}, false), buildGithubHeader(githubToken)); err != nil {
		return err
	} else if r != "" {
		return errors.New("上传secret失败：" + r)
	}

	return nil
}

func encryptSecretWithPublicKey(text, publicKey, keyId string) (string, error) {
	// Decode the public key from base64
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return "", err
	}

	// Decode the public key
	var publicKeyDecoded [32]byte
	copy(publicKeyDecoded[:], publicKeyBytes)

	// Encrypt the secret value
	encrypted, err := box.SealAnonymous(nil, []byte(text), (*[32]byte)(publicKeyBytes), nil)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	// Encode the encrypted value in base64
	encryptedBase64 := base64.StdEncoding.EncodeToString(encrypted)

	return encryptedBase64, nil
}
