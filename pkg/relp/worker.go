package relp

import "fmt"
import "io"
import "log/slog"
import "net"
import "time"

import "github.com/thierry-f-78/go-relp"

import "github.com/thierry-f-78/relp-log-collector/pkg/config"
import "github.com/thierry-f-78/relp-log-collector/pkg/dispatch"

type relpReceiverMsgT struct {
	msg *relp.Message
	err error
}

// Just a go routine which receive log from RELP in bloxking mode
// and forward to worker unsing chan.
func relpReceiver(relpConn *relp.Relp, relpReceiverChan chan<- *relpReceiverMsgT) {
	var msg *relp.Message
	var err error

	for {
		msg, err = relpConn.ReceiveLog()
		relpReceiverChan <- &relpReceiverMsgT{msg: msg, err: err}
		if err != nil {
			return
		}
	}
}

// Handle RELP connection and process messages.
func relpWorker(relpConn *relp.Relp, conn net.Conn, log *slog.Logger) {
	var err error
	var relpReceiverChan chan *relpReceiverMsgT
	var relpReceiverMsg *relpReceiverMsgT
	var doTimeout bool
	var msg *relp.Message
	var spool []*relp.Message
	var count int
	var doClose bool

	// Create chan. Receiver and processor should work in the same thread
	// while processor has not terminated processing log, receiver should be
	// blocked, so we use a chan size of 0, hoping Go uses the same thread for
	// the 2 go routines
	relpReceiverChan = make(chan *relpReceiverMsgT)

	// Start RELP log reader
	go relpReceiver(relpConn, relpReceiverChan)

	// Reserve some room for spool
	spool = make([]*relp.Message, 0, config.Cf.Spool.MaxLogs)

	// Receive log message loop
	for {
		select {
		case relpReceiverMsg = <-relpReceiverChan:
			if relpReceiverMsg == nil {
				panic(fmt.Errorf("unexpected channel closed"))
			}
			msg = relpReceiverMsg.msg
			err = relpReceiverMsg.err
			if err != nil {
				// Send message if we encouter protocol error. Do not quit
				// in order to attempt to save received log and send acquires.
				// Do not send message if we enconter EOF
				if err != io.EOF {
					log.Info(fmt.Sprintf("RELP connection error: %s. flush logs and close connection", err.Error()))
				} else {
					log.Info(fmt.Sprintf("Client close connection"))
				}
				doClose = true
			}
			doTimeout = false
		case <-time.After(config.Cf.Spool.MaxIdle):
			msg = nil
			err = nil
			doTimeout = true
		}

		// Spool log
		if msg != nil {
			spool = append(spool, msg)
		}

		// Process flush if there are some logs in the spool and ont of these
		// condition true:
		//  * timeout
		//  * chan closed
		//  * close order
		if len(spool) > 0 && (doTimeout || doClose || uint(len(spool)) >= config.Cf.Spool.MaxLogs) {
			// Flush chunk of logs.
			if !flushSpool(spool, log) {
				// If the flush return an error, we close the connection without sending
				// ack. The client will sent logs again.
				log.Error(fmt.Sprintf("can't flush spool. close connection without ACK messages"))
				goto relpWorkerEnd
			}
			// We have successfully flush the log on the spool. first we notify
			// the dispatcher about job
			dispatch.Notify()

			// then we send ACK to the client. Note if we have encountered
			count = len(spool)
			for _, msg = range spool {
				err = relpConn.AnswerOk(msg)
				if err != nil {
					log.Error(fmt.Sprintf("can't send RELP ACK messages: %s. close connection, %d logs may be sent twice", err.Error(), count))
					goto relpWorkerEnd
				}
				count--
			}

			// Empty the spool of message, to fill again.
			spool = make([]*relp.Message, 0, config.Cf.Spool.MaxLogs)
		}

		// Process close if required
		if doClose {
			goto relpWorkerEnd
		}
	}

relpWorkerEnd:

	// Close connextions
	relpConn.Close()
	conn.Close()

	// if the end of connexion is not initied by the reader,
	// wait for end of reader which send its last message
	if !doClose {
		<-relpReceiverChan
	}

	// Close chan
	close(relpReceiverChan)
}
