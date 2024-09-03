#!/bin/bash

# Defino las variables
NETWORK_NAME="echo_server_network"
SERVER_CONTAINER_NAME="echo_server"
CLIENT_CONTAINER_NAME="echo_client"
MESSAGE="Hello, EchoServer!"
SERVER_IMAGE="server:latest"  

# Elimina la red docker si ya existe
docker network rm $NETWORK_NAME 2>/dev/null || true

# Elimina el contenedor del servidor si ya existe
docker rm -f $SERVER_CONTAINER_NAME 2>/dev/null || true

# Elimina el contenedor del cliente si ya existe
docker rm -f $CLIENT_CONTAINER_NAME 2>/dev/null || true

# Crea una nueva red docker
docker network create $NETWORK_NAME

# Corre el contenedor del echoserver
docker run -d --name $SERVER_CONTAINER_NAME --network $NETWORK_NAME $SERVER_IMAGE

# Verifica el estado del contenedor del servidor
if [ "$(docker ps -q -f name=$SERVER_CONTAINER_NAME)" ]; then
  echo "Servidor iniciado correctamente."
else
  echo "Error: El contenedor del servidor no se está ejecutando."
  echo "action: test_echo_server | result: fail"
  exit 1
fi

# Corre un contenedor temporal con netcat y verificar respuesta
docker run --rm --name $CLIENT_CONTAINER_NAME --network $NETWORK_NAME busybox sh -c "
  echo '$MESSAGE' | nc $SERVER_CONTAINER_NAME 12345
" > response.txt

RESPONSE=$(cat response.txt)

if [ "$MESSAGE" = "$RESPONSE" ]; then
  echo "action: test_echo_server | result: success"
else
  echo "action: test_echo_server | result: fail"
fi

# Limpio los recursos
docker rm -f $SERVER_CONTAINER_NAME
docker network rm $NETWORK_NAME
rm response.txt