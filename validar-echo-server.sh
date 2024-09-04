# #!/bin/bash
NETWORK_NAME="tp0_testing_net"
SERVER_HOST="server"
SERVER_PORT="12345"
MESSAGE="Hello Server"

RESPONSE=$(docker run --rm --network ${NETWORK_NAME} busybox sh -c "echo '${MESSAGE}' | nc ${SERVER_HOST} ${SERVER_PORT}")

if [ "$RESPONSE" = "$MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi