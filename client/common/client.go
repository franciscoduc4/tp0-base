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

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	bet    string // Nueva propiedad para almacenar la apuesta
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}

	// Leer datos de apuesta desde variables de entorno
	nombre := os.Getenv("NOMBRE")
	apellido := os.Getenv("APELLIDO")
	documento := os.Getenv("DOCUMENTO")
	nacimiento := os.Getenv("NACIMIENTO")
	numero := os.Getenv("NUMERO")

	// Comprobar si todas las variables de entorno están presentes
	if nombre == "" || apellido == "" || documento == "" || nacimiento == "" || numero == "" {
		log.Critical("Faltan variables de entorno para la apuesta")
		os.Exit(1)
	}

	// Construir la apuesta en formato CSV para enviarla al servidor
	client.bet = fmt.Sprintf("%s,%s,%s,%s,%s", nombre, apellido, documento, nacimiento, numero)
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
			c.config.ID,
			err,
		)
		return err // Retorna el error para manejarlo en la llamada
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
			// Create the connection to the server in every loop iteration
			if err := c.createClientSocket(); err != nil {
				break
			}
			// Send the bet information to the server
			fmt.Fprintf(c.conn, "%s\n", c.bet)
			msg, err := bufio.NewReader(c.conn).ReadString('\n')
			c.conn.Close()

			if err != nil {
				log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err,
				)
				return
			}

			// Log the success of sending the bet
			log.Infof("action: apuesta_enviada | result: success | dni: %s | numero: %s",
				strings.Split(c.bet, ",")[2], // DNI
				strings.Split(c.bet, ",")[4], // Número
			)

			log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
				c.config.ID,
				msg,
			)

			// Wait a time between sending one message and the next one
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
