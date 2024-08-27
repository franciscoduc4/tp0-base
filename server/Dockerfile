FROM python:3.9.7-slim

# Copia el contenido de la carpeta local 'server' al contenedor en '/server'
COPY server/ /server/

# Establece el directorio de trabajo en el contenedor
WORKDIR /server

# Instala dependencias necesarias (si las hay)
RUN python -m unittest discover -s /server/tests -p "*.py"

# Define el comando de entrada
ENTRYPOINT ["python", "main.py"]
