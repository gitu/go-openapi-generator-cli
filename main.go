package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("you need to specify version of the openapi-code-generator-cli to use or use latest to fetch most current one")
	}
	version := os.Args[1]

	if version == "latest" {
		version = getLatestVersion()
	}

	cacheDir, _ := os.UserCacheDir()
	file := cacheDir + "/.openapi-codegenerator-cli/" + filename(version)
	if !fileExists(file) {
		download(version, file)
	}
	run(file)
}

func run(file string) {
	args := []string{"-jar", file}
	if len(os.Args) > 2 {
		args = append(args, os.Args[2:]...)
	}
	out, err := exec.Command("java", args...).Output()
	if err != nil {
		log.Fatal("Command finished with error: %v", err)
	}
	fmt.Print(string(out))
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

type Release struct {
	IsPrerelease bool                   `json:"prerelease"`
	TagName      string                 `json:"tag_name"`
	Others       map[string]interface{} `json:"-"` // Rest of the fields should go here.
}

func download(version, target string) {
	uri := "https://search.maven.org/remotecontent?filepath=org/openapitools/openapi-generator-cli/" + version + "/" + filename(version)

	if strings.HasSuffix(version, "-SNAPSHOT") {
		log.Fatal("snapshot versions are not supported")
	}

	resp, err := http.DefaultClient.Get(uri)
	if err != nil {
		log.Fatal("could not download release", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatal("bad status: ", resp.Status)
	}

	err = os.MkdirAll(filepath.Dir(target), 0644)
	if err != nil {
		log.Fatal("could not create parent directory", err)
	}

	out, err := os.Create(target)
	if err != nil {
		log.Fatal("could not create target", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal("error while downloading release", err)
	}
}

func filename(version string) string {
	return "openapi-generator-cli-" + version + ".jar"
}

func getLatestVersion() string {
	uri := "https://api.github.com/repos/openapitools/openapi-generator/releases"

	resp, err := http.DefaultClient.Get(uri)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal("could not fetch latest release", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatal("bad status: ", resp.Status)
	}

	releases := make([]Release, 50)

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		log.Fatal("could not parse latest release info", err)
	}

	for _, r := range releases {
		if !r.IsPrerelease {
			return r.TagName[1:]
		}
	}

	log.Fatal("no releases found")

	return ""
}
