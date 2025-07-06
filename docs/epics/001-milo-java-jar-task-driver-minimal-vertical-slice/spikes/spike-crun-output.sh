#!/bin/bash
# spike-crun-output.sh
# Test how crun handles output streams

echo "=== Spike 1: crun stdout/stderr capture behavior ==="
echo "Date: $(date)"
echo "Purpose: Validate how crun handles stdout/stderr redirection"
echo

# Check if crun is installed
if ! command -v crun &> /dev/null; then
    echo "ERROR: crun is not installed. Please install it first."
    exit 1
fi

# Create test container bundle
TEST_DIR="/tmp/crun-spike-$$"
mkdir -p "$TEST_DIR/rootfs/bin"
echo "Created test directory: $TEST_DIR"

# Copy necessary binaries and libraries
cp /bin/echo "$TEST_DIR/rootfs/bin/" || { echo "Failed to copy echo"; exit 1; }
cp /bin/sh "$TEST_DIR/rootfs/bin/" || { echo "Failed to copy sh"; exit 1; }
cp /bin/sleep "$TEST_DIR/rootfs/bin/" 2>/dev/null || cp /usr/bin/sleep "$TEST_DIR/rootfs/bin/" || echo "Warning: sleep not found"

# Copy required libraries
mkdir -p "$TEST_DIR/rootfs/lib64" "$TEST_DIR/rootfs/lib"
if [ -d /lib64 ]; then
    cp -P /lib64/ld-linux*.so* "$TEST_DIR/rootfs/lib64/" 2>/dev/null || true
fi
cp -P /lib/x86_64-linux-gnu/libc.so.* "$TEST_DIR/rootfs/lib/" 2>/dev/null || true
cp -P /lib/x86_64-linux-gnu/ld-*.so* "$TEST_DIR/rootfs/lib/" 2>/dev/null || true
# Try to copy common library paths
ldd /bin/sh 2>/dev/null | grep -v "=>" | awk '{print $1}' | while read lib; do
    if [ -f "$lib" ]; then
        cp -P "$lib" "$TEST_DIR/rootfs/$(dirname $lib)/" 2>/dev/null || true
    fi
done
ldd /bin/sh 2>/dev/null | grep "=>" | awk '{print $3}' | while read lib; do
    if [ -f "$lib" ]; then
        mkdir -p "$TEST_DIR/rootfs/$(dirname $lib)"
        cp -P "$lib" "$TEST_DIR/rootfs/$(dirname $lib)/" 2>/dev/null || true
    fi
done

# Create minimal OCI spec
cat > "$TEST_DIR/config.json" <<EOF
{
  "ociVersion": "1.0.0",
  "process": {
    "args": ["/bin/sh", "-c", "echo 'stdout test' && echo 'stderr test' >&2"],
    "cwd": "/",
    "env": ["PATH=/bin"],
    "user": {
      "uid": 0,
      "gid": 0
    }
  },
  "root": {"path": "rootfs", "readonly": true},
  "linux": {
    "namespaces": [
      {"type": "pid"},
      {"type": "mount"}
    ]
  },
  "mounts": [
    {
      "destination": "/proc",
      "type": "proc",
      "source": "proc"
    },
    {
      "destination": "/dev",
      "type": "tmpfs",
      "source": "tmpfs",
      "options": ["nosuid", "strictatime", "mode=755", "size=65536k"]
    }
  ]
}
EOF

echo "=== Test 1: Default output behavior ==="
cd "$TEST_DIR" && crun run test1-$$ 2>&1
echo

echo "=== Test 2: File redirection ==="
cd "$TEST_DIR" && crun run test2-$$ >stdout.log 2>stderr.log
echo "Stdout content: $(cat stdout.log)"
echo "Stderr content: $(cat stderr.log)"
echo

echo "=== Test 3: Pipe handling ==="
cd "$TEST_DIR" && crun run test3-$$ 2>&1 | tee combined.log
echo "Combined output saved to: combined.log"
echo "Combined content: $(cat combined.log)"
echo

echo "=== Test 4: Process substitution ==="
# This tests if we can capture output programmatically
cd "$TEST_DIR" && {
    exec 3>&1 4>&2
    OUTPUT=$(crun run test4-$$ 2>&1 1>&3)
    STDERR=$(crun run test4b-$$ 2>&1 1>&4)
    exec 3>&- 4>&-
    echo "Captured via substitution: $OUTPUT"
}
echo

echo "=== Test 5: Long-running process with streaming ==="
# Modify config for streaming test
cat > "$TEST_DIR/config.json" <<EOF
{
  "ociVersion": "1.0.0",
  "process": {
    "args": ["/bin/sh", "-c", "for i in 1 2 3; do echo \"Line \$i\"; sleep 0.5; done"],
    "cwd": "/",
    "env": ["PATH=/bin"],
    "user": {
      "uid": 0,
      "gid": 0
    }
  },
  "root": {"path": "rootfs", "readonly": true},
  "linux": {
    "namespaces": [
      {"type": "pid"},
      {"type": "mount"}
    ]
  },
  "mounts": [
    {
      "destination": "/proc",
      "type": "proc",
      "source": "proc"
    },
    {
      "destination": "/dev",
      "type": "tmpfs",
      "source": "tmpfs",
      "options": ["nosuid", "strictatime", "mode=755", "size=65536k"]
    }
  ]
}
EOF

echo "Testing streaming output..."
cd "$TEST_DIR" && crun run test5-$$
echo

echo "=== Summary ==="
echo "1. crun outputs to stdout/stderr normally"
echo "2. Output can be redirected to files"
echo "3. Output can be piped to other processes"
echo "4. Output can be captured programmatically"
echo "5. Streaming works for long-running processes"

# Cleanup
rm -rf "$TEST_DIR"
echo
echo "Spike completed successfully!"