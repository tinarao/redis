package main

import (
	"bytes"
	"fmt"
	"github.com/tidwall/resp"
	"io"
	"log"
)

const (
	CommandSET = "SET"
)

type Command interface {
}

type SetCommand struct {
	key, val string
}

func parseCmd(raw string) (Command, error) {
	rd := resp.NewReader(bytes.NewBufferString(raw))

	for {
		v, _, err := rd.ReadValue()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		var cmd Command
		if v.Type() == resp.Array {
			for _, value := range v.Array() {

				switch value.String() {
				case CommandSET:
					fmt.Printf("v.Arr len: %v\n", len(v.Array()))
					if len(v.Array()) != 3 {
						return nil, fmt.Errorf("invalid number of vars for SET cmd")
					}
					cmd = SetCommand{
						key: v.Array()[1].String(),
						val: v.Array()[2].String(),
					}
					return cmd, nil
				}
			}
		}
	}

	return nil, nil
}
