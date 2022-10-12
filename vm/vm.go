package vm

import (
	"fmt"

	"github.com/Soj447/gonk/code"
	"github.com/Soj447/gonk/compiler"
	"github.com/Soj447/gonk/object"
)

const StackSize = 2048
const GlobalSize = 65536

type VM struct {
	constants    []object.Object
	instructions code.Instructions
	globals      []object.Object

	stack []object.Object
	sp    int
}

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

func New(bytecode *compiler.ByteCode) *VM {
	return &VM{
		constants:    bytecode.Constants,
		instructions: bytecode.Instructions,
		globals:      make([]object.Object, GlobalSize),
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparisonOperation(op)
			if err != nil {
				return err
			}
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpJump:
			jumpIndex := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip = jumpIndex - 1
		case code.OpJumpNotTruthy:
			jumpIndex := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				ip = jumpIndex - 1
			}
		case code.OpGetGlobal:
			globalIndex := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}

		case code.OpSetGlobal:
			globalIndex := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			vm.globals[globalIndex] = vm.pop()
		case code.OpArray:
			arrayLength := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			array := vm.buildArray(vm.sp-arrayLength, vm.sp)
			vm.sp = vm.sp - arrayLength
			vm.push(array)
		case code.OpHash:
			hashLength := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			hashMap, err := vm.buildHash(vm.sp-hashLength, vm.sp)
			if err != nil {
				return err
			}

			vm.sp = vm.sp - hashLength
			err = vm.push(hashMap)
			if err != nil {
				return err
			}
		case code.OpIndex:
			err := vm.executeIndexExpression()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (vm *VM) push(obj object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = obj
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	obj := vm.stack[vm.sp-1]
	vm.sp--
	return obj
}

func (vm *VM) buildArray(start, end int) *object.Array {
	elements := make([]object.Object, end-start)

	for i := start; i < end; i++ {
		elements[i-start] = vm.stack[i]
	}

	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(start, end int) (*object.Hash, error) {
	hashPairs := make(map[object.HashKey]object.HashPair)

	for i := start; i < end; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		hashPair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("key=%s is not hashable", key.Type())
		}

		hashPairs[hashKey.HashKey()] = hashPair
	}

	return &object.Hash{Pairs: hashPairs}, nil
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	rightObj := vm.pop()
	leftObj := vm.pop()

	if leftObj.Type() == object.INTEGER_OBJ && rightObj.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerBinaryOperation(op, leftObj, rightObj)
	}
	if leftObj.Type() == object.STRING_OBJ && rightObj.Type() == object.STRING_OBJ {
		return vm.executeStringBinaryOperation(op, leftObj, rightObj)
	}

	return fmt.Errorf("unsupported types for binary operation: left=%s, right=%s", leftObj.Type(), rightObj.Type())
}

func (vm *VM) executeIntegerBinaryOperation(op code.Opcode, leftObj, rightObj object.Object) error {
	leftVal := leftObj.(*object.Integer).Value
	rightVal := rightObj.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = leftVal + rightVal
	case code.OpSub:
		result = leftVal - rightVal
	case code.OpMul:
		result = leftVal * rightVal
	case code.OpDiv:
		result = leftVal / rightVal
	default:
		return fmt.Errorf("unsupported integer operation: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeStringBinaryOperation(op code.Opcode, leftObj, rightObj object.Object) error {
	leftVal := leftObj.(*object.String).Value
	rightVal := rightObj.(*object.String).Value

	if op != code.OpAdd {
		return fmt.Errorf("unkown string operation: %d", op)
	}

	return vm.push(&object.String{Value: leftVal + rightVal})
}

func (vm *VM) executeComparisonOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if right.Type() == object.INTEGER_OBJ && left.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparisonOperation(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(left == right))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(left != right))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)",
			op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparisonOperation(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal == rightVal))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal != rightVal))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftVal > rightVal))
	default:
		return fmt.Errorf("unkown operator: %d", op)
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for minus operation: %s", operand.Type())
	}

	val := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -val})
}

func (vm *VM) executeIndexExpression() error {
	index := vm.pop()
	left := vm.pop()

	if left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ {

		return vm.executeArrayIndex(left, index)
	} else if left.Type() == object.HASH_OBJ {
		return vm.executeHashIndex(left, index)
	}
	return fmt.Errorf("index is not supported for object with type=%s", left.Type())
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	indexObject := index.(*object.Integer)

	if indexObject.Value < 0 || indexObject.Value > int64(len(arrayObject.Elements)-1) {
		return vm.push(Null)
	}

	return vm.push(arrayObject.Elements[indexObject.Value])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func nativeBoolToBooleanObject(native bool) *object.Boolean {
	if native {
		return True
	}
	return False
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}
