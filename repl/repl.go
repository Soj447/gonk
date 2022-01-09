package repl

import (
    "bufio"
    "fmt"
    "io"
    "github.com/Soj447/gonk/lexer"
    "github.com/Soj447/gonk/parser"
    "github.com/Soj447/gonk/evaluator"
)

const PS1 = "~>"

func Start(in io.Reader, out io.Writer) {
    scanner := bufio.NewScanner(in)

    for {
        fmt.Printf(PS1)
        scanned := scanner.Scan()
        if !scanned {
            return
        }

        line := scanner.Text()
        l := lexer.New(line)
        p := parser.New(l)

        program := p.ParseProgram()
        if len(p.Errors()) != 0 {
            printParserErrors(out, p.Errors())
            continue
        }

        evaluated := evaluator.Eval(program)
        if evaluated != nil {
            io.WriteString(out, evaluated.Inspect())
            io.WriteString(out, "\n")
        }
        

    }
}

func printParserErrors(out io.Writer, errors []string) {
    io.WriteString(out, "parser errors:\n")
    for _, msg := range errors {
        io.WriteString(out, "\t" + msg + "\n")
    }
}