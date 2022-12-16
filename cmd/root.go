package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ianmuchina/tcb-dl/lib"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tcb-dl",
	Short: "Download Manga from tcbscans",
	Long:  "",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {
	// 	lib.Opt(lib.LoadLocalData())
	// },
}

var downloadCmd = &cobra.Command{
	Use:     "download",
	Aliases: []string{"dl"},
	Short:   "Download Chapter",
	Run: func(cmd *cobra.Command, args []string) {
		lib.FindChapter(Project, float64(Chapter))
	},
}

var downloadLatestCmd = &cobra.Command{
	Use:     "latest",
	Aliases: []string{"latest"},
	Short:   "Download Latest Chapter",
	Run: func(cmd *cobra.Command, args []string) {
		ch := lib.GetLatestChapter(Project)
		prj := lib.ProjectsMap[Project]
		s := strings.Split(prj, "/")

		fmt.Println(ch.Title, s[3])
		// lib.DownloadChapter(s[1], ch)
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync chapter data",
	Run: func(cmd *cobra.Command, args []string) {
		lib.SyncNew()
	},
}

var syncAll = &cobra.Command{
	Use:   "all",
	Short: "Refresh all chapter data",
	Run: func(cmd *cobra.Command, args []string) {
		lib.SyncAll()
		fmt.Println("Done")
	},
}

var syncNew = &cobra.Command{
	Use:   "new",
	Short: "Fetch new Chapter data",
	Run: func(cmd *cobra.Command, args []string) {
		lib.SyncNew()
		fmt.Println("Done")
	},
}

// Flags
var Chapter float64
var Project int

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tcb-dl.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(syncCmd)

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Download
	downloadCmd.Flags().Float64VarP(&Chapter, "chapter", "c", 0, "")
	downloadCmd.Flags().IntVarP(&Project, "project", "p", 5, "")

	downloadCmd.MarkFlagRequired("project")
	downloadCmd.MarkFlagRequired("chapter")

	downloadLatestCmd.Flags().IntVarP(&Project, "project", "p", 5, "")
	downloadLatestCmd.MarkFlagRequired("project")

	downloadCmd.AddCommand(downloadLatestCmd)

	syncCmd.AddCommand(syncAll, syncNew)
}
