## Ejericio 1

Para ejecutar el script y generar un archivo docker-compose.yaml con 5 clientes:

```
./generar-compose.sh docker-compose-dev.yaml 5
```

## Ejercicio 2

Solo modifique el docke-compose-dev.yaml agregando volumes

## Ejercicio 3

Para ejecutar la validación:
```
./validar-echo-server.sh
```

## Ejercicio 4

Para enviar la señal SIGTERM a un contenedor docker
```
docker kill --signal=SIGTERM <container_id>

```

## Ejercicio 5
### Cliente
El cliente recibe los valores de las variables de entorno que representan los detalles de la apuesta (NOMBRE, APELLIDO, DOCUMENTO, NACIMIENTO, NUMERO) y los empaqueta en un mensaje en formato CSV (valores separados por coma). Después de enviar el mensaje, espera la confirmación del servidor.

Se agregó un método SendBet en el cliente, que serializa los datos y los envía al servidor. Si la apuesta se almacena con éxito, se imprime el log con la acción y el resultado.

Al recibir la confirmación de que la apuesta fue enviada, se registra en el log:

`action: apuesta_enviada | result: success | dni: ${DOCUMENTO} | numero: ${NUMERO}`.
### Server
El servidor escucha conexiones de los clientes en un socket TCP. Cuando recibe una apuesta, la procesa y la almacena utilizando la función `store_bet(...)`, provista por la cátedra. Si la apuesta se almacena correctamente, el servidor imprime en el log:

`action: apuesta_almacenada | result: success | dni: ${DOCUMENTO} | numero: ${NUMERO}`.

El servidor también maneja varios casos de error, como el procesamiento de un batch de apuestas incorrecto, mediante logs y respuestas al cliente.

### Protocolo
El protocolo de comunicación entre el cliente y el servidor utiliza TCP y sigue una estructura de mensajes simple en formato CSV, donde cada campo está separado por comas. El flujo es el siguiente:

1. El cliente crea un mensaje con los campos de la apuesta (`ID de agencia`, `nombre`, `apellido`, `DNI`, `fecha de nacimiento`, `número de apuesta`).
2. El mensaje se envía al servidor mediante un socket TCP.
3. El servidor recibe el mensaje, lo deserializa y valida los campos.
4. Si los datos son válidos, el servidor almacena la apuesta y responde con un mensaje de éxito. Si no, envía un mensaje de error.
5. El cliente recibe la respuesta y registra el resultado en los logs.

**Serialización de Datos**:

Los datos se serializan en formato CSV, lo que asegura que cada mensaje enviado sea claro y fácil de parsear. Cada línea de apuesta contiene 6 campos, separados por comas.

**Manejo de Sockets**:

El cliente y el servidor manejan las conexiones utilizando el protocolo TCP. Se implementan medidas para evitar los problemas comunes de `short read` y `short write`.

## Ejercicio 6
Se añadió la capacidad de configurar el tamaño máximo de apuestas que se pueden enviar en un solo lote a través del archivo config.yaml del cliente
  `maxAmount: 100`

### Cliente
Ahora el cliente tiene un metodo llamado SendBets que le permite enviar multiples apuestas

- **Construcción del Mensaje**: Se crea un mensaje por cada apuesta, formateado como una cadena de texto, y se agrupan en un lote.
- **Envío del Lote**: Cuando el tamaño del lote alcanza el límite configurado, se envía al servidor. Si quedan apuestas después de procesar el lote, se envían en un último envío.

### Server
El cliente espera una respuesta del servidor después de cada envío de lote. Dependiendo de la respuesta, se registran mensajes en el log para indicar si el envío fue exitoso o fallido.

- **Validación de Apuestas**: Cada apuesta se valida para asegurar que contenga la cantidad correcta de campos y que no haya campos vacíos.
- **Registro de Apuestas**: Si todas las apuestas en el lote son válidas, se almacenan. Si alguna apuesta falla, se envía un mensaje de error al cliente.
- **Registro en Log**: Se registran los resultados de cada lote procesado en el log, indicando si fue exitoso o fallido, junto con la cantidad de apuestas procesadas.
## **Ejemplo de Uso**

1. **Configuración**: El archivo **`config.yaml`** debe contener la configuración necesaria, incluyendo el tamaño máximo de lote.
2. **Ejecución**: Al ejecutar el cliente, este leerá las apuestas desde un archivo CSV y enviará los lotes al servidor.
3. **Resultados**: El servidor procesará las apuestas y responderá al cliente, que registrará el resultado en su log.

## Ejercicio 7
Se añadió la función NotifyBetsFinished al cliente para que notifique al servidor cuando todas las apuestas han sido enviadas. El cliente también consulta la lista de ganadores tras la notificación usando la función GetWinners.

Una vez notificado el servidor, los clientes consultan la lista de ganadores y registran la cantidad en el log.


```
if err := NotifyEndOfBets(client, clientConfig); err != nil {
    os.Exit(1)
}

winners, err := GetWinners(client, clientConfig)
if err != nil {
    os.Exit(1)
}

log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", len(winners))
```

El servidor recibe notificaciones de cada agencia, y una vez que recibe las 5 notificaciones, procede con el sorteo. Usa las funciones load_bets(...) y has_won(...) para verificar las apuestas y devolver los ganadores a las agencias.

## Ejercicio 8
### Server
En el método run(), el servidor acepta conexiones de manera continua. Cuando un cliente se conecta, se crea un nuevo hilo para manejar esa conexión.

- **`threading.Thread(...)`**: Se utiliza para crear un nuevo hilo. Cada vez que el servidor acepta una nueva conexión, se crea un hilo con el método `target` configurado en `self.__handle_client_connection`, el cual maneja la lógica para procesar los mensajes del cliente.
- **`client_thread.start()`**: Inicia el nuevo hilo para procesar la conexión del cliente de forma independiente.

### Mecanismos de sincronización utilizados

1. **Lock de hilo (threading.Lock)**:
    - Se utiliza para garantizar que las secciones críticas, como la actualización del estado del sorteo (`self._drawn`) y la adición de agencias notificadas (`self._agencies_notified`), sean atómicas. Este mecanismo de sincronización asegura que solo un hilo pueda acceder a estos recursos compartidos al mismo tiempo, evitando condiciones de carrera.
    - Al recibir un mensaje de un cliente, las operaciones que involucran la manipulación del estado compartido (ejemplo, la notificación de agencias o la verificación de si el sorteo ya ha sido realizado) se realizan dentro de un bloque protegido por un `lock`, asegurando que los cambios no entren en conflicto cuando múltiples hilos están activos.
    