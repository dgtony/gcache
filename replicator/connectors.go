package replicator

import (
	"encoding/binary"
	"errors"
	"github.com/dgtony/gcache/utils"

	//remove
	//"fmt"

	"crypto/sha256"
	"io"
	"net"
	//"sync"
	"time"
)

/*
// rename -> MasterConn
type MasterConn struct {
	Conn             net.Conn
	IsConnected      bool
	MasterAddr       string
	MasterSecretHash []byte
	Timeout          time.Duration
	ReconMaxWait     time.Duration
	ReconMaxAttempts int
	sync.Mutex
}
*/

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
/*
func (c *MasterConn) connect() error {
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
*/

const (
	RECONN_MAX_ATTEMPTS = 10
	RECONN_MAX_WAIT     = 60 * time.Second
)

func ConnectMaster(masterAddr string, timeout time.Duration, secretHash []byte) net.Conn {
	for connAttempt := 0; connAttempt < RECONN_MAX_ATTEMPTS; connAttempt++ {
		// try to connect
		conn, err := net.DialTimeout("tcp", masterAddr, timeout)
		if err == nil {
			// auth
			err = SendMsg(conn, ServiceMsg{Type: MSG_TYPE_AUTH_REQ, Payload: secretHash})
			if err != nil {
				//return nil, err
				panic(err)
			}
			resp, err := ReceiveMsg(conn, timeout)
			if err != nil {
				//return nil, err
				panic(err)
			}

			// parse response
			switch resp.Type {
			case MSG_TYPE_AUTH_OK:
				logger.Infof("master node connection established, attempts: %d", connAttempt+1)
				//return conn, nil
				return conn
			case MSG_TYPE_AUTH_DENY:
				//return nil, errors.New("master node authorization failure")
				panic("master node authorization failure")
			default:
				//return nil, errors.New(string(resp.Payload))
				panic(string(resp.Payload))
			}
		}
		// wait next reconnect
		time.Sleep(backoff(connAttempt, RECONN_MAX_WAIT))
	}
	panic("cannot connect to master node")
}

func GetMasterDump(conn net.Conn, timeout time.Duration) ([]byte, error) {
	err := SendMsg(conn, ServiceMsg{Type: MSG_TYPE_GET_DUMP})
	if err != nil {
		return nil, err
	}
	resp, err := ReceiveMsg(conn, timeout)
	if err != nil {
		return nil, err
	}

	switch resp.Type {
	case MSG_TYPE_DUMP:
		return resp.Payload, nil
	case MSG_TYPE_ERR:
		return nil, errors.New(string(resp.Payload))
	default:
		logger.Errorf("unexpected master response, message type: %d", resp.Type)
		return nil, errors.New("unexpected master response")
	}
}

/*
// ???
func (c *MasterConn) ConnectMaster() {
	attempt := 0
	for {
		err := c.connect()
		if err == nil {
			logger.Infof("connection to master node established, attempts: %d", attempt+1)
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

	// TODO send auth with MasterKey hash
	// receive AUTH_OK

	return nil

}

func (c *MasterConn) GetDump() ([]byte, error) {
	c.Lock()
	if !c.IsConnected {
		c.Unlock()
		return nil, errors.New("not connected")
	}
	err := SendMsg(c.Conn, msg)
	resp, err := ReceiveMsg(c.Conn, timeout)
	c.Unlock()
	//asdsa

}
*/

/////////////////////

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

func handleSlaveConn(conn net.Conn, r *Replicator) {
	// auth phase
	authMsg, err := ReceiveMsg(conn, CONN_AUTH_WAIT)
	if err != nil {
		logger.Debugf("slave authentication failed: %s", err)
		return
	}
	if authMsg.Type != MSG_TYPE_AUTH_REQ || !utils.CompareByteSlices(authMsg.Payload, r.MasterSecretHash) {
		logger.Debug("slave authentication failed: bad auth")
		SendMsg(conn, ServiceMsg{Type: MSG_TYPE_AUTH_DENY})
		return
	}

	// auth ok - proceed communication
	SendMsg(conn, ServiceMsg{Type: MSG_TYPE_AUTH_OK})
	logger.Debugf("slave connected: %s", conn.RemoteAddr())

	// waiting for requests
	for {
		msg, err := ReceiveMsg(conn, CONN_MAX_IDLE)
		if err != nil {
			if err == io.EOF {
				logger.Debugf("slave disconnected: %s", conn.RemoteAddr())
			} else {
				// timeout fired
				logger.Debugf("disconnect idle slave: %s", conn.RemoteAddr())
				conn.Close()
			}
			return
		}
		// process message
		switch msg.Type {
		case MSG_TYPE_GET_DUMP:
			r.Lock()
			dump := r.CacheDump
			r.Unlock()
			if err := SendMsg(conn, ServiceMsg{Type: MSG_TYPE_DUMP, Payload: dump}); err != nil {
				// FIXME: mb close connection?
				logger.Debugf("error sending dump to slave: %s", err)
			}

		default:
			logger.Debugf("unsupported message from slave node %s, type: %d", conn.RemoteAddr(), msg.Type)
			SendMsg(conn, ServiceMsg{Type: MSG_TYPE_ERR, Payload: []byte("unsupported command")})
		}
	}

}

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
	MAX_MSG_SIZE = 4294967295
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
	err := binary.Write(conn, binary.BigEndian, msgLen)
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
		return ServiceMsg{}, err
	}

	// get length prefix
	var msgLen uint32
	err := binary.Read(io.LimitReader(conn, 4), binary.BigEndian, &msgLen)
	if err != nil {
		return ServiceMsg{}, err
	}

	if msgLen > MAX_MSG_SIZE {
		return ServiceMsg{}, errors.New("message length exceeds limit")
	} else if msgLen < 1 {
		return ServiceMsg{}, errors.New("wrong message length")
	}

	// get message type
	var msgType MsgType
	err = binary.Read(io.LimitReader(conn, 1), binary.BigEndian, &msgType)
	if err != nil {
		return ServiceMsg{}, err
	}

	// read message payload from buffer
	msgBuff := make([]byte, msgLen-1)
	_, err = io.ReadFull(conn, msgBuff)
	if err != nil {
		return ServiceMsg{}, err
	}
	// FIXME use LimitReader again for msgType to avoid reallocation
	//return ServiceMsg{Type: uint8(msgBuff[0]), Payload: msgBuff[1:]}, nil
	return ServiceMsg{Type: msgType, Payload: msgBuff}, nil
}

/* internal stuff */

// exponential backoff
func backoff(attempt int, maxWait time.Duration) time.Duration {
	wait := time.Duration((utils.Pow(2, attempt) - 1)) * time.Second
	if wait > maxWait {
		return maxWait
	}
	return wait
}

func getSecretHash(secret string) []byte {
	secretHash := make([]byte, 32)
	for i, v := range sha256.Sum256([]byte(secret)) {
		secretHash[i] = v
	}
	return secretHash
}
