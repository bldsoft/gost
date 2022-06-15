package config

import (
	"io"
	"strings"
)

type MarkdownFormatter struct {
	writer     io.Writer
	needHeader bool
}

func NewMarkdownFormatter(writer io.Writer, needHeader bool) *MarkdownFormatter {
	return &MarkdownFormatter{writer, needHeader}
}

func (w *MarkdownFormatter) WriteParam(envName, value string, descriptions []string) (err error) {
	write := func(s string) {
		if err == nil {
			_, err = w.writer.Write([]byte(s))
		}
	}

	if w.needHeader {
		write("|**Environment variable**|**Value**|**Description**|\n")
		write("|------------------------|---------|---------------|\n")
		w.needHeader = false
	}

	write("|")
	write(envName)
	write("|")
	write(value)
	write("|")
	write(strings.Join(descriptions, "<br/>"))

	write("|\n")

	return err
}
