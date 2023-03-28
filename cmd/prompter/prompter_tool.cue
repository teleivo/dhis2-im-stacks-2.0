package foo

import (
        "tool/cli"
        "tool/exec"
        "tool/file"
)

city: "Amsterdam"

// Say hello!
command: prompter: {
            // save transcript to this file
            var: file: *"out.txt" | string @tag(file)

            ask: cli.Ask & {
                    prompt:   "What is your name?"
                    response: string
            }

            // starts after ask
            echo: exec.Run & {
                    cmd:    ["echo", "Hello", ask.response + "!"]
                    stdout: string // capture stdout
            }

            // starts after echo
            append: file.Append & {
                    filename: var.file
                    contents: echo.stdout
            }

            // also starts after echo
            print: cli.Print & {
                    text: echo.stdout
            }
}

