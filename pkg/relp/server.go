package relp

import "fmt"
import "log/slog"
import "net"

import "github.com/thierry-f-78/go-relp"

import "github.com/thierry-f-78/relp-log-collector/pkg/config"

var listener net.Listener
var relpOpts *relp.Options

func InitRELPServer() error {
	var err error
	var action int
	var acl config.ACL

	// Create RELP options
	relpOpts = &relp.Options{
		Tls:           relp.Opt_tls_connection,
		Certificate:   config.Cf.RELP.Certificate,
		PrivateKey:    config.Cf.RELP.PrivateKey,
		CACertificate: config.Cf.RELP.CA,
	}
	for _, acl = range config.Cf.RELP.ACL {
		switch acl.Action {
		case "accept":
			action = relp.Acl_accept
		case "reject":
			action = relp.Acl_reject
		default:
			return fmt.Errorf("Unknown ACL action: %q. Expect 'accept' ou 'reject'", acl.Action)
		}
		relpOpts.CnAcl = append(relpOpts.CnAcl, relp.Acl{
			Value:  acl.Value,
			Action: action,
		})
	}

	// Validate RELP options
	relpOpts, err = relp.ValidateOptions(relpOpts)
	if err != nil {
		return fmt.Errorf("Can't configure RELP: %s", err.Error())
	}

	// Create listen connection
	listener, err = net.Listen("tcp", config.Cf.RELP.Listen)
	if err != nil {
		return fmt.Errorf("Can't listen on %q: %s", config.Cf.RELP.Listen, err.Error())
	}

	return nil
}

func StartRELPServer(inheritLog *slog.Logger) {
	var err error
	var conn net.Conn
	var log *slog.Logger
	var logWorker *slog.Logger
	var relpConn *relp.Relp
	var connRemoteGeneric interface{}
	var connRemote *net.TCPAddr
	var ok bool

	log = inheritLog.WithGroup("relp")

	for {

		// Accept connection
		conn, err = listener.Accept()
		if err != nil {
			conn.Close()
			log.Error(fmt.Sprintf("can't accept incoming connection on %q: %s", config.Cf.RELP.Listen, err.Error()))
			continue
		}

		// Append IP address to logs
		connRemoteGeneric = conn.RemoteAddr()
		connRemote, ok = connRemoteGeneric.(*net.TCPAddr)
		if ok {
			logWorker = log.With(slog.String("ip", connRemote.IP.String()))
		} else {
			logWorker = log
		}

		// Create RELP from net.Conn
		relpConn, err = relp.NewTcp(conn, relpOpts)
		if err != nil {
			conn.Close()
			logWorker.Info(fmt.Sprintf("open incoming RELP protocol error: %s. close connection", err.Error()))
			return
		}

		// Start worker
		go relpWorker(relpConn, conn, logWorker)
	}
}
