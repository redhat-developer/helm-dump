package visitor

import (
	"fmt"
	"github.com/goccy/go-yaml/ast"
)

type MappingNodeVisitor struct {
	Collector   *Collector
	YamlPath    string
	path        string
	BeginOffset int
	EndOffset   int
}

func NewMappingNodeVisitor(path string, collector *Collector) *MappingNodeVisitor {
	return &MappingNodeVisitor{
		Collector: collector,
		YamlPath:  fmt.Sprintf("$%s", path),
		path:      path,
	}
}

func (v *MappingNodeVisitor) Visit(node ast.Node) ast.Visitor {
	if node.GetPath() != v.YamlPath {
		return v
	} else if n, ok := node.(*ast.MappingNode); !ok {
		return v
	} else if node.GetPath() != v.YamlPath {
		return v
	} else {
		v.Collector.AddPatch(v.path, n.GetToken().Position.Offset, n.GetToken().Next.Position.Offset)
		return v
	}
}
