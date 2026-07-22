package cli

import "io"

var version = "dev"

func runVersion(output io.Writer) error {
	_, err := io.WriteString(output, "lexicon version "+version+"\n")
	return err
}
