package evaluator

import (
	"testing"

	"github.com/Soj447/gonk/lexer"
	"github.com/Soj447/gonk/object"
	"github.com/Soj447/gonk/parser"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"123 + 1", 124},
		{"20 / 10 + 3", 5},
		{"10 * (3 + 2)", 50},
		{"-50 * 2", -100},
		{"5 - 1", 4},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input     string
		exptected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"2 == 1", false},
		{"2 != 1", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.exptected)
	}
}

func TestEvalStringExpression(t *testing.T) {
	input := `"Hello World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}

}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input     string
		exptected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.exptected)
	}
}

func TestIfElseExpression(t *testing.T) {
	tests := []struct {
		input     string
		exptected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.exptected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input  string
		expect int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"9; return 10; 9;", 10},
		{"5; return 10 * 2; 10", 20},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expect)
	}

}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input         string
		expectedError string
	}{
		{"5 + true", "type mismatch: INTEGER + BOOLEAN"},
		{"5 + true; 5", "type mismatch: INTEGER + BOOLEAN"},
		{"-true", "unknown operator: -BOOLEAN"},
		{"true + false", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) { true + false; }", "unknown operator: BOOLEAN + BOOLEAN"},
		{
			`
            if (10 > 1) {
                if (10 > 1) {
                    return true + false;
                }
                return 1;
            }
            `,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{"foobar", "identifier not found: foobar"},
		{`"flower" - "gaze"`, "unknown operator: STRING - STRING"},
		{`{"name": "Monkey"}[fn(x) { x }];`, "type FUNCTION is not hashable"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error returned. got=%T(%+v)", evaluated, evaluated)
		}

		if errObj.Message != tt.expectedError {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expectedError, errObj.Message)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b +5; c;", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}

}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2; };"

	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v", fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
    let newAdder = fn(x) {
    fn(y) { x + y };
    };
    let addTwo = newAdder(2);
    addTwo(2);`
	testIntegerObject(t, testEval(input), 4)
}

func TestStringConcatenation(t *testing.T) {
	input := `"flower" + " "  + "gaze";`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "flower gaze" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("flower")`, 6},
		{`len("hell yeah")`, 9},
		{`len(1)`, "argument to `len` not supported. got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", evaluated, evaluated)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. exptected=%q, got=%q", expected, errObj.Message)
			}
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not Arrray. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("aarray has wrong num of elements. got=%d", len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			nil,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestHashLiterals(t *testing.T) {
	input := `let key = "key";
	{
		"one": 1,
		"tw" + "o": 1 + 1,
		"three": 9 / 3,
		false: 4,
		true: 5,
		6: 6,
		key: 14 - 7
	}`

	evaluated := testEval(input)
	hashmap, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("Eval didn't returh Hash. got=%T (%+v)", evaluated, evaluated)
	}

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		FALSE.HashKey():                          4,
		TRUE.HashKey():                           5,
		(&object.Integer{Value: 6}).HashKey():    6,
		(&object.String{Value: "key"}).HashKey(): 7,
	}

	if len(hashmap.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong num of pairs. got=%d", len(hashmap.Pairs))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := hashmap.Pairs[expectedKey]
		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}

		testIntegerObject(t, pair.Value, expectedValue)
	}
}

func TestHashIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`let key = "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
		{
			`{true: 5}[true]`,
			5,
		},
		{
			`{false: 5}[false]`,
			5,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.Object, exptected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != exptected {
		t.Errorf("object has wrong value. got=%d, exptected=%d", result.Value, exptected)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj object.Object, exptected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != exptected {
		t.Errorf("object has wrong value. got=%t, exptected=%t", result.Value, exptected)
		return false
	}

	return true
}
