package util

import (
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// PrintConfig will takes a pointer to a config object, marshalls it to YAML and prints the result to the provided writer
func PrintConfig(w io.Writer, config interface{}) error {
	lc, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "---\n# Config\n%s\n\n", string(lc))
	return nil
}

type StringSliceFlag []string

func (f *StringSliceFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *StringSliceFlag) Set(value string) error {
	*f = strings.Split(value, ",")
	return nil
}
