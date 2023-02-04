package runlang

type Block struct {
	tp         string
	beginIndex int
}

func NewBlock(tp string, beginIndex int) *Block {
	var c Block
	c.tp = tp
	c.beginIndex = beginIndex
	return &c
}
