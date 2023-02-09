package runlang

type BlockType int

const (
	BlockTypeIf    = BlockType(0)
	BlockTypeWhile = BlockType(1)
	BlockTypeFn    = BlockType(2)
)

type Block struct {
	tp         BlockType
	comment    string
	beginIndex int
}

func NewBlock(tp BlockType, beginIndex int, comment string) *Block {
	var c Block
	c.tp = tp
	c.comment = comment
	c.beginIndex = beginIndex
	return &c
}
