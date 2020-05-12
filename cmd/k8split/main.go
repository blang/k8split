package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// ManifestHead defines what the structure of the head of a manifest file
type ManifestHead struct {
	Version  string `yaml:"apiVersion" json:"apiVersion"`
	Kind     string `yaml:"kind" json:"kind"`
	Metadata *struct {
		Name      string `yaml:"name" json:"name"`
		Namespace string `yaml:"namespace" json:"namespace"`
	} `yaml:"metadata"`
}

var sep = regexp.MustCompile("(?:^|\\s*\n)---\\s*")

// SplitManifests takes a string of manifest and returns a map contains individual manifests
func SplitManifests(bigFile string) []string {
	var res []string
	// Making sure that any extra whitespace in YAML stream doesn't interfere in splitting documents correctly.
	bigFileTmp := strings.TrimSpace(bigFile)
	docs := sep.Split(bigFileTmp, -1)
	for _, d := range docs {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}

		res = append(res, d)
	}
	return res
}

var version = "k8split v1.1.0"

func printVersion() {
	fmt.Println(version)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [- | file]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Get version: %s version\n", os.Args[0])
		os.Exit(2)
	}
	var b []byte
	var err error
	arg := strings.TrimSpace(os.Args[1])
	switch {
	case arg == "version":
		printVersion()
		os.Exit(0)
	case arg != "-":
		b, err = ioutil.ReadFile(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: ", err)
			os.Exit(1)
		}
	default:
		b, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: ", err)
			os.Exit(1)
		}
	}

	m := SplitManifests(string(b))

	targetDir := os.Getenv("K8SPLIT_TARGET_DIR")
	if targetDir == "" {
		suffix := os.Getenv("K8SPLIT_DIR_SUFFIX")
		targetDir, err = ioutil.TempDir("", "k8split"+suffix)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating temp dir: ", err)
			os.Exit(1)
		}
	}

	for _, v := range m {
		var entry ManifestHead
		if err := yaml.Unmarshal([]byte(v), &entry); err != nil {
			fmt.Fprintf(os.Stderr, "YAML parse error on %s: %s", string(v), err)
			os.Exit(1)
		}
		if entry.Kind == "" {
			continue
		}
		var name string
		if entry.Metadata != nil {
			name = entry.Metadata.Name
		} else {
			fmt.Fprintf(os.Stderr, "Empty file of kind %s: %s", entry.Kind, v)
			continue
		}
		if err := writeFile(entry, v, targetDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file %s: %s\n", entry.Metadata.Name, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, " - %s %s %s (%s)\n", entry.Version, entry.Kind, name, entry.Metadata.Namespace)
	}
	fmt.Println(targetDir)
}

func writeFile(entry ManifestHead, content string, tmpDir string) error {
	namespace := "no-namespace"
	if entry.Metadata.Namespace != "" {
		namespace = entry.Metadata.Namespace
	}

	apiVersion := "no-version"
	if entry.Version != "" {
		apiVersion = entry.Version
	}

	kind := "no-kind"
	if entry.Kind != "" {
		kind = entry.Kind
	}

	name := entry.Metadata.Name
	path := filepath.Join(tmpDir, namespace, apiVersion, kind)
	if err := os.MkdirAll(path, 0700); err != nil {
		return errors.Wrapf(err, "Could not create path: %s", path)
	}
	filename := filepath.Join(path, name+".yml")
	err := ioutil.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return errors.Wrapf(err, "Could not write file: %s", filename)
	}
	return nil
}
