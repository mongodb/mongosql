package mongodb

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/10gen/mongo-go-driver/mongo/private/auth"
	"github.com/10gen/mongo-go-driver/mongo/private/conn"
	"github.com/xdg/stringprep"
)

// SCRAMSHA256 is the mechanism name for SCRAM-SHA-256.
const (
	SCRAMSHA256         = "SCRAM-SHA-256"
	scramSHA256NonceLen = 24
)

var usernameSanitizer = strings.NewReplacer("=", "=3D", ",", "=2C")

func newScramSHA256Authenticator(cred *auth.Cred) (auth.Authenticator, error) {
	preppedPassword, err := stringprep.SASLprep.Prepare(cred.Password)
	if err != nil {
		return nil, err
	}
	return &ScramSHA256Authenticator{
		DB:       cred.Source,
		Username: cred.Username,
		Password: preppedPassword,
	}, nil
}

// ScramSHA256Authenticator uses the SCRAM-SHA-256 algorithm over SASL to authenticate a connection.
type ScramSHA256Authenticator struct {
	DB             string
	Username       string
	Password       string
	clientKey      []byte
	saltedPassword []byte
	storedKey      []byte
	serverKey      []byte
	NonceGenerator func([]byte) error
}

// Auth authenticates the connection.
func (a *ScramSHA256Authenticator) Auth(ctx context.Context, c conn.Connection) error {
	client := &scramSaslClient{
		username:       a.Username,
		password:       a.Password,
		nonceGenerator: a.NonceGenerator,
		clientKey:      a.clientKey,
		saltedPassword: a.saltedPassword,
		storedKey:      a.storedKey,
		serverKey:      a.serverKey,
	}

	err := auth.ConductSaslConversation(ctx, c, a.DB, client)
	if err != nil {
		return err
	}

	a.clientKey = client.clientKey
	a.saltedPassword = client.saltedPassword
	a.storedKey = client.storedKey
	a.serverKey = client.serverKey

	return nil
}

type scramSaslClient struct {
	username       string
	password       string
	nonceGenerator func([]byte) error
	clientKey      []byte
	saltedPassword []byte
	storedKey      []byte
	serverKey      []byte

	step               uint8
	clientNonce        []byte
	clientFirstMessage string
	serverSignature    []byte
}

func (c *scramSaslClient) Start() (string, []byte, error) {
	if err := c.generateClientNonce(); err != nil {
		return "", nil, err
	}

	username := usernameSanitizer.Replace(c.username)
	c.clientFirstMessage = "n=" + username + ",r=" + string(c.clientNonce)

	return SCRAMSHA256, []byte("n,," + c.clientFirstMessage), nil
}

func (c *scramSaslClient) Next(challenge []byte) ([]byte, error) {
	c.step++
	// See https://bit.ly/2K2qkZv for more.
	switch c.step {
	case 1:
		return c.step1(challenge)
	case 2:
		return c.step2(challenge)
	default:
		return nil, fmt.Errorf("unexpected server challenge")
	}
}

func (c *scramSaslClient) Completed() bool {
	return c.step >= 2
}

func (c *scramSaslClient) step1(challenge []byte) ([]byte, error) {
	fields := bytes.Split(challenge, []byte{','})
	if len(fields) != 3 {
		return nil, fmt.Errorf("invalid server response")
	}

	if !bytes.HasPrefix(fields[0], []byte("r=")) || len(fields[0]) < 2 {
		return nil, fmt.Errorf("invalid nonce")
	}

	combinedNonce := fields[0][2:]
	if !bytes.HasPrefix(combinedNonce, c.clientNonce) {
		return nil, fmt.Errorf("invalid nonce")
	}

	if !bytes.HasPrefix(fields[1], []byte("s=")) || len(fields[1]) < 6 {
		return nil, fmt.Errorf("invalid salt")
	}

	salt := make([]byte, base64.StdEncoding.DecodedLen(len(fields[1][2:])))
	saltLength, err := base64.StdEncoding.Decode(salt, fields[1][2:])
	if err != nil {
		return nil, fmt.Errorf("invalid salt")
	}
	salt = salt[:saltLength]

	if !bytes.HasPrefix(fields[2], []byte("i=")) || len(fields[2]) < 3 {
		return nil, fmt.Errorf("invalid iteration count")
	}

	iterationCount, err := strconv.Atoi(string(fields[2][2:]))
	if err != nil {
		return nil, fmt.Errorf("invalid iteration count")
	}

	if iterationCount < 4096 {
		return nil, fmt.Errorf("server returned an invalid iteration count")
	}

	/**
	* The client proof:
	* AuthMessage     := client-first-message +
							"," +
							server-first-message +
							"," +
							client-final-message-without-proof
	* SaltedPassword  := PBKDF2(password, salt, i)
	* ClientKey       := HMAC(SaltedPassword, "Client Key")
	* ServerKey       := HMAC(SaltedPassword, "Server Key")
	* StoredKey       := H(ClientKey)
	* ClientSignature := HMAC(StoredKey, AuthMessage)
	* ClientProof     := ClientKey XOR ClientSignature
	* ServerSignature := HMAC(ServerKey, AuthMessage)
	*/

	if c.saltedPassword == nil {
		c.saltedPassword, err = c.saltPassword(salt, iterationCount)
		if err != nil {
			return nil, err
		}
	}

	if c.clientKey == nil {
		c.clientKey, err = c.hmac(c.saltedPassword, "Client Key")
		if err != nil {
			return nil, err
		}
	}

	if c.storedKey == nil {
		c.storedKey, err = c.h(c.clientKey)
		if err != nil {
			return nil, err
		}
	}

	if c.serverKey == nil {
		c.serverKey, err = c.hmac(c.saltedPassword, "Server Key")
		if err != nil {
			return nil, err
		}
	}

	clientMessageWithoutProof := "c=biws,r=" + string(combinedNonce)
	authMessage := c.clientFirstMessage + "," + string(challenge) + "," + clientMessageWithoutProof
	clientSignature, err := c.hmac(c.storedKey, authMessage)
	if err != nil {
		return nil, err
	}
	c.serverSignature, err = c.hmac(c.serverKey, authMessage)
	if err != nil {
		return nil, err
	}
	clientProof := c.xor(c.clientKey, clientSignature)
	proof := "p=" + base64.StdEncoding.EncodeToString(clientProof)
	clientFinalMessage := clientMessageWithoutProof + "," + proof

	return []byte(clientFinalMessage), nil
}

func (c *scramSaslClient) step2(challenge []byte) ([]byte, error) {
	var hasV, hasE bool
	fields := bytes.Split(challenge, []byte{','})
	if len(fields) == 1 {
		hasV = bytes.HasPrefix(fields[0], []byte("v="))
		hasE = bytes.HasPrefix(fields[0], []byte("e="))
	}
	if hasE {
		return nil, fmt.Errorf(string(fields[0][2:]))
	}
	if !hasV {
		return nil, fmt.Errorf("invalid final message")
	}

	v := make([]byte, base64.StdEncoding.DecodedLen(len(fields[0][2:])))
	n, err := base64.StdEncoding.Decode(v, fields[0][2:])
	if err != nil {
		return nil, fmt.Errorf("invalid server verification")
	}
	v = v[:n]

	// Verify the server's proof.
	if !bytes.Equal(c.serverSignature, v) {
		return nil, fmt.Errorf("invalid server signature")
	}

	return nil, nil
}

func (c *scramSaslClient) generateClientNonce() error {
	if c.nonceGenerator != nil {
		c.clientNonce = make([]byte, scramSHA256NonceLen)
		return c.nonceGenerator(c.clientNonce)
	}

	buf := make([]byte, scramSHA256NonceLen)
	rand.Read(buf)

	c.clientNonce = make([]byte, base64.StdEncoding.EncodedLen(int(scramSHA256NonceLen)))
	base64.StdEncoding.Encode(c.clientNonce, buf)
	return nil
}

func (c *scramSaslClient) h(data []byte) ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func (c *scramSaslClient) hmac(data []byte, key string) ([]byte, error) {
	h := hmac.New(sha256.New, data)
	if _, err := h.Write([]byte(key)); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// This is just an implementation of the Password-Based Key Derivation Function 2 (aka PBKDF2).
func (c *scramSaslClient) saltPassword(salt []byte, iterCount int) ([]byte, error) {
	mac := hmac.New(sha256.New, []byte(c.password))
	_, err := mac.Write(salt)
	if err != nil {
		return nil, err
	}
	_, err = mac.Write([]byte{0, 0, 0, 1})
	if err != nil {
		return nil, err
	}
	ui := mac.Sum(nil)
	hi := make([]byte, len(ui))
	copy(hi, ui)
	for i := 1; i < iterCount; i++ {
		mac.Reset()
		_, err = mac.Write(ui)
		if err != nil {
			return nil, err
		}
		mac.Sum(ui[:0])
		for j, b := range ui {
			hi[j] ^= b
		}
	}
	return hi, nil
}

func (c *scramSaslClient) xor(a []byte, b []byte) []byte {
	result := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}
