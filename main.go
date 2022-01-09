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
   /~~~~\    \  \         /
  | |~~\ |     )  ~------~\
 / |  | |   /     ____ /~~~)\
(_/   | | |     /    |    ( |
       | | |     \    /   __)/ \
       \  \ \      \/    /' \   \
         \  \|\        /   | |\___|
           \ |  \____/     | |
           /^~>  \        _/ <
          |  |         \       \
          |  | \        \        \
          -^-\  \       |        )
               \_______/^\______/

`
func main() {
    user, err := user.Current()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Hello %s!\n %s", user.Username, WELCOME_TEXT)
    repl.Start(os.Stdin, os.Stdout)
}
