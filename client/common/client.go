package common
import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	bets   []string 
}

// NewClient Initializes a new client receiving the configuration
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}

	return client
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

// StartClientLoop Send messages to the client
func (c *Client) StartClientLoop() {
	// Setup signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	// Start client loop
	go func() {
		for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
			if err := c.createClientSocket(); err != nil {
				break
			}
			for _, betBatch := range c.bets {
				fmt.Fprintf(c.conn, "%s\n", betBatch)
				msg, err := bufio.NewReader(c.conn).ReadString('\n')
				if err != nil {
					log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
					return
				}
				log.Infof("action: batch_enviado | result: success | client_id: %v | msg: %v", c.config.ID, msg)
			}
			c.conn.Close()
			time.Sleep(c.config.LoopPeriod)
		}
		log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
	}()
	// Wait for termination signal
	<-signalChan
	log.Infof("action: shutdown | result: in_progress")
	if c.conn != nil {
		c.conn.Close()
	}
	log.Infof("action: shutdown | result: success")
}

func (c *Client) SendBet(apuesta map[string]string) error {
	if err := c.createClientSocket(); err != nil {
		return err
	}
	defer c.conn.Close()


    // betMessage := fmt.Sprintf("AGENCIA=%s,NOMBRE=%s,APELLIDO=%s,DOCUMENTO=%s,NACIMIENTO=%s,NUMERO=%s",
    // c.config.ID, apuesta["NOMBRE"], apuesta["APELLIDO"], apuesta["DOCUMENTO"], apuesta["NACIMIENTO"], apuesta["NUMERO"])

    betMessage := fmt.Sprintf("%s,%s,%s,%s,%s,%s",
    c.config.ID, apuesta["NOMBRE"], apuesta["APELLIDO"], apuesta["DOCUMENTO"], apuesta["NACIMIENTO"], apuesta["NUMERO"])
    
    fmt.Fprintf(c.conn, "%s\n", betMessage)

	response, err := bufio.NewReader(c.conn).ReadString('\n')
	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return err
	}

	log.Infof("action: respuesta_recibida | result: success | client_id: %v | response: %v", c.config.ID, response)
	return nil
}
