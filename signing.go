package main

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

func MakeHostname(signer Signer, hostName string, sourceIp string, sourcePort string) (string, int) {
	originalHostnameBuff := &bytes.Buffer{}
	originalHostnameBuff.Write(String(hostName))

	forgeSplit := strings.SplitN(hostName, "\000", 2)
	modifiedHostname := forgeSplit[0] + "///" + sourceIp + ":" + sourcePort + "///" + strconv.FormatInt(time.Now().Unix()/1000, 10)

	signed, err := signer.Sign([]byte(modifiedHostname))
	if err != nil {
		fmt.Errorf("could not sign request: %v", err)
	}
	sig := base64.StdEncoding.EncodeToString(signed)

	forgeBuilt := ""
	if len(forgeSplit) > 1 {
		forgeBuilt = "\000" + forgeSplit[1]
	}

	encodedHost := modifiedHostname + "///" + sig + forgeBuilt

	newHostname := &bytes.Buffer{}
	newHostname.Write(String(encodedHost))

	return encodedHost, len(newHostname.Bytes()) - len(originalHostnameBuff.Bytes())
}

func LoadSigner(path string) (Signer, error) {
	var err error
	signer, err := loadPrivateKey(path)
	if err != nil {
		return nil, fmt.Errorf("private key is damaged")
	}

	log.Info("Loaded private key")
	return signer, nil
}

// loadPrivateKey loads an parses a PEM encoded private key file.
func loadPrivateKey(path string) (Signer, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parsePrivateKey(dat)
}

// parsePublicKey parses a PEM encoded private key.
func parsePrivateKey(pemBytes []byte) (Signer, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("ssh: no key found")
	}

	var rawKey interface{}
	switch block.Type {
	case "RSA PRIVATE KEY":
		rsa, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rawKey = rsa
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %q", block.Type)
	}
	return newSignerFromKey(rawKey)
}

// A Signer is can create signatures that verify against a public key.
type Signer interface {
	// Sign returns raw signature for the given data. This method
	// will apply the hash specified for the keytype to the data.
	Sign(data []byte) ([]byte, error)
}

func newSignerFromKey(k interface{}) (Signer, error) {
	var sshKey Signer
	switch t := k.(type) {
	case *rsa.PrivateKey:
		sshKey = &rsaPrivateKey{t}
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %T", k)
	}
	return sshKey, nil
}

type rsaPrivateKey struct {
	*rsa.PrivateKey
}

// Sign signs data with rsa-sha256
func (r *rsaPrivateKey) Sign(data []byte) ([]byte, error) {
	h := sha512.New()
	h.Write(data)
	d := h.Sum(nil)
	return rsa.SignPKCS1v15(rand.Reader, r.PrivateKey, crypto.SHA512, d)
}
