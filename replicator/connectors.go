package replicator

import (
	"encoding/binary"
	"errors"
	"github.com/dgtony/gcache/utils"

	//remove
	"fmt"

	"io"
	"net"
	"sync"
	"time"
)

// TODO change type??
// WTF is it in fact?
type MasterConn struct {
	Conn net.Conn
	Addr string
	//Timeout time.Duration
	sync.Mutex
}

// rename -> MasterConn
type SlaveConn struct {
	Conn             net.Conn
	IsConnected      bool
	MasterAddr       string
	Timeout          time.Duration
	ReconMaxWait     time.Duration
	ReconMaxAttempts int
	sync.Mutex
}

// TODO implement link control and reconnect to master

/*
   // checkout receive message error
   if err == io.EOF {
       // todo set to offline and reconnect

       logger("master conection lost")
       c.ConnectMaster()
   }
*/

// addr is a pair host:port
// fmt.Sprintf("%s:%d", addr, port)
func (c *SlaveConn) connect() error {
	conn, err := net.DialTimeout("tcp", c.Addr, c.Timeout)
	if err != nil {
		return err
	}

	c.Lock()
	c.Conn = conn
	c.IsConnected = true
	c.Unlock()
	return nil
}

func (c *SlaveConn) ConnectMaster() {
	attempt := 0
	for {
		err := c.connect()
		if err == nil {

			// TODO replace with logger
			fmt.Printf("connection to master node established, attempts: %d", attempt+1)

			break
		}

		attempt++
		if attempt > c.ReconMaxAttempts {
			// mb panic?
			return errors.New("cannot connect to master node")
		}

		// wait before reconnect
		time.Sleep(backoff(attempt, c.ReconMaxWait))
	}

	// TODO send auth with MasterKey
	// receive AUTH_OK

	return nil

}

/*
// previous version
func (c *SlaveConn) ConnectMaster() {
    go func() {
        attempt := 1
        for {
            err := c.connect()
            if err == nil {

                // TODO replace with logger
                fmt.Printf("connection to master node established, attempts: %d", attempt)

                break
            }
            // wait before reconnect
            time.Sleep(backoff(attempt, c.ReconMaxWait))
            attempt++
        }

        // TODO send auth with MasterKey
        // receive AUTH_OK

    }()
}
*/

// message types
const (
	// auth
	MSG_TYPE_AUTH_REQ  = 1
	MSG_TYPE_AUTH_OK   = 2
	MSG_TYPE_AUTH_DENY = 3
	// replication
	MSG_TYPE_GET_DUMP = 10
	MSG_TYPE_DUMP     = 11

	// errors
	MSG_TYPE_ERR = 255
)

const (
	MAX_MSG_SIZE = 4 * 1024 * 1024 * 1024
)

type ServiceMsg struct {
	Type    MsgType
	Payload []byte
}

type MsgType uint8

/* service message encoding scheme
+------------+---------+---------+
| len_prefix | msgType | payload |
+------------+---------+---------+
|   4 bytes  |  1 byte |  []byte |
+------------+---------+---------+
*/

func SendMsg(conn net.Conn, msg ServiceMsg) error {
	// length prefix
	msgLen := uint32(len(msg.Payload) + 1)
	err = binary.Write(conn, binary.BigEndian, msgLen)
	if err != nil {
		return err
	}

	// write message type
	err = binary.Write(conn, binary.BigEndian, msg.Type)
	if err != nil {
		return err
	}

	// write payload
	err = binary.Write(conn, binary.BigEndian, msg.Payload)
	if err != nil {
		return err
	}
	return nil
}

func ReceiveMsg(conn net.Conn, timeout time.Duration) (ServiceMsg, error) {
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return ServiceMsg, err
	}

	// get length prefix
	err = binary.Read(io.LimitReader(conn, 4), binary.BigEndian, &msgLen)
	if err != nil {
		return ServiceMsg{}, err
	}

	if msgLen > MAX_MSG_SIZE {
		return ServiceMsg{}, errors.New("message length exceeds limit")
	} else if msgLen < 1 {
		return ServiceMsg{}, errors.New("wrong message length")
	}

	// read entire message from buffer
	msgBuff := make([]byte, msgLen)
	_, err = io.ReadFull(conn, msgBuff)
	if err != nil {
		return ServiceMsg{}, err
	}
	// FIXME use LimitReader again for msgType to avoid reallocation
	return ServiceMsg{Type: msgBuff[0], Payload: msgBuff[1:]}, nil
}

/* internal stuff */

// exponential backoff
func backoff(attempt int, maxWait time.Duration) int {
	wait := (utils.Pow(2, attempt) - 1) * time.Second
	if wait > maxWait {
		return maxWait
	}
	return wait
}
