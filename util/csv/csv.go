package csv

import (
	"bytes"
	"encoding/csv"
	"github.com/injoyai/conv"
)

func Export(data [][]interface{}) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	if _, err := buf.WriteString("\xEF\xBB\xBF"); err != nil {
		return nil, err
	}
	w := csv.NewWriter(buf)
	for _, v := range data {
		if err := w.Write(conv.Strings(v)); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf, nil
}
