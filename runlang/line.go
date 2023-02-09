package runlang

type Instruction int

const (
	InstructionSet    = Instruction(0)
	InstructionWhile  = Instruction(1)
	InstructionIf     = Instruction(2)
	InstructionReturn = Instruction(3)
	InstructionFn     = Instruction(4)
	InstructionBreak  = Instruction(5)
	InstructionEnd    = Instruction(6)
	InstructionDump   = Instruction(7)
)

type Line struct {
	Tp     string
	Lexems []string
	Instruction

	// If & While
	Condition []string

	// Set
	LeftPart             []string
	RightPart            []string
	RightPartIsFunction  bool
	RightPartOperationFn string
}
