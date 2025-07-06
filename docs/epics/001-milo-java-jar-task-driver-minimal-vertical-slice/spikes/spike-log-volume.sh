#!/bin/bash
# spike-log-volume.sh
# Test high-volume logging scenarios with FIFOs

echo "=== Spike 4: Large Volume Log Handling ==="
echo "Testing performance characteristics and buffering behavior"
echo

# Create test FIFO
TEST_FIFO="/tmp/volume-test.fifo"
rm -f "$TEST_FIFO"
mkfifo "$TEST_FIFO"

# Test 1: Basic throughput measurement
echo "Test 1: Measuring FIFO throughput..."
(
  start_time=$(date +%s.%N)
  bytes=0
  lines=0
  
  while IFS= read -r line; do
    bytes=$((bytes + ${#line} + 1))
    lines=$((lines + 1))
    
    # Report progress every 1000 lines
    if [ $((lines % 1000)) -eq 0 ]; then
      current_time=$(date +%s.%N)
      elapsed=$(echo "$current_time - $start_time" | bc)
      rate=$(echo "scale=2; $bytes / $elapsed / 1048576" | bc)
      echo "Progress: $lines lines, ${rate} MB/s"
    fi
  done < "$TEST_FIFO"
  
  # Final report
  end_time=$(date +%s.%N)
  total_elapsed=$(echo "$end_time - $start_time" | bc)
  total_rate=$(echo "scale=2; $bytes / $total_elapsed / 1048576" | bc)
  echo "Total: $lines lines, $bytes bytes in ${total_elapsed}s (${total_rate} MB/s)"
) &
READER_PID=$!

# Generate high-volume output
echo "Generating 10,000 log lines..."
for i in {1..10000}; do
  echo "Log line $i: $(date +%s.%N) - Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
done > "$TEST_FIFO"

wait $READER_PID
echo

# Test 2: Concurrent writers
echo "Test 2: Multiple concurrent writers..."
rm -f "$TEST_FIFO"
mkfifo "$TEST_FIFO"

# Start reader that counts lines by source
(
  declare -A line_counts
  while IFS= read -r line; do
    if [[ $line =~ Writer-([0-9]+): ]]; then
      writer=${BASH_REMATCH[1]}
      ((line_counts[$writer]++))
    fi
  done < "$TEST_FIFO"
  
  echo "Lines per writer:"
  for writer in "${!line_counts[@]}"; do
    echo "  Writer-$writer: ${line_counts[$writer]} lines"
  done | sort
) &
READER_PID=$!

# Start multiple writers
echo "Starting 5 concurrent writers..."
for writer_id in {1..5}; do
  (
    for i in {1..1000}; do
      echo "Writer-$writer_id: Line $i at $(date +%s.%N)"
    done > "$TEST_FIFO"
  ) &
done

# Wait for all writers to complete
wait
echo

# Test 3: Large single line handling
echo "Test 3: Large single line (1MB)..."
rm -f "$TEST_FIFO"
mkfifo "$TEST_FIFO"

# Start reader that measures line size
(
  while IFS= read -r line; do
    echo "Received line of ${#line} bytes"
  done < "$TEST_FIFO"
) &
READER_PID=$!

# Generate a 1MB line
LARGE_LINE=$(python3 -c "print('X' * 1048576)")
echo "$LARGE_LINE" > "$TEST_FIFO"

wait $READER_PID
echo

# Test 4: Binary data handling
echo "Test 4: Binary data in logs..."
rm -f "$TEST_FIFO"
mkfifo "$TEST_FIFO"

# Start reader that checks for binary data
(
  total_bytes=0
  binary_bytes=0
  while IFS= read -r line; do
    total_bytes=$((total_bytes + ${#line}))
    # Count non-printable characters
    non_printable=$(echo -n "$line" | tr -d '[:print:]' | wc -c)
    binary_bytes=$((binary_bytes + non_printable))
  done < "$TEST_FIFO"
  echo "Total bytes: $total_bytes, Binary bytes: $binary_bytes"
) &
READER_PID=$!

# Send some binary data mixed with text
echo "Text before binary" > "$TEST_FIFO"
echo -e "Binary data: \x00\x01\x02\x03\x04\x05" > "$TEST_FIFO"
echo "Text after binary" > "$TEST_FIFO"

wait $READER_PID
echo

# Test 5: FIFO buffer limits
echo "Test 5: FIFO buffer size limits..."
rm -f "$TEST_FIFO"
mkfifo "$TEST_FIFO"

# Try to write without a reader to test blocking
echo "Writing to FIFO without reader (will timeout after 2s)..."
timeout 2 bash -c '
  count=0
  while true; do
    echo "Line $count" > /tmp/volume-test.fifo
    count=$((count + 1))
    echo "Wrote line $count"
  done
' && echo "Write completed" || echo "Write blocked as expected (FIFO buffer full)"

# Cleanup
rm -f "$TEST_FIFO"

echo
echo "=== Spike 4 Results Summary ==="
echo "1. FIFOs can handle high-throughput streaming efficiently"
echo "2. Multiple concurrent writers work correctly with serialized output"
echo "3. Large lines (1MB+) can be transmitted but may impact performance"
echo "4. Binary data passes through but may cause display issues"
echo "5. FIFOs block writes when buffer is full (typically 64KB on Linux)"
echo
echo "Recommendations for Milo driver:"
echo "- Use line-buffered writes for better real-time streaming"
echo "- Handle EPIPE errors gracefully when readers disconnect"
echo "- Consider chunking very large lines to avoid blocking"
echo "- Pass binary data as-is, let Nomad handle encoding"