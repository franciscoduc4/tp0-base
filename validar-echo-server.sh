#!/bin/bash

# Variables de configuración
SERVER_CONTAINER_NAME="server"
SERVER_PORT=12345  # El puerto debe coincidir con el puerto configurado en tu servidor
TEST_MESSAGE="Hola, servidor"

# Utiliza docker exec para ejecutar el comando netcat dentro del contenedor del servidor
RESPONSE=$(docker exec -i "$SERVER_CONTAINER_NAME" sh -c "echo '$TEST_MESSAGE' | nc localhost $SERVER_PORT")

# Verifica si la respuesta coincide con el mensaje de prueba
if [ "$RESPONSE" == "$TEST_MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi
