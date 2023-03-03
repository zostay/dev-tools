package goalsImpl

import (
	"context"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/zostay/dev-tools/zxpm/plugin"
	"github.com/zostay/dev-tools/zxpm/storage"
)

type OutputFormatter func(io.Writer, storage.KV) error

const InfoOutputFormatKey = "info.outputFormat"

var outputFormats = map[string]OutputFormatter{
	"properties": WriteOutProperties,
	"yaml":       WriteOutYaml,
}

var DefaultInfoOutputFormatter = WriteOutProperties

func InfoOutputFormatter(ctx context.Context) OutputFormatter {
	format := plugin.GetString(ctx, InfoOutputFormatKey)
	formatter := outputFormats[format]
	if formatter != nil {
		return formatter
	}
	return DefaultInfoOutputFormatter
}

func WriteOutProperties(w io.Writer, values storage.KV) error {
	for _, key := range values.AllKeys() {
		// TODO is key.subkey.subsubkey....=value the best output format?
		_, err := fmt.Fprintf(w, "%s = %#v\n", key, values.Get(key))
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteOutYaml(w io.Writer, values storage.KV) error {
	enc := yaml.NewEncoder(w)
	return enc.Encode(values.AllSettings())
}
