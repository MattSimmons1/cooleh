package main

import (
	"github.com/MattSimmons1/cooleh/server"
	"github.com/MattSimmons1/cooleh/utils/ogre"
	"github.com/spf13/cobra"
	"sync"
)

func main() {

	if err := func() (rootCmd *cobra.Command) {

		var port string
		rootCmd = &cobra.Command{
			Use:   "cooleh",
			Short: "\u001B[1;38;2;116;132;116mcooleh\u001B[0m · ultra lightweight dev server · https://github.com/MattSimmons1/cooleh",
			Args:  cobra.ArbitraryArgs,
			Run: func(c *cobra.Command, args []string) {

				var wg sync.WaitGroup

				// we are going to wait for one goroutine to finish (but it never will)
				wg.Add(1)

				Ogre := ogre.New("o")

				Ogre.Growl()

				go server.Serve(port)

				wg.Wait()

				return
			},
		}
		rootCmd.Flags().StringVarP(&port, "port", "p", "5000", "change port")

		return
	}().Execute(); err != nil {
		panic(err)
	}
}
