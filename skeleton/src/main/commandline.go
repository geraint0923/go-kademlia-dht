package main

import (
	"fmt"
	"strings"
)

type CommandHandlerFunc func(cmd string, args []string, data interface{}) error

type CommandHandler struct {
	CommandName string
	CommandFunc CommandHandlerFunc
}

var CommandArray = [...]CommandHandler{
	CommandHandler{
		CommandName: "test",
		CommandFunc: TestHehe,
	},
	CommandHandler{
		CommandName: "test1",
		CommandFunc: TestHehe,
	},
	CommandHandler{
		CommandName: "whoami",
		CommandFunc: CmdDoWhoami,
	},
	CommandHandler{
		CommandName: "local_find_value",
		CommandFunc: CmdDoLocalFindValue,
	},
	CommandHandler{
		CommandName: "get_contact",
		CommandFunc: CmdDoGetContact,
	},
	CommandHandler{
		CommandName: "iterativeStore",
		CommandFunc: CmdDoIterativeStore,
	},
	CommandHandler{
		CommandName: "iterativeFindNode",
		CommandFunc: CmdDoIterativeFindNode,
	},
	CommandHandler{
		CommandName: "iterativeFindValue",
		CommandFunc: CmdDoIterativeFindValue,
	},
	CommandHandler{
		CommandName: "ping",
		CommandFunc: CmdDoPing,
	},
	CommandHandler{
		CommandName: "store",
		CommandFunc: CmdDoStore,
	},
	CommandHandler{
		CommandName: "find_node",
		CommandFunc: CmdDoFindNode,
	},
	CommandHandler{
		CommandName: "find_value",
		CommandFunc: CmdDoFindValue,
	},
}

func ExecuteCommand(line string, data interface{}) {
	splitRes := strings.Split(line, " ")
	var cmdFound *CommandHandler = nil
	cmd := ""
	args := []string{}
	// set command name
	if len(splitRes) > 0 {
		cmd = strings.TrimSpace(splitRes[0])
	}
	for _, val := range CommandArray {
		if val.CommandName == cmd {
			cmdFound = &val
			break
		}
	}
	if cmdFound != nil {
		err := cmdFound.CommandFunc(cmd, args, data)
		if err != nil {
			fmt.Println("Error: " + err.Error())
		}
	} else {
		fmt.Println("Command not found: " + cmd)
	}
}
