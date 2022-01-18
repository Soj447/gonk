package main

import (
	"fmt"
	"os"
	"os/user"

	"github.com/Soj447/gonk/repl"
)

const WELCOME_TEXT = `
                __------__
              /~          ~\
             |    //^\\//^\|
           /~~\  ||  o| |o|:~\
          | |6   ||___|_|_||:|
           \__.  /      o  \/'
            |   (       O   )
              \  \         /
               )  ~------~\
            Hello %s!

`

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf(WELCOME_TEXT, user.Username)
	repl.Start(os.Stdin, os.Stdout)
}
