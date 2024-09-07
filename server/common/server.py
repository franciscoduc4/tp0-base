import socket
import logging
import signal
import sys
import threading
from common.utils import Bet, store_bets, load_bets, has_won

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

        # Initialize state variables
        self._agencies_notified = set()
        self._drawn = False

        # Create a lock for synchronization
        self._lock = threading.Lock()

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
        Server loop to handle incoming connections and requests
        """
        while True:
            client_sock = self.__accept_new_connection()
            client_thread = threading.Thread(target=self.__handle_client_connection, args=(client_sock,))
            client_thread.start()

    def __handle_client_connection(self, client_sock):
        try:
            # Receive header with batch size
            message = self._receive_full_message(client_sock)

            self.__process_payload(client_sock, message)

        except (OSError, ValueError) as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            client_sock.send("An error occurred during processing\n".encode('utf-8'))
        finally:
            client_sock.close()
    
    def _receive_full_message(self, sock):
        BUFFER_SIZE = 8192
        data = b''
        header = sock.recv(2)
        message_size = int.from_bytes(header, byteorder='big')

        while True:
            part = sock.recv(min(BUFFER_SIZE, message_size))
            if not part:
                raise ConnectionError("Connection closed unexpectedly")
            data += part
            
            if len(part) < BUFFER_SIZE and len(data) >= message_size:
                break
        return data.decode('utf-8')

    def __accept_new_connection(self):
        """
        Accept new connections from clients
        """
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def __process_payload(self, client_sock, payload):
        """
        Process the received payload and handle bet storage or notifications
        """
        data = payload.strip()

        with self._lock:
            if self._drawn:
                logging.warning("action: sorteo_rechazado | result: fail | reason: draw already performed")
                client_sock.send("Sorteo ya realizado. No se pueden procesar más apuestas\n".encode('utf-8'))
                return

            if data.startswith("NOTIFY_BETS_FINISHED"):
                agency_id = data.split(' ')[1]
                self._agencies_notified.add(agency_id)
                if len(self._agencies_notified) == 5:
                    self._drawn = True
                    logging.info("action: sorteo | result: success")
                else:
                    logging.info(f"action: notificacion_recibida | result: success | agencia: {agency_id}")
                    logging.info(f"Agencias notificadas: {self._agencies_notified}")
                client_sock.send("Notificación recibida\n".encode('utf-8'))
                return

            if data.startswith("GET_WINNERS"):
                if not self._drawn:
                    client_sock.send("Sorteo no realizado aún\n".encode('utf-8'))
                    return
                
                agency_id = data.split(' ')[1]
                winners = self.__get_winners(agency_id)
                if winners:
                    response = "\n".join(winners)
                else:
                    response = "No hay ganadores para esta agencia"
                client_sock.send((response + "\n").encode('utf-8'))
                return

        # Process bets if the message is not a notification or winner request
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
            logging.info(f"action: apuesta_almacenada | result: success | cantidad: {len(bet_objects)}")

    def __get_winners(self, agency_id):
        """
        Retrieve the list of winners for a specific agency
        """
        winners = []
        for bet in load_bets():
            if bet.agency == int(agency_id) and has_won(bet):
                winners.append(bet.document)
        return winners
