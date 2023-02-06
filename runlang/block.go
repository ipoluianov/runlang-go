package runlang

type Block struct {
	tp         string
	comment    string
	beginIndex int
}

func NewBlock(tp string, beginIndex int, comment string) *Block {
	var c Block
	c.tp = tp
	c.comment = comment
	c.beginIndex = beginIndex
	return &c
}
