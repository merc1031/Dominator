package setupclient

import (
	"crypto/tls"
	"github.com/Symantec/Dominator/lib/srpc"
	"os"
)

func setupTls(ignoreMissingCerts bool) error {
	if *certDirectory == "" {
		return nil
	}
	// Load certificates.
	certs, err := srpc.LoadCertificates(*certDirectory)
	if ignoreMissingCerts && os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	// Setup client.
	clientConfig := new(tls.Config)
	clientConfig.InsecureSkipVerify = true
	clientConfig.MinVersion = tls.VersionTLS12
	clientConfig.Certificates = certs
	srpc.RegisterClientTlsConfig(clientConfig)
	return nil
}
