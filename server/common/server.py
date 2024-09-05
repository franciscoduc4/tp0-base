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
            data = self._receive_full_message(client_sock)
            if not data:
                return

            bets = data.split('\n')
            bet_objects = []
            errors = False
            for bet_data in bets:
                bet_data = bet_data.strip()  
                if not bet_data: 
                    continue
                try:
                    fields = bet_data.split(',')
                    if len(fields) != 6:
                        raise ValueError("Incorrect number of fields")
                    agency, first_name, last_name, document, birthdate, number = fields
                    if not all([agency, first_name, last_name, document, birthdate, number]):
                        raise ValueError("Missing or empty fields in bet data")
                    
                    bet = Bet(agency, first_name, last_name, document, birthdate, number)
                    bet_objects.append(bet)
                    logging.info(f"action: apuesta_almacenada | result: success | dni: {document} | numero: {number}")

                except ValueError as e:
                    logging.error(f"action: receive_message | result: fail | error: {e}")
                    errors = True
                    break
            if errors:
                client_sock.send("Batch processing failed\n".encode('utf-8'))
                logging.info(f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}")
            else:
                store_bets(bet_objects)
                client_sock.send("Batch processed successfully\n".encode('utf-8'))

        except (OSError, ValueError) as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            client_sock.send("An error occurred during processing\n".encode('utf-8'))
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
    
    def _receive_full_message(self, sock):
        """
        Helper method to ensure full message is read from the socket
        """
        BUFFER_SIZE = 8192
        data = b''
        while True:
            part = sock.recv(BUFFER_SIZE)
            data += part
            if len(part) < BUFFER_SIZE:
                # Either 0 or end of data
                break
        return data.decode('utf-8')