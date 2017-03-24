package arguments

import (
        "os"
)

const serviceFlag string = "-service"

func ServiceCall() bool {
    for _, arg := range os.Args {
        if arg == serviceFlag {
                return true
        }
    }
    return false
}