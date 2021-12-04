package assembler

import (
	"bufio"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/pqkallio/nand2tetris-hack-assembler/util"
)

var (
	nextRamAddress uint16 = 15
	lineNumber     uint16 = 0
)

func getNextRAMAddress() uint16 {
	nextRamAddress++

	return nextRamAddress
}

type Writer interface {
	Write(w *bufio.Writer, i uint16) error
}

type StringWriter struct{}

func (sw *StringWriter) Write(w *bufio.Writer, i uint16) error {
	_, err := w.WriteString(sw.stringifyInstruction(i) + "\n")
	return err
}

func (sw *StringWriter) stringifyInstruction(instr uint16) string {
	strInstr := make([]byte, 16)

	for i := 15; i > -1; i-- {
		switch instr >> i & 1 {
		case 0:
			strInstr[i] = '0'
		case 1:
			strInstr[i] = '1'
		}
	}

	return util.Reverse(string(strInstr))
}

type BinaryWriter struct{}

func (bw *BinaryWriter) Write(w *bufio.Writer, i uint16) error {
	_, err := w.Write([]byte{byte(i >> 8), byte(i)})
	return err
}

type Server struct {
	writer Writer
}

func NewServer(format Format) *Server {
	switch format {
	case Binary:
		return &Server{&BinaryWriter{}}
	default:
		return &Server{&StringWriter{}}
	}
}

func (s *Server) Assemble(in *os.File, out *os.File) {
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
		log.Fatalf("unable to rewind the input file %s: %s", in.Name(), err)
	}

	lineNumber = 0
	scanner = bufio.NewScanner(in)

	// second pass:
	// translation
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// empty line, comment line or label
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

			s.writer.Write(writer, address)
			if err != nil {
				log.Fatalf("error writing instruction to file %s: %s", out.Name(), err)
			}
		} else {
			// handle C-instructions
			var (
				dest, jump, comp uint16
				exists           bool
			)

			destStr := ""
			jumpStr := ""

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

			instr := dest | comp | jump

			err := s.writer.Write(writer, instr)
			if err != nil {
				log.Fatalf("error writing instruction to file %s: %s", out.Name(), err)
			}
		}

		lineNumber++
	}

	err = writer.Flush()
	if err != nil {
		log.Fatalf("unable to write buffered data to file %s: %s", out.Name(), err)
	}
}

type Format string

const (
	Binary      Format = "binary"
	Text        Format = "text"
	NotSelected Format = ""
)

func (f Format) String() string {
	return string(f)
}

func (f Format) IsValid() bool {
	return f == Binary || f == Text || f == NotSelected
}
