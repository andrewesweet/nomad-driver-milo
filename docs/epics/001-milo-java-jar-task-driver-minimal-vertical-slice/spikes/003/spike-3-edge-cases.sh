#!/bin/bash
# Spike 3: Test edge cases for artifact validation

echo "=== Spike 3: Edge Case Testing ==="
echo

TEST_DIR="/tmp/artifact-edge-cases"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR/task/local"

# Test 1: Case sensitivity variations
echo "Test 1: Case sensitivity (.jar, .JAR, .Jar, .jAr)..."
touch "$TEST_DIR/task/local/app.jar"
touch "$TEST_DIR/task/local/APP.JAR"
touch "$TEST_DIR/task/local/Config.Jar"
touch "$TEST_DIR/task/local/weird.jAr"
echo "Files created:"
ls -la "$TEST_DIR/task/local/" | grep -i jar
echo "Finding JARs (case-insensitive):"
find "$TEST_DIR/task/local" -iname "*.jar" -type f

# Test 2: Multiple JARs
echo -e "\nTest 2: Multiple JARs in directory..."
echo "PK\x03\x04main" > "$TEST_DIR/task/local/main.jar"
echo "PK\x03\x04lib1" > "$TEST_DIR/task/local/lib1.jar"
echo "PK\x03\x04lib2" > "$TEST_DIR/task/local/lib2.jar"
echo "Multiple JARs found:"
ls -la "$TEST_DIR/task/local/"*.jar
echo "Challenge: Which JAR to execute?"

# Test 3: Symbolic links
echo -e "\nTest 3: Symbolic links to JAR files..."
mkdir -p "$TEST_DIR/jars"
echo "PK\x03\x04real-jar" > "$TEST_DIR/jars/real.jar"
ln -s "$TEST_DIR/jars/real.jar" "$TEST_DIR/task/local/linked.jar"
echo "Symlink created:"
ls -la "$TEST_DIR/task/local/linked.jar"
echo "Following symlink:"
readlink -f "$TEST_DIR/task/local/linked.jar"
echo "Security consideration: Should we follow symlinks?"

# Test 4: JAR within extracted archive
echo -e "\nTest 4: Nested JAR scenario..."
mkdir -p "$TEST_DIR/task/local/extracted"
echo "PK\x03\x04nested" > "$TEST_DIR/task/local/extracted/nested.jar"
echo "Directory structure after extraction:"
find "$TEST_DIR/task/local" -name "*.jar" -type f

# Test 5: Non-standard extensions
echo -e "\nTest 5: Files that might be JARs but wrong extension..."
echo "PK\x03\x04war-file" > "$TEST_DIR/task/local/app.war"
echo "PK\x03\x04ear-file" > "$TEST_DIR/task/local/app.ear"
echo "PK\x03\x04renamed" > "$TEST_DIR/task/local/app.zip"
echo "Files with ZIP structure but not .jar:"
file "$TEST_DIR/task/local/"* | grep -v ".jar"

# Test 6: Hidden files
echo -e "\nTest 6: Hidden JAR files..."
echo "PK\x03\x04hidden" > "$TEST_DIR/task/local/.hidden.jar"
echo "Hidden files:"
ls -la "$TEST_DIR/task/local/" | grep "^\."

# Test 7: Filename edge cases
echo -e "\nTest 7: Problematic filenames..."
touch "$TEST_DIR/task/local/spaces in name.jar"
touch "$TEST_DIR/task/local/special!@#\$%chars.jar"
touch "$TEST_DIR/task/local/unicode-文件.jar"
echo "Problematic filenames:"
ls -la "$TEST_DIR/task/local/" | tail -3

# Summary
echo -e "\n=== Edge Case Findings ==="
echo "1. Case sensitivity: Use case-insensitive search (strings.ToLower or filepath.Match)"
echo "2. Multiple JARs: Need clear selection logic or error"
echo "3. Symlinks: Should validate target, not link itself"
echo "4. Nested JARs: Search recursively or only top-level?"
echo "5. Wrong extensions: Stick to .jar only for clarity"
echo "6. Hidden files: Decide whether to include .*.jar files"
echo "7. Special characters: Use proper escaping and UTF-8 handling"

# Cleanup
rm -rf "$TEST_DIR"