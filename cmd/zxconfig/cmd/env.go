package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Output the contenst of .zx.toml file as an env file",
	RunE:  RunEnv,
}

func RunEnv(cmd *cobra.Command, args []string) error {
	generic := make(map[string]interface{})
	bs, err := os.ReadFile(".zx.toml")
	if err != nil {
		return err
	}

	toml.Unmarshal(bs, &generic)
	walkMap("", generic)

	return nil
}

func walkMap(p string, ms map[string]interface{}) {
	for k, v := range ms {
		np := fmt.Sprintf("%s_%s", p, cleanName(k))

		switch nv := v.(type) {
		case map[string]interface{}:
			walkMap(np, nv)
		case []interface{}:
			fmt.Printf("# List key %q to env is not supported", np)
		default:
			fmt.Printf(`%s_%s="%s"`, p, cleanValue(v))
		}
	}
}

func cleanName(n string) string {
	ns := strings.FieldsFunc(n, func(r rune) bool {
		return (r >= 'A' && r <= 'Z') ||
			(r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') ||
			r == '_'
	})
	return strings.ToUpper(strings.Join(ns, "_"))
}

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
