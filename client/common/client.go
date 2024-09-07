package common

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"io"
	"encoding/binary"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BatchMaxSize  int
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
func NewClient(config ClientConfig) *Client {
	return &Client{
		config: config,
	}
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID, err,
		)
		return err
	}
	c.conn = conn
	return nil
}

// NotifyBetsFinished Sends a notification to the server that all bets have been sent
func (c *Client) NotifyBetsFinished() error {
	if err := c.createClientSocket(); err != nil {
		return err
	}
	defer c.conn.Close()

	message := fmt.Sprintf("NOTIFY_BETS_FINISHED %s", c.config.ID)
	notifySize := uint16(len(message))
	header := make([]byte, 2)
	binary.BigEndian.PutUint16(header, notifySize)
	c.conn.Write(header)
	io.WriteString(c.conn, message)
	
	response, err := bufio.NewReader(c.conn).ReadString('\n')
	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return err
	}
	log.Infof("action: notify_bets_finished | result: success | client_id: %v | response: %v", c.config.ID, response)
	return nil
}

// GetWinners Requests the list of winners from the server
func (c *Client) GetWinners() (string) {
	if err := c.createClientSocket(); err != nil {
		return "error"
	}
	defer c.conn.Close()

	message := fmt.Sprintf("GET_WINNERS %s", c.config.ID)
	winnersMsgSize := uint16(len(message))
	header := make([]byte, 2)
	binary.BigEndian.PutUint16(header, winnersMsgSize)
	c.conn.Write(header)
	io.WriteString(c.conn, message)

	response, err := bufio.NewReader(c.conn).ReadString('\n')
	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return "error"
	}

	return response
}

// StartClientLoop Handles the client loop to process batches and handle signals
func (c *Client) StartClientLoop() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
			if err := c.createClientSocket(); err != nil {
				break
			}

			// Use a batch size determined by the configuration
			for i := 0; i < c.config.BatchMaxSize; i++ {
				// Logic to process each batch would be here

				// Close connection after each batch
				c.conn.Close()
				time.Sleep(c.config.LoopPeriod)
			}
		}
		log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
	}()
	<-signalChan
	log.Infof("action: shutdown | result: in_progress")
	if c.conn != nil {
		c.conn.Close()
	}
	log.Infof("action: shutdown | result: success")
}

func (c *Client) SendBets(bets []map[string]string) error {
    if err := c.createClientSocket(); err != nil {
        return err
    }
    defer c.conn.Close()

    var batch []string
    for _, bet := range bets {
        betMessage := strings.Join([]string{
            c.config.ID, bet["NOMBRE"], bet["APELLIDO"], bet["DOCUMENTO"], bet["NACIMIENTO"], bet["NUMERO"],
        }, ",")

        batch = append(batch, betMessage)

        if len(batch) >= c.config.BatchMaxSize {
            // Prepare the batch message
            batchMessage := strings.Join(batch, "\n")

			batchSize := uint16(len(batchMessage))
            header := make([]byte, 2)
            binary.BigEndian.PutUint16(header, batchSize)

			if _, err := c.conn.Write(header); err != nil {
                log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
                return err
            }

            // Send the batch message
            if _, err := io.WriteString(c.conn, batchMessage); err != nil {
                log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
                return err
            }

            // Read the response from the server
            response, err := bufio.NewReader(c.conn).ReadString('\n')
            if err != nil {
                log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
                return err
            }
            log.Infof("action: response_received | result: success | client_id: %v | response: %v", c.config.ID, response)
            batch = nil // Clear batch after sending
        }
    }

    if len(batch) > 0 {
        // Prepare the last batch message
        batchMessage := strings.Join(batch, "\n")

		batchSize := uint16(len(batchMessage))
        header := make([]byte, 2)
        binary.BigEndian.PutUint16(header, batchSize)

        // Send the header
        if _, err := c.conn.Write(header); err != nil {
            log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
            return err
        }

        // Send the batch message
        if _, err := io.WriteString(c.conn, batchMessage); err != nil {
            log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
            return err
        }

        // Read the response from the server
        response, err := bufio.NewReader(c.conn).ReadString('\n')
        if err != nil {
            log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
            return err
        }
        log.Infof("action: response_received | result: success | client_id: %v | response: %v", c.config.ID, response)
    }
    return nil
}