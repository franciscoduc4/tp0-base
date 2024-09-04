import sys

def generate_docker_compose(output_file, num_clients):
    with open(output_file, 'w') as f:
        f.write("version: '3.8'\n")
        f.write("services:\n")
        f.write("  server:\n")
        f.write("    container_name: server\n")
        f.write("    image: server:latest\n")
        f.write("    entrypoint: python3 /main.py\n")
        f.write("    environment:\n")
        f.write("      - PYTHONUNBUFFERED=1\n")
        f.write("    networks:\n")
        f.write("      - testing_net\n")
        f.write("    volumes:\n")
        f.write("      - ./server/config.ini:/config.ini\n")
        
        for i in range(1, num_clients + 1):
            f.write(f"\n  client{i}:\n")
            f.write(f"    container_name: client{i}\n")
            f.write("    image: client:latest\n")
            f.write("    entrypoint: /client\n")
            f.write("    environment:\n")
            f.write(f"      - CLI_ID={i}\n")
            f.write("    networks:\n")
            f.write("      - testing_net\n")
            f.write("    depends_on:\n")
            f.write("      - server\n")
            f.write("    volumes:\n")
            f.write(f"      - ./client/config.yaml:/config.yaml\n")
        
        f.write("\nnetworks:\n")
        f.write("  testing_net:\n")
        f.write("    ipam:\n")
        f.write("      driver: default\n")
        f.write("      config:\n")
        f.write("        - subnet: 172.25.125.0/24\n")
    
    print(f"Archivo Docker Compose generado en {output_file} con {num_clients} clientes.")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Uso: python3 mi-generador.py <nombre_del_archivo_de_salida> <cantidad_de_clientes>")
        sys.exit(1)
    
    output_file = sys.argv[1]
    num_clients = int(sys.argv[2])
    
    generate_docker_compose(output_file, num_clients)
