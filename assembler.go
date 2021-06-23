package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	nextRamAddress uint16 = 15
	lineNumber uint16 = 0
)

func getNextRAMAddress() uint16 {
	nextRamAddress++

	return nextRamAddress
}

func assemble(in *os.File, out *os.File) {
	scanner := bufio.NewScanner(in)
	writer := bufio.NewWriter(out)

	// first pass:
	// resolve label addresses
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if len(line) == 0 || strings.HasPrefix(line, "/") {
			continue
		}

		if strings.HasPrefix(line, "(") {
			label := strings.Trim(line, "()")
			symbolTable[label] = lineNumber
			continue
		}

		lineNumber++
	}

	_, err := in.Seek(0, io.SeekStart)
	if err != nil {
		log.Fatalf("unable to rewind the input file %s: %w", in.Name(), err)
	}

	lineNumber = 0
	scanner = bufio.NewScanner(in)

	// second pass:
	// translation
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if len(line) == 0 || strings.HasPrefix(line, "/") || strings.HasPrefix(line, "(") {
			continue
		}

		if strings.HasPrefix(line, "@") {
			// handle A-instructions
			var address uint16
			uncommented := strings.SplitN(line, " ", 2)[0]
			symbol := strings.TrimPrefix(uncommented, "@")

			if val, err := strconv.Atoi(symbol); err == nil {
				address = uint16(val)
			} else {
				if _, exists := symbolTable[symbol]; !exists {
					symbolTable[symbol] = getNextRAMAddress()
				}

				address = symbolTable[symbol]
			}

			_, err := writer.WriteString(Reverse(aInstruction(address)) + "\n")
			if err != nil {
				log.Fatalf("error writing instruction to file %s: %w", out.Name(), err)
			}
		} else {
			// handle C-instructions
			var (
				dest, jump, comp uint8
				exists bool
			)

			destStr := "null"
			jumpStr := "null"

			line = strings.SplitN(line, " ", 2)[0]
			destSplit := strings.SplitN(line, "=", 2)
			compJump := destSplit[0]

			if len(destSplit) == 2 {
				compJump = destSplit[1]
				destStr = destSplit[0]
			}

			jumpSplit := strings.SplitN(compJump, ";", 2)
			compStr := jumpSplit[0]

			if len(jumpSplit) == 2 {
				jumpStr = jumpSplit[1]
			}

			if dest, exists = destinationTable[destStr]; !exists {
				log.Fatalf("illegal destination in instruction '%s'", line)
			}

			if jump, exists = jumpTable[jumpStr]; !exists {
				log.Fatalf("illegal jump '%s' in instruction '%s'", jumpStr, line)
			}

			if comp, exists = compTable[compStr]; !exists {
				log.Fatalf("illegal computation in instruction '%s'", line)
			}

			_, err := writer.WriteString(Reverse(cInstruction(dest, comp, jump)) + "\n")
			if err != nil {
				log.Fatalf("error writing instruction to file %s: %w", out.Name(), err)
			}
		}

		lineNumber++
	}

	err = writer.Flush()
	if err != nil {
		log.Fatalf("unable to write buffered data to file %s: %w", out.Name(), err)
	}
}

func aInstruction(address uint16) string {
	aInstr := make([]byte, 16, 16)
	aInstr[15] = '0'

	for i := 14; i > -1; i-- {
		switch address >> i & 1 {
		case 0:
			aInstr[i] = '0'
		case 1:
			aInstr[i] = '1'
		}
	}

	return string(aInstr)
}

func cInstruction(dest, comp, jump uint8) string {
	cInstr := make([]byte, 16, 16)

	i := 15

	for ; i > 12; i-- {
		cInstr[i] = '1'
	}

	for j := 6; i > 5; i, j = i-1, j-1 {
		switch comp >> j & 1 {
		case 0:
			cInstr[i] = '0'
		case 1:
			cInstr[i] = '1'
		}
	}

	for j := 2; i > 2; i, j = i-1, j-1 {
		switch dest >> j & 1 {
		case 0:
			cInstr[i] = '0'
		case 1:
			cInstr[i] = '1'
		}
	}

	for j := 2; i > -1; i, j = i-1, j-1 {
		switch jump >> j & 1 {
		case 0:
			cInstr[i] = '0'
		case 1:
			cInstr[i] = '1'
		}
	}

	return string(cInstr)
}