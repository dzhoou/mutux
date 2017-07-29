package server19

import (
	"crypto/tls"
	"net"
	"net/http"
)

type Server struct {
	http.Server
	//nextProtoOnce sync.Once // guards setupHTTP2_* init
	//nextProtoErr  error     // result of http2.ConfigureServer if used
}

func (srv *Server) ServeTLS(l net.Listener, certFile, keyFile string) error {
	// Setup HTTP/2 before srv.Serve, to initialize srv.TLSConfig
	// before we clone it and create the TLS Listener.
	// if err := srv.setupHTTP2_ServeTLS(); err != nil {
	// 	return err
	// }

	config := cloneTLSConfig(srv.TLSConfig)
	if !strSliceContains(config.NextProtos, "http/1.1") {
		config.NextProtos = append(config.NextProtos, "http/1.1")
	}

	configHasCert := len(config.Certificates) > 0 || config.GetCertificate != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}
	}

	tlsListener := tls.NewListener(l, config)
	return srv.Serve(tlsListener)
}

func cloneTLSConfig(cfg *tls.Config) *tls.Config {
	if cfg == nil {
		return &tls.Config{}
	}
	return cfg.Clone()
}

func strSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// func (srv *Server19) setupHTTP2_ServeTLS() error {
// 	srv.nextProtoOnce.Do(srv.onceSetNextProtoDefaults)
// 	return srv.nextProtoErr
// }

// func (srv *Server19) onceSetNextProtoDefaults() {
// 	if strings.Contains(os.Getenv("GODEBUG"), "http2server=0") {
// 		return
// 	}
// 	// Enable HTTP/2 by default if the user hasn't otherwise
// 	// configured their TLSNextProto map.
// 	if srv.TLSNextProto == nil {
// 		conf := &http2Server{
// 			NewWriteScheduler: func() http2WriteScheduler { return http2NewPriorityWriteScheduler(nil) },
// 		}
// 		srv.nextProtoErr = http2ConfigureServer(srv, conf)
// 	}
// }
