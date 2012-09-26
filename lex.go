package lex

import "fmt"
import "os"
import "os/exec"
import "encoding/json"
import "github.com/timtadh/regex-machines/inst"
import "github.com/timtadh/regex-machines/machines"

type Token interface {
    Name() string
    Attribute() Attribute
}

type Attribute interface {
    String() string
}

type ProcessMatch func([]byte)(bool, Token)

type Pattern struct {
    regex string
    process ProcessMatch
}

type JsonInst struct {
    Op string
    X uint32
    Y uint32
}

func (self *JsonInst) String() string {
    return fmt.Sprintf("<JsonInst %v %v %v>", self.Op, self.X, self.Y)
}

var opcode map[string]byte = map[string]byte{
    "CHAR":inst.CHAR,
    "MATCH":inst.MATCH,
    "SPLIT":inst.SPLIT,
    "JMP":inst.JMP }

func Lex(patterns []*Pattern, text []byte) (<-chan Token) {
    tokens := make(chan Token)
    programs := make([][]*inst.Inst, 0, len(patterns))
    for _, pattern := range patterns {
        cmd := exec.Command("rajax", "--format=json", pattern.regex)
        if output, err := cmd.Output(); err != nil {
            panic("could not compile expressions")
        } else {
            var jinsts []*JsonInst
            insts := make([]*inst.Inst, 0)
            json.Unmarshal(output, &jinsts)
            for _, jinst := range jinsts {
                insts = append(insts, inst.New(opcode[jinst.Op], jinst.X, jinst.Y))
            }
            programs = append(programs, insts)
        }
    }
    program := make([]*inst.Inst, 0)
    offsets := make([]uint32, 0)
    offset := uint32(len(programs) - 1)
    process_funcs := make(map[int]ProcessMatch)
    for i, insts := range programs {
        offsets = append(offsets, offset)
        offset += uint32(len(insts))
        if i + 1 < len(programs) {
            program = append(program, inst.New(inst.SPLIT, uint32(i+1), offset))
        }
        process_funcs[int(offset-1)] = patterns[i].process
    }
    for i, insts := range programs {
        offset := offsets[i]
        for _, _inst := range insts {
            switch _inst.Op {
            case inst.SPLIT:
                program = append(program, inst.New(inst.SPLIT, offset+_inst.X, offset+_inst.Y))
            case inst.JMP:
                program = append(program, inst.New(inst.JMP, offset+_inst.X, 0))
            default:
                program = append(program, _inst)
            }
        }
    }
    go func() {
        success, matches := machines.LexerEngine(program, text)
        go func() {
            for match := range matches {
                if ok, Token := process_funcs[match.PC](match.Bytes); ok {
                    tokens <- Token
                }
            }
            close(tokens)
        }()
        if !(<-success) {
            fmt.Fprintln(os.Stderr, "lexing failed")
        }
    }()
    return tokens
}

