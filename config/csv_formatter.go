package config

import (
	"encoding/csv"
	"io"
	"strings"
)

type CsvFormatter struct {
	writer     *csv.Writer
	needHeader bool
}

func NewCsvFormatter(writer io.Writer, needHeader bool) *CsvFormatter {
	return &CsvFormatter{csv.NewWriter(writer), needHeader}
}

func (w *CsvFormatter) WriteParam(envName, value string, descriptions []string) (err error) {
	defer w.writer.Flush()
	write := func(env, value, description string) {
		if err == nil {
			err = w.writer.Write([]string{env, value, description})
		}
	}

	if w.needHeader {
		write("Environment variable", "Value", "Description")
		w.needHeader = false
	}
	write(envName, value, strings.Join(descriptions, ","))

	return err
}
