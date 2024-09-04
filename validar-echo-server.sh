# #!/bin/bash
NETWORK_NAME="tp0_testing_net"

if ! docker network inspect $NETWORK_NAME >/dev/null 2>&1; then
    echo "Network $NETWORK_NAME not found"
    exit 1
fi

SERVER_HOST="server"
SERVER_PORT="12345"
MESSAGE="Hello EchoServer"

RESPONSE=$(docker run --rm --network ${NETWORK_NAME} busybox sh -c "echo '${MESSAGE}' | nc ${SERVER_HOST} ${SERVER_PORT}")

if [ "$RESPONSE" = "$MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi
