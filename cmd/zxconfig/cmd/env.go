package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/zostay/dev-tools/pkg/config"
)

const zxPrefix = "ZX_"

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Output the contents of ZX configs as an env file",
	RunE:  RunEnv,
}

// RunEnv reads in the .zx.yaml file and outputs all the configuration values
// found there as a environment file.
func RunEnv(cmd *cobra.Command, args []string) error {
	config.Init(verbosity)
	fmt.Println(`ZXCONFIG="If good works cannot gain you your salvation, how can bad works cause you to lose your salvation?"`)
	walkMap("", viper.AllSettings())

	return nil
}

// makePrefix converts a prefix and key value into a "prefix_key" string and
// returns it.
func makePrefix(p, k string) string {
	np := cleanName(k)
	if p != "" {
		np = fmt.Sprintf("%s_%s", p, np)
	}
	return np
}

// walkMap walks a key-value map in the configuration file and generates an
// environment value for every value found in it. Each key value file is
// rendered as a KEY="VALUE", properly escaped.
func walkMap(p string, ms map[string]interface{}) {
	keys := make([]string, len(ms))
	i := 0
	for k := range ms {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := ms[k]
		np := makePrefix(p, k)

		switch nv := v.(type) {
		case map[string]interface{}:
			walkMap(np, nv)
		case []interface{}:
			walkSlice(np, nv)
		default:
			cv := cleanValue(v)
			if isExpandablePathVar(np) {
				cv = expandPath(cv)
			}
			fmt.Printf(`%s%s="%s"`, zxPrefix, np, cv)
			fmt.Println("")
		}
	}
}

// walkSlice walks a list of values in the configuration file and generates
// environment value for every value found in it. These array values are output
// using array declarations.
func walkSlice(p string, vs []interface{}) {
	if len(vs) > 0 {
		fmt.Printf("declare -a %s\n", p)
		fmt.Printf("%s=(\n", p)
		for _, v := range vs {
			sv := fmt.Sprintf("%v", v)
			fmt.Printf("\t\"%s\"\n", cleanValue(sv))
		}
		fmt.Printf(")\n")
	}
}

// cleanName converts keys found in the configuration file to names appropriate
// for use in a env file.
func cleanName(n string) string {
	ns := strings.FieldsFunc(n, func(r rune) bool {
		return !((r >= 'A' && r <= 'Z') ||
			(r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') ||
			r == '_')
	})
	return strings.ToUpper(strings.Join(ns, "_"))
}

// cleanValue cleans up a value for embedding in an environment string with
// proper escaping to be read back by a shell program.
func cleanValue(v interface{}) string {
	sv := fmt.Sprintf("%v", v)
	cv := new(strings.Builder)
	for _, r := range sv {
		switch r {
		case '\n':
			fmt.Fprint(cv, `\n`)
		case '\r':
			fmt.Fprint(cv, `\r`)
		case 0:
			fmt.Fprint(cv, `\0`)
		case '$', '"', '\\', '`':
			fmt.Fprintf(cv, `\%c`, r)
		default:
			cv.WriteRune(r)
		}
	}
	return cv.String()
}

// isExpandablePathVar returns true if the given clean name ends with FILES or
// DIR.
func isExpandablePathVar(cn string) bool {
	return strings.HasSuffix(cn, "FILES") ||
		strings.HasSuffix(cn, "DIR")
}

// expandPath expands paths into absolute paths for tools, including turning ~
// prefix into a home directory. If the user's home directory cannot be
// determined, then there's no error, just ~ is not expanded. If there is an
// error turning the path into an absolute path, then the path is returned
// unexpanded.
func expandPath(p string) string {
	home, err := os.UserHomeDir()
	if err == nil {
		if p == "~" {
			p = home
		} else if strings.HasPrefix(p, "~/") {
			p = home + p[1:]
		}
	}

	if ap, err := filepath.Abs(p); err == nil {
		return ap
	} else {
		return p
	}
}
