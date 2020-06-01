package registry

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/micro/go-micro/v2/registry"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"microcore/registry/etcd"
	"os"
)

func getTLSConf() *tls.Config {
	var registryTLS *tls.Config = nil
	if os.Getenv("MICRO_ENABLE_REGISTRY_TLS") == "true" {
		cert, e := tls.LoadX509KeyPair(os.Getenv("MICRO_REGISTRY_TLS_CERT"), os.Getenv("MICRO_REGISTRY_TLS_KEY"))
		if e != nil {
			log.Warn("couldn't load x509 key pair:" + e.Error())
		} else {
			caData, err := ioutil.ReadFile(os.Getenv("MICRO_REGISTRY_TLS_CA"))
			if err != nil {
				log.Warn("couldn't load ca cert")
			} else {
				pool := x509.NewCertPool()
				pool.AppendCertsFromPEM(caData)
				registryTLS = &tls.Config{
					Certificates: []tls.Certificate{cert},
					RootCAs: pool,
				}
			}
		}
	}
	return registryTLS
}

func NewRegistryFromEnv(opts ...registry.Option) registry.Registry {
	op := make([]registry.Option, 0)
	op = append(op, opts...)
	op = append(op, registry.Addrs(os.Getenv("MICRO_REGISTRY_ADDRESS")))
	tlsCnf := getTLSConf()
	if tlsCnf != nil {
		op = append(op, registry.TLSConfig(tlsCnf))
	}
	reg := os.Getenv("MICRO_REGISTRY")
	switch reg {
	case "etcd":
		return etcd.NewRegistry(op...)
	default:
		return registry.DefaultRegistry
	}
}
