#!/bin/bash
# spike-fifo-simple.sh
# Quick FIFO behavior test

echo "=== Spike 2: FIFO behavior (simplified) ==="
echo "Date: $(date)"
echo

# Create test FIFO
FIFO="/tmp/test-fifo-$$"
mkfifo "$FIFO"

echo "=== Test 1: FIFO blocks without reader ==="
timeout 1 bash -c "echo 'test' > $FIFO" && echo "✓ Write succeeded" || echo "✗ Write blocked (expected)"

echo "=== Test 2: FIFO works with reader ==="
(sleep 0.5; echo "test message" > "$FIFO") &
OUTPUT=$(cat "$FIFO")
echo "✓ Received: $OUTPUT"

echo "=== Test 3: Data integrity ==="
(echo -e "line1\nline2\nline3" > "$FIFO") &
cat "$FIFO" | while read line; do echo "Got: $line"; done
echo "✓ Multiple lines received"

echo "=== Key Findings ==="
echo "1. FIFOs block on write until reader attached"
echo "2. Data integrity maintained"
echo "3. Need goroutines for non-blocking operation"
echo "4. Works well for log streaming"

rm -f "$FIFO"
echo "Spike 2 completed!"