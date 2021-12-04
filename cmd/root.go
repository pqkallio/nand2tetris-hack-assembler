package cmd

import (
	"log"
	"os"
	"strings"

	"github.com/pqkallio/nand2tetris-hack-assembler/assembler"
	"github.com/spf13/cobra"
)

var (
	input, output string
	format        string

	rootCmd = &cobra.Command{
		Use:   "compile",
		Short: "Hack machine language compiler",
		Long:  "Compile a Hack assembly language file into Hack machine language ROM file",
		Run: func(cmd *cobra.Command, args []string) {
			if input == "" {
				panic("input file is required")
			}
			if output == "" {
				output = strings.TrimSuffix(input, ".asm") + ".hack"
			}

			f := assembler.Format(format)

			if !f.IsValid() {
				log.Fatalf("invalid format %s", format)
			}

			asmFile, err := os.Open(input)
			if err != nil {
				log.Fatalf("Unable to open %s: %s", input, err)
			}

			defer asmFile.Close()

			hackFile, err := os.Create(output)
			if err != nil {
				log.Fatalf("Unable to create %s: %s", output, err)
			}

			defer hackFile.Close()

			s := assembler.NewServer(f)

			s.Assemble(asmFile, hackFile)
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&input, "input", "i", "", "input file")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "output file, default <input>.hack")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "binary", "output format: binary | text")
}
