#!/bin/bash

# Usage: ./simple_load_test.sh [endpoint] [requests_per_second] [duration_seconds]

#default params
ENDPOINT=${1:-"http://localhost:8080/"}
RPS=${2:-50}
DURATION=${3:-30}

echo "========================================"
echo "AdaptLimit Simple Load Test"
echo "========================================"
echo "Target: $ENDPOINT"
echo "Rate: $RPS requests per second"
echo "Duration: $DURATION seconds"
echo "Starting in 3 seconds..."
echo "========================================"
sleep 3

echo "Test started at $(date)"

#run for a specific duration
for (( i=1; i<=$DURATION; i++ )); do
    for (( j=1; j<=$RPS; j++ )); do
        curl -s "$ENDPOINT" > /dev/null 2>&1 &
    done

    echo "Second $i: Sent $RPS requests"
    sleep 1
done

wait

echo "========================================"
echo "Test completed at $(date)"
echo "========================================"
echo "Total requests sent: $((RPS * DURATION))"
echo "Check server logs and metrics for results"
echo "========================================"

#current metrics
echo "Server metrics:"
curl -s "$ENDPOINT/../metrics"

echo ""
echo "========================================"
echo "Load test complete!"
echo "========================================"
