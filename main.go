package main

import (
	"log"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]

	if len(args) != 1 {
		log.Fatalln("Supply only one argument, the name of the hack assembler (.asm) file to be assembled.")
	}

	asmFileName := args[0]

	asmFile, err := os.Open(asmFileName)
	if err != nil {
		log.Fatalf("Unable to open %s: %w", asmFileName, err)
	}

	defer asmFile.Close()

	hacks := strings.SplitN(Reverse(asmFileName), ".", 2)
	hackFileName := Reverse(hacks[len(hacks) - 1]) + ".hack"

	hackFile, err := os.Create(hackFileName)
	if err != nil {
		log.Fatalf("Unable to create %s: %w", hackFileName, err)
	}

	defer hackFile.Close()

	assemble(asmFile, hackFile)

}
