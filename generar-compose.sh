#!/bin/bash

# Verifica que se hayan proporcionado los dos parámetros requeridos
if [ $# -ne 2 ]; then
  echo "Uso: $0 <nombre_del_archivo_de_salida> <cantidad_de_clientes>"
  exit 1
fi

# Asigna los parámetros a variables para mayor claridad
output_file=$1
num_clients=$2

# Inicia el contenido del archivo YAML con la configuración básica
cat <<EOL > $output_file
version: '3.8'
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net
EOL

# Agrega cada cliente al archivo YAML
for ((i=1; i<=$num_clients; i++)); do
  cat <<EOL >> $output_file

  client$i:
    container_name: client$i
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=$i
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
EOL
done

# Añade la configuración de la red al final del archivo YAML
cat <<EOL >> $output_file

networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
EOL

echo "Archivo Docker Compose generado en $output_file con $num_clients clientes."
