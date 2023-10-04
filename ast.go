package main

type TermKind string

const (
	KindInt      TermKind = "Int"
	KindStr      TermKind = "Str"
	KindBool     TermKind = "Bool"
	KindBinary   TermKind = "Binary"
	KindCall     TermKind = "Call"
	KindFunction TermKind = "Function"
	KindLet      TermKind = "Let"
	KindIf       TermKind = "If"
	KindPrint    TermKind = "Print"
	KindFirst    TermKind = "First"
	KindSecond   TermKind = "Second"
	KindTuple    TermKind = "Tuple"
	KindVar      TermKind = "Var"
)

type BinaryOp string

const (
	Add BinaryOp = "Add"
	Sub BinaryOp = "Sub"
	Mul BinaryOp = "Mul"
	Div BinaryOp = "Div"
	Rem BinaryOp = "Rem"
	Eq  BinaryOp = "Eq"
	Neq BinaryOp = "Neq"
	Lt  BinaryOp = "Lt"
	Gt  BinaryOp = "Gt"
	Lte BinaryOp = "Lte"
	Gte BinaryOp = "Gte"
	And BinaryOp = "And"
	Or  BinaryOp = "Or"
)

type Term interface{}

type Int struct {
	Kind     TermKind `json:"kind"`
	Value    int32    `json:"value"`
	Location Location `json:"location"`
}

type Str struct {
	Kind     TermKind `json:"kind"`
	Value    string   `json:"value"`
	Location Location `json:"location"`
}

type Bool struct {
	Kind     TermKind `json:"kind"`
	Value    bool     `json:"value"`
	Location Location `json:"location"`
}

type Binary struct {
	Kind     TermKind `json:"kind"`
	LHS      Term     `json:"lhs"`
	Op       BinaryOp `json:"op"`
	RHS      Term     `json:"rhs"`
	Location Location `json:"location"`
}

type Print struct {
	Kind     TermKind `json:"kind"`
	Value    Term     `json:"value"`
	Location Location `json:"location"`
}

type First struct {
	Kind     TermKind `json:"kind"`
	Value    Term     `json:"value"`
	Location Location `json:"location"`
}

type Second struct {
	Kind     TermKind `json:"kind"`
	Value    Term     `json:"value"`
	Location Location `json:"location"`
}

type If struct {
	Kind      TermKind `json:"kind"`
	Condition Term     `json:"condition"`
	Then      Term     `json:"then"`
	Otherwise Term     `json:"otherwise"`
	Location  Location `json:"location"`
}

type Tuple struct {
	Kind     TermKind `json:"kind"`
	First    Term     `json:"first"`
	Second   Term     `json:"second"`
	Location Location `json:"location"`
}

type Parameter struct {
	Text     string   `json:"text"`
	Location Location `json:"location"`
}

type Call struct {
	Kind      TermKind `json:"kind"`
	Callee    Term     `json:"callee"`
	Arguments []Term   `json:"arguments"`
	Location  Location `json:"location"`
}

type Let struct {
	Kind     TermKind  `json:"kind"`
	Name     Parameter `json:"name"`
	Value    Term      `json:"value"`
	Next     Term      `json:"next"`
	Location Location  `json:"location"`
}

type Var struct {
	Kind     TermKind `json:"kind"`
	Text     string   `json:"text"`
	Location Location `json:"location"`
}

type Function struct {
	Kind       TermKind    `json:"kind"`
	Parameters []Parameter `json:"parameters"`
	Value      Term        `json:"value"`
	Location   Location    `json:"location"`
}

type File struct {
	Name       string   `json:"name"`
	Expression Term     `json:"expression"`
	Location   Location `json:"location"`
}

type Location struct {
	Start    uint32 `json:"start"`
	End      uint32 `json:"end"`
	Filename string `json:"filename"`
}

type ClosureValue struct {
	Body       Term
	Parameters Term
}

type Closure struct {
	Kind  string
	Value ClosureValue
}
