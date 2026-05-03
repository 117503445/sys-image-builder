package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	registryPort = "5123"
	registryUser = "testuser"
	registryPass = "testpass"
	networkName  = "sib-e2e-net"
	registryName = "sib-e2e-registry"
	alpineName   = "sib-e2e-alpine"
	imageRef     = registryName + ":5000/rootfs:latest"
)

func run(name string, args ...string) {
	log.Info().Str("cmd", name+" "+strings.Join(args, " ")).Msg("exec")
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Panic().Err(err).Str("cmd", name).Msg("command failed")
	}
}

func runOutput(name string, args ...string) string {
	log.Info().Str("cmd", name+" "+strings.Join(args, " ")).Msg("exec")
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		log.Panic().Err(err).Str("cmd", name).Str("output", string(out)).Msg("command failed")
	}
	return strings.TrimSpace(string(out))
}

func runIgnoreErr(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func cleanup() {
	log.Info().Msg("cleaning up e2e resources")
	runIgnoreErr("docker", "rm", "-f", registryName, alpineName)
	runIgnoreErr("docker", "network", "rm", networkName)
}

func waitForRegistry() {
	log.Info().Msg("waiting for registry to be ready")
	for i := 0; i < 30; i++ {
		// registry with auth returns 401 on /v2/, which still means it's up
		out, err := exec.Command("docker", "exec", registryName,
			"wget", "-q", "-O", "/dev/null", "-S", "http://localhost:5000/v2/").CombinedOutput()
		if err == nil || strings.Contains(string(out), "401") || strings.Contains(string(out), "200") {
			log.Info().Msg("registry is ready")
			return
		}
		time.Sleep(time.Second)
	}
	log.Panic().Msg("registry did not become ready in time")
}

func ensureInsecureRegistry() {
	daemonJSON := "/etc/docker/daemon.json"
	entry := "localhost:" + registryPort

	var config map[string]interface{}

	data, err := os.ReadFile(daemonJSON)
	if err == nil {
		_ = json.Unmarshal(data, &config)
	}
	if config == nil {
		config = make(map[string]interface{})
	}

	registries, _ := config["insecure-registries"].([]interface{})
	for _, r := range registries {
		if r == entry {
			return
		}
	}

	registries = append(registries, entry)
	config["insecure-registries"] = registries

	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Panic().Err(err).Msg("failed to marshal daemon.json")
	}

	if err := os.WriteFile(daemonJSON, newData, 0644); err != nil {
		log.Panic().Err(err).Msg("failed to write daemon.json")
	}

	log.Info().Str("entry", entry).Msg("added insecure registry, reloading docker")
	run("systemctl", "reload", "docker")
	time.Sleep(2 * time.Second)
}

func e2e() {
	cleanup()
	defer cleanup()

	binaryPath := filepath.Join(dirProjectRoot, "data", "e2e", "sys-image-builder")

	// Step 1: build static binary
	log.Info().Msg("[1/6] building static binary")
	os.MkdirAll(filepath.Dir(binaryPath), 0755)
	{
		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/cli")
		cmd.Dir = dirProjectRoot
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Panic().Err(err).Msg("failed to build binary")
		}
	}

	// Step 2: start registry with htpasswd auth
	log.Info().Msg("[2/6] starting registry with auth")
	run("docker", "network", "create", networkName)

	htpasswdDir := filepath.Join(dirProjectRoot, "data", "e2e", "auth")
	os.MkdirAll(htpasswdDir, 0755)
	htpasswdFile := filepath.Join(htpasswdDir, "htpasswd")

	htpasswdContent := runOutput("docker", "run", "--rm",
		"httpd:2", "htpasswd", "-Bbn", registryUser, registryPass)
	if err := os.WriteFile(htpasswdFile, []byte(htpasswdContent+"\n"), 0644); err != nil {
		log.Panic().Err(err).Msg("failed to write htpasswd")
	}

	run("docker", "run", "-d",
		"--name", registryName,
		"--network", networkName,
		"-p", registryPort+":5000",
		"-e", "REGISTRY_AUTH=htpasswd",
		"-e", "REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd",
		"-e", "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm",
		"-v", htpasswdDir+":/auth",
		"registry:2",
	)

	waitForRegistry()

	// Step 3: start alpine container with the binary
	log.Info().Msg("[3/6] starting alpine test container")
	run("docker", "run", "-d",
		"--name", alpineName,
		"--network", networkName,
		"alpine:latest", "sleep", "infinity",
	)
	run("docker", "cp", binaryPath, alpineName+":/usr/local/bin/sys-image-builder")

	// Step 4: create test file
	log.Info().Msg("[4/6] creating /root/test.txt in alpine")
	run("docker", "exec", alpineName,
		"sh", "-c", "mkdir -p /root && echo hello-e2e > /root/test.txt")

	// Step 5: push from inside alpine
	log.Info().Msg("[5/6] pushing filesystem from alpine to registry")
	run("docker", "exec", alpineName,
		"sys-image-builder", "push",
		"--image", imageRef,
		"--username", registryUser,
		"--password", registryPass,
		"--insecure",
		"--default-excludes",
	)

	// Step 6: pull and verify
	log.Info().Msg("[6/6] verifying pushed image")
	ensureInsecureRegistry()

	localRef := fmt.Sprintf("localhost:%s/rootfs:latest", registryPort)

	loginCmd := exec.Command("docker", "login", fmt.Sprintf("localhost:%s", registryPort),
		"-u", registryUser, "-p", registryPass)
	loginCmd.Stdout = os.Stdout
	loginCmd.Stderr = os.Stderr
	if err := loginCmd.Run(); err != nil {
		log.Panic().Err(err).Msg("docker login failed")
	}

	run("docker", "pull", localRef)

	result := runOutput("docker", "run", "--rm", localRef, "cat", "/root/test.txt")

	if result == "hello-e2e" {
		log.Info().Msg("E2E TEST PASSED")
	} else {
		log.Panic().Str("expected", "hello-e2e").Str("got", result).Msg("E2E TEST FAILED")
	}
}
