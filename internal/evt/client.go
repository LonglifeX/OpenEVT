package evt

import (
	"errors"
	"fmt"
	"net"

	"github.com/brandon1024/evt-client/internal/types"
)

// Envertec EVT800 microinverter client.
type Client struct {
	Address    string
	InverterID string

	conn *net.TCPConn
}

var (
	ErrFrameDiscarded = errors.New("frame discarded")
)

// Setup a connection to the inverter.
//
// Call Close() to terminate the connection.
func (c *Client) Connect() error {
	addr, err := net.ResolveTCPAddr("tcp", c.Address)
	if err != nil {
		return err
	}

	c.conn, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}

	err = c.conn.SetKeepAlive(true)
	if err != nil {
		c.conn.Close()
		return err
	}

	return nil
}

// Poll the inverter for it's state.
func (c *Client) Poll() error {
	poll, err := types.NewPollMessage(c.InverterID)
	if err != nil {
		return err
	}

	_, err = c.Write(poll)
	if err != nil {
		return err
	}

	return nil
}

// Acknowledge a message from the inverter.
func (c *Client) Acknowledge() error {
	ack, err := types.NewAckMessage(c.InverterID)
	if err != nil {
		return err
	}

	_, err = c.Write(ack)
	if err != nil {
		return err
	}

	return nil
}

// Read the next inverter status frame. Upon receipt, the message is acknowledged with 'Acknowledge()'.
//
// May return ErrFrameDiscarded if the message from the inverter is unrecognized, which can be safely ignored.
func (c *Client) ReadFrame(msg *types.InverterStatus) error {
	frame := make([]byte, 512)
	w, err := c.Read(frame)
	if err != nil {
		return err
	}

	err = msg.UnmarshalBinary(frame[0:w])
	if err != nil {
		return errors.Join(ErrFrameDiscarded, err)
	}

	err = c.Acknowledge()
	if err != nil {
		return err
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
		return "EVT DISCONNECTED"
	}

	return fmt.Sprintf("EVT CONNECTED [%s <--> %s]", c.conn.LocalAddr().String(), c.conn.RemoteAddr().String())
}
