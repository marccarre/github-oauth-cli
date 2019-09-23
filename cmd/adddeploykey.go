package cmd

import (
	"context"

	gh "github.com/google/go-github/v28/github"
	"github.com/marccarre/github-oauth-cli/pkg/github"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// addDeployKeyCmd represents the add-deploy-key command
var addDeployKeyCmd = &cobra.Command{
	Use:   "add-deploy-key",
	Short: "Add a GitHub deploy key to the selected repository",
	Run:   addDeployKeyRun,
}

type addDeployKeyParams struct {
	repository string
}

var params addDeployKeyParams

func init() {
	rootCmd.AddCommand(addDeployKeyCmd)
	addDeployKeyCmd.Flags().StringVarP(&params.repository, "repository", "r", "", "GitHub repository to add the deploy key to")
}

func addDeployKeyRun(cmd *cobra.Command, args []string) {
	log.WithFields(log.Fields{"repository": params.repository}).Info("add-deploy-key called")
	ctx := context.Background()
	client, err := github.NewClient(ctx, github.Owner(params.repository))
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("failed to initialise GitHub client")
	}
	client.AddDeployKey(ctx, params.repository, &gh.Key{
		ID:       "flux",
		Title:    "flux",
		ReadOnly: false,
		Key:      "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCvsl/Sv2HJX/Wod0YbetsnHAVChHBv1u/QR0tuDSVptXCV5NPklkMztn4F9/0BD9mqC41F32iglX8CFBUi6OEApmtY7+ixgDj/KjCdj08HY90TJoS77pL+bvsszwvoL8P8ET5d3IiYE+CglpS2qFxggb9jcMoWtlHHRGIME0EO7FeNyira3T48DhJOaAloUsGyOCtWswrWvAPkSnfKUXzPgoChNPsJIuHUgMEySPYIm7gjGPeIywMcdu5dcM/E6W7utfBTNo9yzcJG48vAquq9hiOvyi+aoa01QNJMtIkhn33FXyjPHSQv83QveXKS+RqQjc2chJMHBNbk3z1P7OrNtlzwlrRWwC+8w5/dam28Rk8L5ejdBqLmmLmR+/h4WEmh8R7jAEsOcy7Lc94zHVpguZo1Mq9jAMFo8CnH8raxtheGXbx6WG4l9B7HNOl2Y+Nx8w3H1sgCLvhiSNkNMXomO+ZRYZqe4XbLrmuvtQdxWMy2dkhCqI3K3OJ8Fs1+igvj3ZCYcj7VSR+PA0ZssdeIVoQbB6EZZrJEhWCYOlt8xXEgnudghpN1vY1ZBi/sY6JiHg1qKdgCLcTkMQ2qZ0UiDaEp/XUbNaBVq9+CT7M14jq9DjLCjsnzUv/mdWJA8mAWJLOd98UjRnNWMwmOxZrxCOwwuPW20GXlpGU7KWqbfw==",
	})
}
