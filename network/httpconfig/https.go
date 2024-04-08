package httpconfig

import (
	"os"
    "crypto/tls"
)

func GetRequireHttp() (*tls.Config, error) {
	cert, err := os.ReadFile("ca.pem")
	if err != nil {
		return nil, err
	}
	key, err := os.ReadFile("privatekey.pem")
	if err != nil {
		return nil, err
	}

	certificate, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}
	return tlsConfig, nil
}