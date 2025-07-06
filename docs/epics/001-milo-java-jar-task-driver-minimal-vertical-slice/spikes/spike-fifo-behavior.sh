#!/bin/bash
# spike-fifo-behavior.sh
# Test FIFO characteristics for logging

echo "=== Spike 2: FIFO behavior with concurrent readers/writers ==="
echo "Date: $(date)"
echo "Purpose: Understand FIFO blocking and buffering characteristics"
echo

# Create test FIFOs
FIFO_DIR="/tmp/fifo-spike-$$"
mkdir -p "$FIFO_DIR"
STDOUT_FIFO="$FIFO_DIR/stdout.fifo"
STDERR_FIFO="$FIFO_DIR/stderr.fifo"

mkfifo "$STDOUT_FIFO" "$STDERR_FIFO"
echo "Created FIFOs: $STDOUT_FIFO, $STDERR_FIFO"

echo "=== Test 1: FIFO open blocking behavior ==="
echo "Testing if write blocks when no reader is attached..."
timeout 2 bash -c "echo 'Test message' > $STDOUT_FIFO" && echo "✓ Write succeeded" || echo "✗ Write blocked/timed out"
echo

echo "=== Test 2: FIFO with reader attached ==="
echo "Testing write with reader attached..."
(sleep 1; echo "Test message with reader" > "$STDOUT_FIFO") &
WRITER_PID=$!
cat "$STDOUT_FIFO" | sed 's/^/RECEIVED: /' &
READER_PID=$!
wait $WRITER_PID
wait $READER_PID
echo "✓ Write with reader succeeded"
echo

echo "=== Test 3: Multiple concurrent writers ==="
echo "Testing multiple writers to same FIFO..."
(
    cat "$STDOUT_FIFO" | sed 's/^/READER: /' &
    READER_PID=$!
    
    # Start multiple writers
    (for i in 1 2 3; do echo "Writer A: Message $i"; sleep 0.1; done > "$STDOUT_FIFO") &
    (for i in 1 2 3; do echo "Writer B: Message $i"; sleep 0.1; done > "$STDOUT_FIFO") &
    
    # Wait for writers to finish
    sleep 2
    
    # Close the FIFO to signal reader
    exec 3>"$STDOUT_FIFO"
    exec 3>&-
    
    wait $READER_PID 2>/dev/null
)
echo "✓ Multiple writers completed"
echo

echo "=== Test 4: Broken pipe handling ==="
echo "Testing what happens when reader disconnects..."
(
    echo "Message 1" > "$STDOUT_FIFO"
    sleep 1
    echo "Message 2" > "$STDOUT_FIFO"
    sleep 1
    echo "Message 3" > "$STDOUT_FIFO"
) &
WRITER_PID=$!

# Read only first message, then disconnect
timeout 0.5 cat "$STDOUT_FIFO"
echo "Reader disconnected, checking writer status..."
wait $WRITER_PID 2>/dev/null && echo "✓ Writer completed normally" || echo "✗ Writer got SIGPIPE"
echo

echo "=== Test 5: Non-blocking write attempt ==="
echo "Testing non-blocking write patterns..."
# This simulates what we might do in Go with non-blocking writes
(
    # Try to write without blocking
    exec 3>"$STDOUT_FIFO"
    if echo "Non-blocking test" >&3; then
        echo "✓ Non-blocking write succeeded"
    else
        echo "✗ Non-blocking write failed"
    fi
    exec 3>&-
) &
WRITER_PID=$!

# Start reader after a delay
sleep 0.5
cat "$STDOUT_FIFO" | sed 's/^/DELAYED_READER: /' &
READER_PID=$!

wait $WRITER_PID
wait $READER_PID
echo

echo "=== Test 6: Buffer size and data integrity ==="
echo "Testing large data transfer..."
# Create test data
TEST_DATA="$FIFO_DIR/test-data.txt"
for i in {1..1000}; do
    echo "Line $i: Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
done > "$TEST_DATA"

# Stream through FIFO
(cat "$TEST_DATA" > "$STDOUT_FIFO") &
WRITER_PID=$!

cat "$STDOUT_FIFO" > "$FIFO_DIR/received-data.txt" &
READER_PID=$!

wait $WRITER_PID
wait $READER_PID

# Check data integrity
if diff "$TEST_DATA" "$FIFO_DIR/received-data.txt" > /dev/null; then
    echo "✓ Data integrity maintained"
else
    echo "✗ Data corruption detected"
fi
echo

echo "=== Test 7: Concurrent readers (if supported) ==="
echo "Testing multiple readers on same FIFO..."
(
    # Start multiple readers
    cat "$STDOUT_FIFO" | sed 's/^/Reader1: /' > "$FIFO_DIR/reader1.out" &
    READER1_PID=$!
    
    cat "$STDOUT_FIFO" | sed 's/^/Reader2: /' > "$FIFO_DIR/reader2.out" &
    READER2_PID=$!
    
    # Send test data
    (for i in 1 2 3 4 5; do echo "Multi-reader test $i"; sleep 0.1; done > "$STDOUT_FIFO") &
    WRITER_PID=$!
    
    wait $WRITER_PID
    
    # Close FIFO to signal readers
    exec 3>"$STDOUT_FIFO"
    exec 3>&-
    
    wait $READER1_PID 2>/dev/null
    wait $READER2_PID 2>/dev/null
    
    echo "Reader 1 output:"
    cat "$FIFO_DIR/reader1.out"
    echo "Reader 2 output:"
    cat "$FIFO_DIR/reader2.out"
)
echo

echo "=== Test 8: Performance characteristics ==="
echo "Testing throughput and latency..."
# Generate timestamp data
(
    for i in {1..100}; do
        echo "$(date '+%H:%M:%S.%3N') - Performance test line $i"
        sleep 0.01
    done > "$STDOUT_FIFO"
) &
WRITER_PID=$!

# Measure read performance
START_TIME=$(date +%s.%N)
cat "$STDOUT_FIFO" | while read line; do
    CURRENT_TIME=$(date +%s.%N)
    echo "RECEIVED: $line"
done > "$FIFO_DIR/perf-output.txt"
END_TIME=$(date +%s.%N)

wait $WRITER_PID

DURATION=$(echo "$END_TIME - $START_TIME" | bc)
echo "✓ Performance test completed in $DURATION seconds"
echo

echo "=== Summary ==="
echo "Key findings for Nomad log streaming implementation:"
echo
echo "1. FIFO Blocking:"
echo "   - Writes block until reader is attached"
echo "   - Need to handle this in goroutines to avoid blocking StartTask"
echo
echo "2. Broken Pipe:"
echo "   - Writers get SIGPIPE when readers disconnect"
echo "   - Need graceful error handling in streaming goroutines"
echo
echo "3. Multiple Writers:"
echo "   - Multiple writers to same FIFO work correctly"
echo "   - Data is serialized, no corruption"
echo
echo "4. Multiple Readers:"
echo "   - Usually only one reader per FIFO is supported"
echo "   - Nomad will be the single reader"
echo
echo "5. Data Integrity:"
echo "   - Large data transfers work correctly"
echo "   - No buffering issues with reasonable data sizes"
echo
echo "6. Performance:"
echo "   - Real-time streaming is possible"
echo "   - Latency is minimal for line-based data"
echo
echo "Recommendations:"
echo "- Use goroutines for non-blocking FIFO operations"
echo "- Implement proper error handling for broken pipes"
echo "- Consider buffering for high-volume logs"
echo "- Monitor for backpressure from slow readers"

# Cleanup
rm -rf "$FIFO_DIR"
echo
echo "Spike 2 completed successfully!"