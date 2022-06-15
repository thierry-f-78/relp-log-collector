package utilities

import "net"
import "os"

func NotifySystemd() error {
	var notifySocket string
	var addr net.UnixAddr
	var conn net.Conn
	var err error

	// Get systemd socker from env NOTIFY_SOCKET, do nothing if not found
	notifySocket = os.Getenv("NOTIFY_SOCKET")
	if notifySocket == "" {
		return nil
	}

	// Send ready message
	addr = net.UnixAddr{
		Name: notifySocket,
		Net:  "unixgram",
	}
	conn, err = net.DialUnix("unixgram", nil, &addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte("READY=1"))
	if err != nil {
		return err
	}

	return nil
}
