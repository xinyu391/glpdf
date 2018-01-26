package glpdf

import (
	"fmt"
)

/**
BT BeginText
ET EndText
Td: Move text position.
TD: Move text position and set leading.
T*: Move to start of next text line.
Tc: Set character spacing.
Tf: Set text font and size.
Tz: Set horizontal text scaling.
TL: Set text leading.
Tr: Set text rendering mode.
Ts: Set text rise.
Tw: Set word spacing.
Tj: Show text.
TJ: Show text, with position adjustments.
': Move to the next line and show text.  = T* + Tj
": Set word and character spacing, move to next line, and show text. Tw+Tc+'

/////State
cm Concatenate matrix to current transformation matrix.
Q: Restore the graphics state.
q: Save the graphics state.
i: Set the flatness tolerance.
gs: Set parameters from graphics state parameter dictionary.
 J: Set the line cap style.
d: Set the line dash pattern.
j: Set the line join style.
M: Set miter limit.
 w: Set line width.
Tm: Set text matrix and text line matrix.
 ri: Set the rendering intent.

//// color
sc SetNonStrokingColor
scn SetNonStrokingColorN
cs SetNonStrokingColorSpace
k  SetNonStrokingDeviceCMYKColor
g  SetNonStrokingDeviceGrayColor
rg SetNonStrokingDeviceRGBColor
SC SetStrokingColor
SCN SetStrokingColorN
CS SetStrokingColorSpace
K  SetStrokingDeviceCMYKColor
G  SetStrokingDeviceGrayColor
RG SetStrokingDeviceRGBColor
*/
type Operator interface {
	Name() string
	Process()
	Args() []DataType
}

type BaseOp struct {
	name string
	args []DataType
}

func NewOp(name string, args []DataType) (op *BaseOp) {
	op = &BaseOp{name, args}
	return
}
func (op *BaseOp) Name() string {
	return op.name
}
func (op *BaseOp) Args() []DataType {
	return op.args
}
func (op *BaseOp) Process() {

}
func (op *BaseOp) String() string {
	return fmt.Sprint("{", op.name, " : ", op.args, "}")

}

func (op *BaseOp) addArg(arg DataType) {
	op.args = append(op.args, arg)
}
func (op *BaseOp) setName(str string) {
	op.name = str
}
