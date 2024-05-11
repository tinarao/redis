package main

import (
	"bytes"
	"fmt"
	"github.com/tidwall/resp"
	"io"
	"log/slog"
)

const (
	CommandSET   = "SET"
	CommandHello = "hello"
)

type Command interface {
}

type SetCommand struct {
	key, val string
}

func parseCmd(msg string) (Command, error) {
	rd := resp.NewReader(bytes.NewBufferString(msg))
	for {
		v, _, err := rd.ReadValue()
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, err
		}
		if err != nil {
			slog.Error("err while reading a msg", "err", err.Error())
			return nil, fmt.Errorf("invalid or unknown command recieved")
		}

		//fmt.Printf("Read %s\n", v.Type())
		if v.Type() == resp.Array {
			for _, value := range v.Array() {
				switch value.String() {
				case CommandSET:
					fmt.Printf("%v\n", len(v.Array()))
					if len(v.Array()) != 3 {
						return nil, fmt.Errorf("invalid number of args for set cmd")
					}
					cmd := SetCommand{
						key: v.Array()[1].String(),
						val: v.Array()[2].String(),
					}
					fmt.Printf("%+v\n", cmd)
					return cmd, nil
				default:
					return nil, fmt.Errorf("invalid or unknown command recieved: %s", msg)
				}
			}
			return nil, fmt.Errorf("invalid or unknown command recieved: %s", msg)
		}
		return nil, fmt.Errorf("invalid or unknown command recieved: %s", msg)
	}
}
