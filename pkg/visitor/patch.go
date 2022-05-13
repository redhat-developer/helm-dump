package visitor

import (
	"bytes"
	"fmt"
)

type Patch struct {
	Path        string
	BeginOffset int
	EndOffset   int
}

func (p Patch) Apply(valuesKey string, data []byte) []byte {
	fst := data[0 : p.BeginOffset+1]
	snd := data[p.EndOffset:]

	templateNewData := bytes.Join(
		[][]byte{
			fst,
			[]byte(fmt.Sprintf("{{ .Values.%s }}\n", valuesKey)),
			snd,
		},
		[]byte{})

	return templateNewData

}
