package prsr

import (
	"fmt"
	"go/format"
	"regexp"
	"strings"
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

func (s *Parser) extractFullBlock(input string, offset int) (content string, fullBlock string) {
	start := strings.Index(input[offset:], "{")
	if start == -1 {
		return "", ""
	}
	start += offset

	count := 0
	for i := start; i < len(input); i++ {
		if input[i] == '{' {
			count++
		} else if input[i] == '}' {
			count--
			if count == 0 {
				return input[start+1 : i], input[offset : i+1]
			}
		}
	}

	return "", ""
}

func (s *Parser) Parse(data string) ([]byte, error) {
	reClassHeader := regexp.MustCompile(`(?m)^class\s+(\w+)(?:\s+extends\s+(\w+))?\s*\{`)
	reMethod := regexp.MustCompile(`(?s)func\s+(\w+)\s*\((.*?)\)\s*(\w+)?\s*\{`)
	reField := regexp.MustCompile(`(?m)^\s*(\w+)\s+([a-zA-Z0-9_\[\]\*]+)\s*$`)

	finalResult := data

	for {
		loc := reClassHeader.FindStringIndex(finalResult)
		if loc == nil {
			break
		}

		headerMatch := reClassHeader.FindStringSubmatch(finalResult[loc[0]:])
		className := headerMatch[1]
		parent := headerMatch[2]

		content, fullBlock := s.extractFullBlock(finalResult, loc[0])
		if fullBlock == "" {
			break
		}

		var gen strings.Builder
		gen.WriteString(fmt.Sprintf("\ntype %s struct {\n", className))
		if parent != "" {
			gen.WriteString(fmt.Sprintf("\t%s\n", parent))
		}

		fields := reField.FindAllStringSubmatch(content, -1)
		for _, f := range fields {
			if f[1] == "func" || f[1] == "return" || f[1] == "if" {
				continue
			}
			gen.WriteString(fmt.Sprintf("\t%s %s\n", f[1], f[2]))
		}
		gen.WriteString("}\n")

		methodCursor := content
		for {
			mMatch := reMethod.FindStringSubmatchIndex(methodCursor)
			if mMatch == nil {
				break
			}

			name := methodCursor[mMatch[2]:mMatch[3]]
			params := methodCursor[mMatch[4]:mMatch[5]]
			retType := ""
			if mMatch[6] != -1 {
				retType = " " + methodCursor[mMatch[6]:mMatch[7]]
			}

			mBody, mFull := s.extractFullBlock(methodCursor, mMatch[0])
			methodCursor = strings.Replace(methodCursor, mFull, "", 1)

			mBody = strings.ReplaceAll(mBody, "this.", "self.")
			mBody = strings.ReplaceAll(mBody, "this", "self")

			if name == "constructor" {
				gen.WriteString(fmt.Sprintf("\nfunc New%s(%s) *%s {\n\tself := &%s{}\n\t%s\n\treturn self\n}\n",
					className, params, className, className, mBody))
			} else {
				gen.WriteString(fmt.Sprintf("\nfunc (self *%s) %s(%s)%s {\n%s\n}\n",
					className, name, params, retType, mBody))
			}
		}

		finalResult = strings.Replace(finalResult, fullBlock, gen.String(), 1)
	}

	return format.Source([]byte(finalResult))
}
