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

	return Eval(program)
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
