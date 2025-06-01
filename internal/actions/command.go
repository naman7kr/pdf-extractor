package actions

import "fmt"

type Command interface {
	Execute() error
}

type Invoker struct {
	Command []Command
}
type Describable interface {
	Description() string
}

func (i *Invoker) ExecuteCommand() error {
	for _, cmd := range i.Command {
		err := cmd.Execute()
		if err != nil {
			return err // Return the error if any command fails
		}

		// Check if the command implements the Describable interface
		if describableCmd, ok := cmd.(Describable); ok {
			fmt.Printf("Command executed successfully: %s\n", describableCmd.Description())
		} else {
			fmt.Println("Command executed successfully. unknown description.")
		}
	}
	return nil // Return nil if all commands execute successfully
}
