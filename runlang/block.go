package runlang

type BlockType int

const (
	BlockTypeIf    = BlockType(0)
	BlockTypeWhile = BlockType(1)
	BlockTypeFn    = BlockType(2)
)

type Block struct {
	tp         BlockType
	beginIndex int
}

func NewBlock(tp BlockType, beginIndex int) *Block {
	var c Block
	c.tp = tp
	c.beginIndex = beginIndex
	return &c
}
