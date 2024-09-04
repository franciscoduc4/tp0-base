import sys

def generateCompose(fileName, clientsNumber):
    content = """
services:
"""

    server = """  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net

"""
    
    content += server


    for i in range(1, int(clientsNumber) + 1):
        client = f"""  client{i}:
    container_name: client{i}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={i}
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server

"""
        content += client

    networks =  """networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""

    content += networks

    with open(fileName, 'w') as file:
        file.write(content)

    print(f"OK")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        sys.exit(1)

    fileName = sys.argv[1]
    clientsNumber = sys.argv[2]

    generateCompose(fileName, clientsNumber)