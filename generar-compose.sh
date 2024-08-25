#!/bin/bash

# Verifico que se hayan proporcionado los argumentos necesarios
if [ $# -ne 2 ]; then
  echo "Uso: $0 <nombre del archivo de salida> <cantidad de clientes>"
  exit 1
fi

OUTPUT_FILE=$1
CLIENT_COUNT=$2

# Inicializo el archivo de salida con la configuración básica de docker-compose
cat <<EOL > $OUTPUT_FILE
version: '3.8'

services:
  server:
    build: ./server
    ports:
      - "12345:12345"
    networks:
      - mynetwork
EOL

# Agrego la configuración de cada cliente
for ((i=1; i<=CLIENT_COUNT; i++))
do
  cat <<EOL >> $OUTPUT_FILE
  client$i:
    build: ./client
    environment:
      - CLIENT_ID=$i
    depends_on:
      - server
    networks:
      - mynetwork
EOL
done

# Agrego la red al final del archivo
cat <<EOL >> $OUTPUT_FILE

networks:
  mynetwork:
    driver: bridge
EOL

echo "Archivo $OUTPUT_FILE generado con $CLIENT_COUNT clientes."
