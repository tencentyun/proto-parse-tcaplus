package main

import (
	"fmt"
	"os"

	"github.com/tencentyun/proto-parse-tcaplus/comm"

	"github.com/spf13/cobra"
	"github.com/tencentyun/proto-parse-tcaplus/tools"
)

func parseArgs() {
	var protoSrcPath, protoDstPath string
	var cfgFile string
	var rootCmd = &cobra.Command{
		Use:     "proto-parse-tcaplus",
		Short:   "Parse business proto files and write to new proto files for TcaplusDB",
		Long:    "Parse business proto files and write to new proto files for TcaplusDB ",
		Example: `  ./proto-parse-tcaplus -s "./testdata/test" -d "./out/test"  -c "./config/proto_parse.cfg`,
		Run: func(cmd *cobra.Command, args []string) {

			if protoSrcPath == "" || protoDstPath == "" || cfgFile == "" {
				cmd.Help()
				os.Exit(0)
			}
			//check dest path is existed or not, if not create.
			if err := tools.CreateDir(protoDstPath); err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
			//read config file
			cfg, err := tools.ReadIni(cfgFile)
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
			//parse config file
			err = tools.ParseCfg(cfg)
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
			ProtoParseAndWrite(protoSrcPath, protoDstPath, comm.IgnoreProtoFiles)
		},
	}

	rootCmd.Flags().StringVarP(&protoSrcPath, "source-path", "s", "", "source path of proto files")
	rootCmd.Flags().StringVarP(&protoDstPath, "dest-path", "d", "", "destination path of generated proto files")
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "tool config file")
	rootCmd.Execute()

}
func main() {
	parseArgs()
}
