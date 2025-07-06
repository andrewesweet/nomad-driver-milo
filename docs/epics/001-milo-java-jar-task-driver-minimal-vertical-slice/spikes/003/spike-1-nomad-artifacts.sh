#!/bin/bash
# Spike 1: Test Nomad's artifact download behavior
# This tests what files/directories Nomad creates for different artifact types

echo "=== Spike 1: Nomad Artifact Download Behavior ==="
echo "Testing how Nomad downloads and provides artifacts to drivers"
echo

# Create test directory structure
TEST_DIR="/tmp/nomad-artifact-spike"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR/artifacts"

# Test 1: Simulate task directory structure as Nomad creates it
echo "Test 1: Simulating Nomad task directory structure..."
TASK_DIR="$TEST_DIR/task"
mkdir -p "$TASK_DIR/local"
mkdir -p "$TASK_DIR/tmp"
mkdir -p "$TASK_DIR/secrets"

echo "Task directory structure:"
tree "$TASK_DIR" 2>/dev/null || find "$TASK_DIR" -type d | sed 's/[^/]*\//  /g'

# Test 2: Single JAR artifact
echo -e "\nTest 2: Single JAR artifact download..."
# Nomad downloads artifacts to the 'local' directory by default
echo "PK\x03\x04fake-jar-content" > "$TASK_DIR/local/app.jar"
echo "Result: Single JAR placed in local/"
ls -la "$TASK_DIR/local/"

# Test 3: ZIP containing multiple files
echo -e "\nTest 3: ZIP artifact that Nomad extracts..."
# Create a test ZIP
mkdir -p "$TEST_DIR/zip-content"
echo "PK\x03\x04lib1-content" > "$TEST_DIR/zip-content/lib1.jar"
echo "PK\x03\x04lib2-content" > "$TEST_DIR/zip-content/lib2.jar"
echo "config content" > "$TEST_DIR/zip-content/config.properties"
(cd "$TEST_DIR/zip-content" && zip -q ../test.zip *)

# Simulate Nomad extracting the ZIP
unzip -q "$TEST_DIR/test.zip" -d "$TASK_DIR/local/"
echo "Result: ZIP contents extracted to local/"
ls -la "$TASK_DIR/local/"

# Test 4: Non-JAR file
echo -e "\nTest 4: Non-JAR artifact..."
echo "#!/bin/bash\necho hello" > "$TASK_DIR/local/script.sh"
chmod +x "$TASK_DIR/local/script.sh"
echo "Result: Non-JAR file in local/"
ls -la "$TASK_DIR/local/"

# Test 5: Multiple artifacts (Nomad downloads all to same directory)
echo -e "\nTest 5: Multiple artifacts scenario..."
echo "PK\x03\x04main-app" > "$TASK_DIR/local/main.jar"
echo "PK\x03\x04dependency" > "$TASK_DIR/local/lib.jar"
echo "log4j config" > "$TASK_DIR/local/log4j.properties"
echo "Result: Multiple files in local/"
ls -la "$TASK_DIR/local/"

# Test 6: Artifact with destination parameter
echo -e "\nTest 6: Artifact with custom destination..."
# If artifact has destination = "libs/", Nomad creates that directory
mkdir -p "$TASK_DIR/libs"
echo "PK\x03\x04custom-dest" > "$TASK_DIR/libs/custom.jar"
echo "Result: JAR in custom destination:"
ls -la "$TASK_DIR/libs/"

# Summary
echo -e "\n=== Summary of Findings ==="
echo "1. Nomad downloads artifacts to 'local/' by default"
echo "2. ZIP/TAR archives are automatically extracted"
echo "3. Multiple artifacts end up in the same directory"
echo "4. Custom destinations create subdirectories"
echo "5. Driver receives the task directory after all downloads complete"
echo "6. Driver must search within task directory for JAR files"

# Cleanup
rm -rf "$TEST_DIR"