package evt

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/brandon1024/OpenEVT/internal/types"
)

// Envertec EVT800 microinverter client.
type Client struct {
	Address     string
	InverterID  string
	ReadTimeout time.Duration

	conn *net.TCPConn
}

var (
	ErrConnect        = errors.New("failed to connect to inverter")
	ErrPoll           = errors.New("failed to poll inverter")
	ErrAck            = errors.New("failed to ack inverter")
	ErrReadFrame      = errors.New("failed to read frame from inverter")
	ErrFrameDiscarded = errors.New("frame discarded")
)

// Setup a connection to the inverter.
//
// Call Close() to terminate the connection.
func (c *Client) Connect() error {
	if c.Address == "" {
		return errors.Join(ErrConnect, fmt.Errorf("address is empty"))
	}
	if c.InverterID == "" {
		return errors.Join(ErrConnect, fmt.Errorf("inverter ID is empty"))
	}

	addr, err := net.ResolveTCPAddr("tcp", c.Address)
	if err != nil {
		return errors.Join(ErrConnect, err)
	}

	c.conn, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		return errors.Join(ErrConnect, err)
	}

	err = c.conn.SetKeepAlive(true)
	if err != nil {
		c.conn.Close()
		return errors.Join(ErrConnect, err)
	}

	return nil
}

// Poll the inverter for it's state.
func (c *Client) Poll() error {
	poll, err := types.NewPollMessage(c.InverterID)
	if err != nil {
		return errors.Join(ErrPoll, err)
	}

	_, err = c.Write(poll)
	if err != nil {
		return errors.Join(ErrPoll, err)
	}

	return nil
}

// Acknowledge a message from the inverter.
func (c *Client) Acknowledge() error {
	ack, err := types.NewAckMessage(c.InverterID)
	if err != nil {
		return errors.Join(ErrAck, err)
	}

	_, err = c.Write(ack)
	if err != nil {
		return errors.Join(ErrAck, err)
	}

	return nil
}

// Read the next inverter status frame. Upon receipt, the message is acknowledged with 'Acknowledge()'.
//
// If a 'ReadTimeout' is configured on the client, ReadFrame will return an [os.ErrDeadlineExceeded] if the inverter
// doesn't send a message after the deadline.
//
// May return ErrFrameDiscarded if the message from the inverter is unrecognized, which can be safely ignored.
func (c *Client) ReadFrame(msg *types.InverterStatus) error {
	if c.ReadTimeout != time.Duration(0) {
		err := c.conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
		if err != nil {
			return errors.Join(ErrReadFrame, err)
		}
	}

	frame := make([]byte, 512)
	w, err := c.Read(frame)
	if err != nil {
		return errors.Join(ErrReadFrame, err)
	}

	err = msg.UnmarshalBinary(frame[0:w])
	if err != nil {
		return errors.Join(ErrReadFrame, ErrFrameDiscarded, err)
	}

	err = c.Acknowledge()
	if err != nil {
		return errors.Join(ErrReadFrame, err)
	}

	return nil
}

// Read raw data from the underlying TCP connection.
func (c *Client) Read(p []byte) (int, error) {
	return c.conn.Read(p)
}

// Write raw data to the underlying TCP connection.
func (c *Client) Write(p []byte) (int, error) {
	return c.conn.Write(p)
}

// Close the underlying TCP connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Generate a string representation of this client.
func (c *Client) String() string {
	if c.conn == nil {
		return "DISCONNECTED"
	}

	return fmt.Sprintf("CONNECTED [%s <--> %s]", c.conn.LocalAddr().String(), c.conn.RemoteAddr().String())
}
