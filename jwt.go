package main

import (
	"fmt"
	"log"
	"time"

	_ "embed"

	jwt "github.com/golang-jwt/jwt/v5"
)

//go:embed private.pem
var privateKeyPEM []byte

//go:embed public.pem
var publicKeyPEM []byte

// How to generate private and public key:
// https://rietta.com/blog/openssl-generating-rsa-key-from-command/
//
// Private key:
// $ openssl genrsa -out private.pem 4096
//
// Public key:
// $ openssl rsa -in private.pem -outform PEM -pubout -out public.pem

func main() {
	// Create a JWT key pair.
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		log.Fatal("failed to parse private key:", err)
	}

	// Create a JWT claims object.
	now := time.Now()
	ttl := 1 * time.Hour
	claims := make(jwt.MapClaims)
	claims["dat"] = nil // Our custom data.
	//claims["sub"] = "john@mail.com"     // Our subject, normally user id or unique identifier.
	claims["sub"] = 1                   // WARN: it is important that this field is a string, otherwise you get really weird behaviours (see below).
	claims["exp"] = now.Add(ttl).Unix() // The expiration time after which the token must be disregarded.
	claims["iat"] = now.Unix()          // The time at which the token was issued.
	claims["nbf"] = now.Unix()          // The time before which the token must be disregarded.

	// Sign the JWT with the private key.
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
	if err != nil {
		log.Fatal("failed to sign claims:", err)
	}
	log.Println("got jwt token:", token)

	// Send the JWT to the second microservice.
	fmt.Println("Sending JWT to second microservice:", token)

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		log.Fatal("failed to parse public key:", err)
	}

	tok, err := jwt.Parse(token, func(jwtToken *jwt.Token) (any, error) {

		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", err)
		}
		return publicKey, nil
	})
	if err != nil {
		log.Fatal("failed to verify token", err)
	}
	{
		claims, ok := tok.Claims.(jwt.MapClaims)
		if !ok || !tok.Valid {
			log.Fatal("invalid token")
		}
		log.Println("claims:", claims)
		fmt.Println("JWT validated successfully.")

		// If you set the subject as non-string, you will get the result as float64
		// instead of int64 or int.
		// This is due to how json.Unmarshal works by storing the interface value
		// in float64 for JSON numbers [^1].
		//
		// [^1]: Reference: https://pkg.go.dev/encoding/json#Unmarshal
		s, ok := claims["sub"].(string)
		fmt.Println(s, ok)

		f, ok := claims["sub"].(float64)
		fmt.Println(f, ok)

		i64, ok := claims["sub"].(int64)
		fmt.Println(i64, ok)

		i, ok := claims["sub"].(int)
		fmt.Println(i, ok)
	}
}
