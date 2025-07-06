# Spike 4: Error Message Design

## Error Message Template

```
Error: <What went wrong>
Expected: <What should be correct>
Got: <What was actually provided>
Suggestion: <How to fix it>
```

## Error Scenarios and Messages

### 1. Non-JAR Extension

```
Error: Artifact must be a JAR file
Expected: File with .jar extension
Got: script.py
Suggestion: Ensure your artifact points to a compiled Java JAR file
```

### 2. Missing Artifact File

```
Error: Artifact file not found
Expected: JAR file in task directory
Got: No JAR files found in local/
Suggestion: Check that your artifact was downloaded successfully
```

### 3. Multiple JAR Files

```
Error: Multiple JAR files found
Expected: Single JAR file to execute
Got: 3 JAR files (main.jar, lib1.jar, lib2.jar)
Suggestion: Use a single artifact or specify which JAR to execute
```

### 4. Corrupt JAR File

```
Error: Invalid JAR file format
Expected: Valid Java archive (ZIP format)
Got: Corrupted or incomplete file
Suggestion: Re-download the artifact or verify the source file
```

### 5. Empty JAR File

```
Error: JAR file is empty
Expected: JAR containing Java classes
Got: 0 byte file
Suggestion: Check the artifact source and download process
```

### 6. HTML Error Page

```
Error: Downloaded file is not a JAR
Expected: Java archive file
Got: HTML document (possibly error page)
Suggestion: Verify the artifact URL is correct and accessible
```

### 7. Permission Denied

```
Error: Cannot read JAR file
Expected: Readable JAR file
Got: Permission denied on app.jar
Suggestion: Check file permissions or contact your administrator
```

### 8. Symlink Security Issue

```
Error: JAR file is a symbolic link
Expected: Regular JAR file
Got: Symlink pointing outside task directory
Suggestion: Use direct file paths instead of symbolic links
```

## Implementation Guidelines

1. **Be Specific**: Include actual filenames and paths
2. **Be Actionable**: Always provide a suggestion
3. **Be Consistent**: Use the same format for all errors
4. **Be Clear**: Avoid technical jargon where possible
5. **Be Helpful**: Guide users to the solution

## Error Categories

1. **Validation Errors** - Caught before container creation
2. **Runtime Errors** - Discovered during execution
3. **Security Errors** - Policy violations
4. **System Errors** - Environmental issues