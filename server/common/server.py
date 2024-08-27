import socket
import logging
import signal
import sys
from common.utils import Bet, store_bets  


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

        # Register signal handler for graceful shutdown       
        signal.signal(signal.SIGTERM, self._shutdown)

    def _shutdown(self, signum, frame):
        """Handle shutdown signal"""
        logging.info("action: shutdown | result: in_progress")
        self._server_socket.close()
        logging.info("action: shutdown | result: success")
        sys.exit(0)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        # TODO: Modify this program to handle signal to graceful shutdown
        # the server
        while True:
            client_sock = self.__accept_new_connection()
            self.__handle_client_connection(client_sock)

    def __handle_client_connection(self, client_sock):
            """
            Read message from a specific client socket and closes the socket

            If a problem arises in the communication with the client, the
            client socket will also be closed
            """
            try:
                # Receive data from client
                data = client_sock.recv(1024).rstrip().decode('utf-8')
                # Parse the bet data sent by the client (assumed to be in CSV format)
                agency, first_name, last_name, document, birthdate, number = data.split(',')
                
                # Create Bet object
                bet = Bet(agency, first_name, last_name, document, birthdate, number)
                
                # Store bet
                store_bets([bet])
                
                # Log the successful storage of the bet
                logging.info(f'action: apuesta_almacenada | result: success | dni: {document} | numero: {number}')
                
                # Send a confirmation back to the client
                client_sock.send("Bet stored successfully\n".encode('utf-8'))
            except (OSError, ValueError) as e:
                logging.error(f"action: receive_message | result: fail | error: {e}")
            finally:
                client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c