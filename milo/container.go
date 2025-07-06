package milo

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/opencontainers/runtime-spec/specs-go"
)

// GenerateCrunCommand generates the command arguments for running a container with crun
func GenerateCrunCommand(bundlePath, containerID string) []string {
	return []string{"crun", "run", "--bundle", bundlePath, containerID}
}

// CreateOCISpec creates an OCI runtime specification for executing a JAR file
func CreateOCISpec(javaHome, jarPath, taskDir string) (*specs.Spec, error) {
	spec := &specs.Spec{
		Version: "1.0.0",
		Process: &specs.Process{
			Terminal: false,
			Args:     []string{"/usr/lib/jvm/java/bin/java", "-jar", jarPath},
			Env:      []string{"PATH=/usr/lib/jvm/java/bin:/usr/bin:/bin", "JAVA_HOME=/usr/lib/jvm/java"},
			Cwd:      "/app",
		},
		Root: &specs.Root{
			Path:     "rootfs",
			Readonly: false,
		},
		Linux: &specs.Linux{
			Namespaces: []specs.LinuxNamespace{
				{Type: "pid"},
				{Type: "ipc"},
				{Type: "uts"},
				{Type: "mount"},
			},
		},
		Mounts: []specs.Mount{
			{
				Destination: "/proc",
				Type:        "proc",
				Source:      "proc",
			},
			{
				Destination: "/dev",
				Type:        "tmpfs",
				Source:      "tmpfs",
				Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
			},
			{
				Destination: "/sys",
				Type:        "sysfs",
				Source:      "sysfs",
				Options:     []string{"nosuid", "noexec", "nodev", "ro"},
			},
			{
				Destination: "/usr/lib/jvm/java",
				Source:      javaHome,
				Type:        "bind",
				Options:     []string{"rbind", "ro"},
			},
			{
				Destination: "/app",
				Source:      taskDir,
				Type:        "bind",
				Options:     []string{"rbind"},
			},
			{
				Destination: "/lib",
				Source:      "/lib",
				Type:        "bind",
				Options:     []string{"rbind", "ro"},
			},
			{
				Destination: "/lib64",
				Source:      "/lib64",
				Type:        "bind",
				Options:     []string{"rbind", "ro"},
			},
			{
				Destination: "/etc",
				Source:      "/etc",
				Type:        "bind",
				Options:     []string{"rbind", "ro"},
			},
		},
	}

	return spec, nil
}

// CreateContainerBundle creates a container bundle directory with config.json and rootfs
func CreateContainerBundle(bundlePath string, spec *specs.Spec) error {
	// Create the bundle directory
	if err := os.MkdirAll(bundlePath, 0755); err != nil {
		return err
	}

	// Create the rootfs directory and basic filesystem structure
	rootfsPath := filepath.Join(bundlePath, "rootfs")
	if err := os.MkdirAll(rootfsPath, 0755); err != nil {
		return err
	}

	// Create essential directories in rootfs
	essentialDirs := []string{
		"bin", "usr/bin", "usr/lib", "usr/lib/jvm", "app", "tmp", "var", "etc", "lib", "lib64",
	}
	for _, dir := range essentialDirs {
		dirPath := filepath.Join(rootfsPath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}
	}

	// Write the OCI spec as config.json
	configPath := filepath.Join(bundlePath, "config.json")
	configData, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, configData, 0600)
}
