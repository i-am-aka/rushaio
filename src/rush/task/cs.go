package task

import (
	"fmt"
	"gopkg.in/square/go-jose.v2"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/url"
	// "rush/net/http"
	"crypto/rand"
	"strings"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
)

func CSCardTypeNumberFromName(cardType string) string {
	m := map[string]string {
		"VISA": "001",
		"MASTER": "002",
		"AMEX": "003",
		"DISCOVER": "004",
	}
	return m[cardType]
}

type CSToken struct {
  Token string `json:"token"`
  CardType string `json:"cardType"`
  MaskedPan string `json:"maskedPan"`
}
func (cst *CSToken) IssuingBankBin() string {
  // TODO is this right?
  return strings.SplitN(cst.MaskedPan, "X", 2)[0]
}

type CSPayKeysDer struct {
  Algorithm string `json:"algorithm"`
  Format string `json:"format"`
  PublicKey *string `json:"publicKey,omitempty"`
  DicksPublicKey *string `json:"public_key,omitempty"`
}

type CSPayKeys struct {
  Der CSPayKeysDer `json:"der"`
  Jwk jose.JSONWebKey `json:"jwk"`
  KeyId *string `json:"keyId,omitempty"`
  DicksKeyId *string `json:"key_id,omitempty"`
  TokenizationDomain *string `json:"tokenizationDomain,omitempty"`
}

func (cpk *CSPayKeys) EncryptCard(cardNumber string) (string, error) {
	key := cpk.Jwk.Key.(*rsa.PublicKey)
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, key, []byte(cardNumber), []byte{})
	if err != nil {
    return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
	// pub, err := LoadPublicKey(keyBytes())
	// app.FatalIfError(err, "unable to read public key")

	// alg := jose.KeyAlgorithm(*encryptAlgFlag)
	// enc := jose.ContentEncryption(*encryptEncFlag)

	// crypter, err := jose.NewEncrypter(enc, jose.Recipient{Algorithm: alg, Key: pub}, nil)
	// app.FatalIfError(err, "unable to instantiate encrypter")

	// obj, err := crypter.Encrypt(readInput(*inFile))
	// app.FatalIfError(err, "unable to encrypt")

	// var msg string
	// if *encryptFullFlag {
	// 	msg = obj.FullSerialize()
	// } else {
	// 	msg, err = obj.CompactSerialize()
	// 	app.FatalIfError(err, "unable to serialize message")
	// }
	// return ""
	// ciphertext, err := EncryptOAEP(sha256.New(), rng, &test2048Key.PublicKey, secretMessage, label)
}

func (t *CheckoutTask) GetCsToken(payKeys CSPayKeys, card Card, origin string) (CSToken, error) {
	var csToken CSToken
	cardNumberEnc, err := payKeys.EncryptCard(card.Number)
	if err != nil {
		return csToken, err
	}

	client, err := t.newHttpClient()
	if err != nil {
		return csToken, err
	}

	csTokenUrl, err := url.Parse("https://flex.cybersource.com/cybersource/flex/v1/tokens")
	if err != nil {
		return csToken, err
	}
	csTokenHeaders := [][2]string {
		{"accept", "application/json"},
		{"origin", origin},
		{"content-type", "application/json"},
		{"sec-fetch-dest", "empty"},
	  {"sec-fetch-mode", "cors"},
	  {"sec-fetch-site", "same-origin"},
		{"referer", "https://flex.cybersource.com/"},
	}
	csTokenBodyMap := map[string]interface{} {
	    "cardInfo": map[string]string {
	        "cardExpirationMonth": fmt.Sprintf("%02d", card.ExpMonth),
	        "cardExpirationYear": fmt.Sprintf("%d", card.ExpYear),
	        "cardNumber": cardNumberEnc,
	        "cardType": CSCardTypeNumberFromName(card.Type),
	    },
	    "keyId": payKeys.KeyId,
	}
	csTokenBody, err := json.Marshal(csTokenBodyMap)
	if err != nil {
		return csToken, err
	}
	csTokenResp, err := t.doReq(client, t.makeReq("POST", csTokenUrl, &csTokenHeaders, nil, &csTokenBody))
	if err != nil {
		if csTokenResp != nil && csTokenResp.StatusCode == 400 {
			err = retry.Unrecoverable(errors.New("Card Rejected Pre-Authorization"))
		}
		return csToken, err
	}
	csTokenBytes, err := readBodyBytes(csTokenResp)
	if err != nil {
		return csToken, err
	}
	err = json.Unmarshal(csTokenBytes, &csToken)
	return csToken, err
}
