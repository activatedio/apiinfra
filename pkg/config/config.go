// Package config wires runtime configuration via activatedio/cs.
//
// NewConfig loads YAML/JSON files (from CONFIG_PATHS env and/or
// explicit paths), then layers environment-variable overrides on
// top via a late-binding source. Downstream consumers root their
// own typed config under whatever cs prefix they choose; for the
// runtime in pkg/gateway, that prefix is PrefixServer ("server").
package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/activatedio/cs"
	"github.com/activatedio/cs/sources"
	"github.com/activatedio/cs/sources/json"
	"github.com/activatedio/cs/sources/yaml"
	"github.com/rs/zerolog/log"
)

const (
	// KeyConfigPaths is the env var used to specify additional
	// comma-separated config paths at startup.
	KeyConfigPaths = "CONFIG_PATHS"
	// PrefixServer is the cs key under which gateway.ServerConfig
	// is rooted by convention.
	PrefixServer = "server"

	suffixYAML = ".yaml"
	suffixYML  = ".yml"
	suffixJSON = ".json"
)

// NewConfig assembles a cs.Config tree from the supplied file paths
// plus any paths in CONFIG_PATHS. YAML and JSON files are
// recognized; other extensions are ignored. The environment is
// added as a late-binding source so env vars override file values.
func NewConfig(paths ...string) cs.Config {
	c := cs.New()

	var all []string
	for _, p := range strings.Split(os.Getenv(KeyConfigPaths), ",") {
		if p = strings.TrimSpace(p); p != "" {
			all = append(all, p)
		}
	}
	all = append(all, paths...)

	for _, p := range all {
		switch filepath.Ext(p) {
		case suffixYAML, suffixYML:
			c.AddSource(yaml.FromPath(p, "", yaml.ExpandEnv()))
		case suffixJSON:
			c.AddSource(json.FromPath(p, "", json.ExpandEnv()))
		default:
			log.Debug().Str("path", p).Msg("config: ignoring file (unrecognized extension)")
		}
	}

	c.AddLateBindingSource(sources.FromEnvironment(""))
	return c
}
