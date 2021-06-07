package config

import (
	"bytes"
	"fmt"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const (
	PUSH      = "push"
	PULL      = "pull"
	MECHANISM = "mechanism"
)

var (
	envVarRe      = regexp.MustCompile(`\$\{(\w+)\}|\$(\w+)`)
	envVarEscaper = strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
	)
)

type CBConfig struct {
	Tags map[string]string
}

func NewConfig() *CBConfig {
	c := &CBConfig{
		Tags: make(map[string]string),
	}
	return c
}

func (c *CBConfig) LoadConfig(path string) error {
	var err error
	if path == "" {
		if path, err = getDefaultConfigPath(); err != nil {
			return err
		}
	}
	data, err := loadConfig(path)

	if err != nil {
		return fmt.Errorf("Error loading config file %s: %w", path, err)
	}

	if err = c.LoadConfigData(data); err != nil {
		return fmt.Errorf("Error loading config file %s: %w", path, err)
	}
	return nil
}

// LoadConfigData loads TOML-formatted config data
func (c *CBConfig) LoadConfigData(data []byte) error {
	tbl, err := parseConfig(data)
	if err != nil {
		return fmt.Errorf("Error parsing data: %s", err)
	}

	// Parse tags tables first:
	for _, tableName := range []string{"tags", "global_tags"} {
		if val, ok := tbl.Fields[tableName]; ok {
			subTable, ok := val.(*ast.Table)
			if !ok {
				return fmt.Errorf("invalid configuration, bad table name %q", tableName)
			}
			if err = toml.UnmarshalTable(subTable, c.Tags); err != nil {
				return fmt.Errorf("error parsing table name %q: %w", tableName, err)
			}
		}
	}
	return nil
}

func parseConfig(contents []byte) (*ast.Table, error) {
	contents = trimBOM(contents)

	parameters := envVarRe.FindAllSubmatch(contents, -1)
	for _, parameter := range parameters {
		if len(parameter) != 3 {
			continue
		}

		var env_var []byte
		if parameter[1] != nil {
			env_var = parameter[1]
		} else if parameter[2] != nil {
			env_var = parameter[2]
		} else {
			continue
		}

		env_val, ok := os.LookupEnv(strings.TrimPrefix(string(env_var), "$"))
		if ok {
			env_val = escapeEnv(env_val)
			contents = bytes.Replace(contents, parameter[0], []byte(env_val), 1)
		}
	}

	return toml.Parse(contents)
}

func getDefaultConfigPath() (string, error) {
	envfile := os.Getenv("TELEGRAF_CONFIG_PATH")
	homefile := os.ExpandEnv("${HOME}/.telegraf/telegraf.conf")
	etcfile := "/etc/telegraf/telegraf.conf"
	for _, path := range []string{envfile, homefile, etcfile} {
		if _, err := os.Stat(path); err == nil {
			log.Printf("I! Using config file: %s", path)
			return path, nil
		}
	}

	// if we got here, we didn't find a file in a default location
	return "", fmt.Errorf("No config file specified, and could not find one"+
		" in $TELEGRAF_CONFIG_PATH, %s, or %s", homefile, etcfile)
}

func loadConfig(config string) ([]byte, error) {
	_, err := url.Parse(config)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(config)

}

func trimBOM(f []byte) []byte {
	return bytes.TrimPrefix(f, []byte("\xef\xbb\xbf"))
}

func escapeEnv(value string) string {
	return envVarEscaper.Replace(value)
}
