package attributecheckercommand

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/attributechecker"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/usage"
	"github.com/urfave/cli"
)

// AttributecheckerCommand checks if all Capabilities/attributes are available to run the task on a specified Cluster or on a given Container Instance specified.
func AttributecheckerCommand() cli.Command {
	return cli.Command{
		Name:         "check-attributes",
		Usage:        usage.Attributechecker,
		Flags:        append(flags.OptionalConfigFlags(), attributecheckerFlags()...),
		Action:       attributechecker.AttributeChecker,
		OnUsageError: flags.UsageErrorFactory("attribute-checker"),
	}
}

func attributecheckerFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  flags.TaskDefinitionFlag,
			Usage: "Specifies the name or full Amazon Resource Name (ARN) of the ECS Task Definition. This is required to gather attributes of a Task Definition.",
		},
		cli.StringFlag{
			Name:  flags.ContainerInstancesFlag,
			Usage: "A list of container instance IDs or full ARN entries to check if all required attributes are available for the Task Definition to RunTask.",
		},
	}
}
