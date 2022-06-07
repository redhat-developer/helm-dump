package visitor

type Collector struct {
	Patches []Patch
}

func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) AddPatch(path string, beginOffset int, endOffset int) {
	c.Patches = append(c.Patches, Patch{Path: path, BeginOffset: beginOffset, EndOffset: endOffset})
}
