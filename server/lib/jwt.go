package lib

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const (
	ClaimSub    = "sub"
	ClaimIat    = "iat"
	ClaimExp    = "exp"
	ClaimIssuer = "iss"
)

const (
	HeaderFieldAlg = "alg"
	headerFieldTyp = "typ"
)

const (
	AlgoES256 = "ES256"
	TypeJWT   = "JWT"
)

type JWTProvider interface {
	Sign(claims map[string]any) (string, error)
	Verify(token string) (map[string]any, error)
}

type OAuthEC256JWTAuthenticator struct {
	privKey *ecdsa.PrivateKey
	pubKey  *ecdsa.PublicKey
	issuer  string
	expiry  time.Duration
}

func NewOAuthEC256JWTAuthenticator(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, issuer string,
	expiry time.Duration) *OAuthEC256JWTAuthenticator {
	return &OAuthEC256JWTAuthenticator{
		privKey: privateKey,
		pubKey:  publicKey,
		issuer:  issuer,
		expiry:  expiry,
	}
}

func (g *OAuthEC256JWTAuthenticator) Sign(claims map[string]any) (string, error) {
	// 1. Check subject claim
	sub, ok := claims[ClaimSub]
	if !ok || sub == "" {
		return "", fmt.Errorf("invalid subject claim")
	}

	now := time.Now()

	claims[ClaimIssuer] = g.issuer
	claims[ClaimIat] = now.Unix()
	claims[ClaimExp] = now.Add(g.expiry).Unix()

	bodyBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("jwt body failed: %w", err)
	}

	header := map[string]string{
		HeaderFieldAlg: AlgoES256,
		headerFieldTyp: TypeJWT,
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("jwt header failed: %w", err)
	}

	bodyEncoded := b64Encode(bodyBytes)
	headerEncoded := b64Encode(headerBytes)

	unsignedToken := headerEncoded + "." + bodyEncoded

	hasher := sha256.New()
	n, err := hasher.Write([]byte(unsignedToken))
	if n != len(unsignedToken) {
		return "", fmt.Errorf("jwt generation hashing failed")
	}
	if err != nil {
		return "", fmt.Errorf("jwt generation hashing failed: %w", err)
	}

	hashed := hasher.Sum(nil)

	// ECDSA signing
	r, s, err := ecdsa.Sign(rand.Reader, g.privKey, hashed)
	if err != nil {
		return "", fmt.Errorf("jwt signing failed: %w", err)
	}

	params := g.privKey.Curve.Params()
	byteSize := params.BitSize / 8 // Since we only consider P-256, We can divide by 8 withou any issues

	signature := make([]byte, 2*byteSize)
	r.FillBytes(signature[0:byteSize])
	s.FillBytes(signature[byteSize:])

	return unsignedToken + "." + b64Encode(signature), nil
}

func (g *OAuthEC256JWTAuthenticator) Verify(token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid jwt token format")
	}

	headerBytes, err := b64Decode(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decoding header failed")
	}
	var header map[string]string
	err = json.Unmarshal(headerBytes, &header)
	if err != nil {
		return nil, fmt.Errorf("invald jwt header")
	}

	if header[HeaderFieldAlg] != AlgoES256 {
		return nil, fmt.Errorf("unsupported jwt algorithm")
	}

	if header[headerFieldTyp] != TypeJWT {
		return nil, fmt.Errorf("not a jwt")
	}

	unsignedToken := parts[0] + "." + parts[1]
	hasher := sha256.New()
	n, err := hasher.Write([]byte(unsignedToken))
	if err != nil {
		return nil, fmt.Errorf("jwt hash calculation failed: %w", err)
	}
	if n != len(unsignedToken) {
		return nil, fmt.Errorf("jwt hash calculation faild")
	}

	hashed := hasher.Sum(nil)

	sig, err := b64Decode(parts[2])
	if err != nil {
		return nil, fmt.Errorf("signature decode failed: %w", err)
	}

	byteSize := len(sig) / 2
	r := new(big.Int).SetBytes(sig[:byteSize])
	s := new(big.Int).SetBytes(sig[byteSize:])

	if !ecdsa.Verify(g.pubKey, hashed, r, s) {
		return nil, fmt.Errorf("invalid signature")
	}

	bodyBytes, err := b64Decode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decoding body failed: %w", err)
	}

	var body map[string]any
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		return nil, fmt.Errorf("reading body failed: %w", err)
	}

	exp, ok := body[ClaimExp]
	if !ok {
		return nil, fmt.Errorf("exp field is missing")
	}
	expAt, ok := exp.(float64)
	if !ok {
		return nil, fmt.Errorf("exp field is invalid")
	}

	now := time.Now()
	expTime := time.Unix(int64(expAt), 0)

	iat, ok := body[ClaimIat]
	if !ok {
		return nil, fmt.Errorf("iat field is missing")
	}
	iatVal, ok := iat.(float64)
	if !ok {
		return nil, fmt.Errorf("iat field is invalid")
	}

	issuedTime := time.Unix(int64(iatVal), 0)

	if issuedTime.After(expTime) {
		return nil, fmt.Errorf("iat is after exp")
	}

	if expTime.After(now) {
		body[ClaimExp] = int64(expAt)
		body[ClaimIat] = int64(iatVal)
		return body, nil
	}

	return nil, fmt.Errorf("token expired")
}

func b64Encode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func b64Decode(data string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(data)
}