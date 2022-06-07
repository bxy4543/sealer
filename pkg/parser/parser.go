// Copyright © 2021 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/sealerio/sealer/types/api/v1"
	strUtils "github.com/sealerio/sealer/utils/strings"
	"github.com/sealerio/sealer/version"
	"github.com/sirupsen/logrus"
)

const (
	Run  = "RUN"
	Cmd  = "CMD"
	Copy = "COPY"
	From = "FROM"
	Arg  = "ARG"
)

var validCommands = map[string]bool{
	Run:  true,
	Cmd:  true,
	Copy: true,
	From: true,
	Arg:  true,
}

var (
	reWhitespace = regexp.MustCompile(`[\t\v\f\r ]+`)
	utf8bom      = []byte{0xEF, 0xBB, 0xBF}
)

type Interface interface {
	Parse(kubeFile []byte) (*v1.Image, error)
}

type Parser struct{}

func NewParse() Interface {
	return &Parser{}
}

func (p *Parser) Parse(kubeFile []byte) (*v1.Image, error) {
	image := &v1.Image{
		TypeMeta: metaV1.TypeMeta{APIVersion: "", Kind: "Image"},
		Spec:     v1.ImageSpec{SealerVersion: version.Get().GitVersion},
		Status:   v1.ImageStatus{},
	}

	currentLine := 0
	scanner := bufio.NewScanner(bytes.NewReader(kubeFile))
	scanner.Split(scanLines)
	for scanner.Scan() {
		bytesRead := scanner.Bytes()
		if currentLine == 0 {
			// First line, strip the BOM.
			bytesRead = bytes.TrimPrefix(bytesRead, utf8bom)
		}
		if bytes.HasPrefix(bytesRead, []byte("#")) {
			continue
		}
		bytesRead = processLine(bytesRead, true)
		currentLine++

		line, isEndOfLine := trimContinuationCharacter(string(bytesRead))
		if isEndOfLine && line == "" {
			continue
		}

		for !isEndOfLine && scanner.Scan() {
			bytesRead = processLine(scanner.Bytes(), false)
			if bytes.HasPrefix(bytesRead, []byte("#")) {
				continue
			}

			currentLine++

			if isEmptyContinuationLine(bytesRead) {
				continue
			}
			continuationLine := string(bytesRead)
			continuationLine, isEndOfLine = trimContinuationCharacter(continuationLine)
			line += continuationLine
		}

		layerType, layerValue, err := decodeLine(line)
		if err != nil {
			return nil, fmt.Errorf("decode kubeFile line failed, line: %d ,err: %v", currentLine, err)
		}

		switch layerType {
		case Arg:
			dispatchArg(layerValue, image)
		case Cmd:
			dispatchCmd(layerValue, image)
		default:
			dispatchDefault(layerType, layerValue, image)
		}
	}
	return image, nil
}

func decodeLine(line string) (string, string, error) {
	cmdline := trimCommand(line)
	if len(cmdline) != 2 {
		return "", "", fmt.Errorf("invalid command line %s", line)
	}
	cmd := strings.ToUpper(cmdline[0])
	if !validCommands[cmd] {
		return "", "", fmt.Errorf("invalid command %s %s", cmdline[0], line)
	}

	return cmd, cmdline[1], nil
}

func dispatchArg(layerValue string, ima *v1.Image) {
	if ima.Spec.ImageConfig.Args.Current == nil {
		ima.Spec.ImageConfig.Args.Current = map[string]string{}
	}

	kv := strings.Split(layerValue, ",")
	for _, element := range kv {
		valueLine := strings.SplitN(element, "=", 2)
		if len(valueLine) != 2 {
			logrus.Errorf("invalid ARG value %s. ARG format must be key=value", layerValue)
			return
		}
		k := strings.TrimSpace(valueLine[0])
		if !strUtils.IsLetterOrNumber(k) {
			logrus.Errorf("ARG key must be letter or number,invalid ARG format will ignore this key %s.", k)
			return
		}
		ima.Spec.ImageConfig.Args.Current[k] = strings.TrimSpace(valueLine[1])
	}
}

func dispatchCmd(layerValue string, ima *v1.Image) {
	if ima.Spec.ImageConfig.Cmd.Current == nil {
		ima.Spec.ImageConfig.Cmd.Current = make([]string, 0)
	}

	var cmdList []string
	for _, value := range strings.Split(layerValue, ",") {
		if value == "" {
			continue
		}
		cmdList = append(cmdList, strings.TrimSpace(value))
	}
	ima.Spec.ImageConfig.Cmd.Current = append(ima.Spec.ImageConfig.Cmd.Current, cmdList...)
}

func dispatchDefault(layerType, layerValue string, ima *v1.Image) {
	ima.Spec.Layers = append(ima.Spec.Layers, v1.Layer{
		ID:    "",
		Type:  layerType,
		Value: layerValue,
	})
}

func trimNewline(src []byte) []byte {
	return bytes.TrimRight(src, "\r\n")
}
func trimLeadingWhitespace(src []byte) []byte {
	return bytes.TrimLeftFunc(src, unicode.IsSpace)
}

func isEmptyContinuationLine(line []byte) bool {
	return len(trimLeadingWhitespace(trimNewline(line))) == 0
}

func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[0 : i+1], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func processLine(token []byte, stripLeftWhitespace bool) []byte {
	token = trimNewline(token)
	if stripLeftWhitespace {
		token = trimLeadingWhitespace(token)
	}
	return token
}

func trimContinuationCharacter(line string) (string, bool) {
	s := "\\"
	var re = regexp.MustCompile(`([^\` + s + `])\` + s + `[ \t]*$|^\` + s + `[ \t]*$`)
	if re.MatchString(line) {
		line = re.ReplaceAllString(line, "$1")
		return line, false
	}
	return line, true
}

func trimCommand(line string) []string {
	return reWhitespace.Split(strings.TrimSpace(line), 2)
}
