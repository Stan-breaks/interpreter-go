package interpreter

import (
	"fmt"
	"strings"

	"github.com/Stan-breaks/app/environment"
	"github.com/Stan-breaks/app/models"
	"github.com/Stan-breaks/app/parse"
	"github.com/Stan-breaks/app/utils"
)

func Interprete(tokens []models.TokenInfo) error {
	currentPosition := 0
	for currentPosition < len(tokens) {
		if currentPosition >= len(tokens) {
			break
		}
		token := tokens[currentPosition]
		switch {
		case strings.HasPrefix(token.Token, "IF"):
			end, err := handleIf(tokens[currentPosition:])
			if err != nil {
				return err
			}
			currentPosition += end
		case strings.HasPrefix(token.Token, "PRINT"):
			end, err := handlePrint(tokens[currentPosition:])
			if err != nil {
				return err
			}
			currentPosition += end
		case strings.HasPrefix(token.Token, "VAR"):
			end, err := handleAssignment(tokens[currentPosition:])
			if err != nil {
				return err
			}
			currentPosition += end
		case strings.HasPrefix(token.Token, "IDENTIFIER"):
			end, err := handleReassignment(tokens[currentPosition:])
			if err != nil {
				return err
			}
			currentPosition += end
		default:
			currentPosition++
		}
	}
	return nil
}

func handleAssignment(tokens []models.TokenInfo) (int, error) {
	if len(tokens) < 4 {
		return 0, fmt.Errorf("incomplete variable declaration")
	}
	nameToken := tokens[1]
	if !strings.HasPrefix(nameToken.Token, "IDENTIFIER") {
		return 0, fmt.Errorf("no identifier found")
	}
	valName := strings.Split(nameToken.Token, " ")[1]
	if !strings.HasPrefix(tokens[2].Token, "EQUAL") {
		return 0, fmt.Errorf("equal not found")
	}
	end := 0
	for i := 3; i < len(tokens); i++ {
		if strings.HasPrefix(tokens[i].Token, "SEMICOLON") {
			end = i
			break
		}
	}
	if end == 0 {
		return 0, fmt.Errorf("no semicolon found")
	}
	expr, err := parse.Parse(tokens[3:end])
	if err != nil {
		return 0, fmt.Errorf("invalid assignment expression")
	}
	value := expr.Evaluate()

	environment.Environment[valName] = value
	return end + 1, nil
}

func handleReassignment(tokens []models.TokenInfo) (int, error) {
	if !strings.HasPrefix(tokens[1].Token, "EQUAL") {
		return 0, fmt.Errorf("no equal found in reassignment")
	}
	valname := strings.Split(tokens[0].Token, " ")[1]
	end := 0
	for i := 2; i < len(tokens); i++ {
		if strings.HasPrefix(tokens[i].Token, "SEMICOLON") {
			end = i
			break
		}
	}
	if end == 0 {
		return 0, fmt.Errorf("no semicolon found in reassignment")
	}
	val, err := parse.Parse(tokens[2:end])
	if err != nil {
		return 0, fmt.Errorf("%s", err[0])
	}
	environment.Environment[valname] = val.Evaluate()
	return end + 1, nil
}

func handleReassignmentCondition(tokens []models.TokenInfo) (models.Node, error) {
	valname := strings.Split(tokens[0].Token, " ")[1]
	val, err := parse.Parse(tokens[2:])
	if err != nil {
		return models.NilNode{}, fmt.Errorf("invalid reassignment expression")
	}
	environment.Environment[valname] = val.Evaluate()
	return val, nil
}

func handleExpression(tokens []models.TokenInfo) (models.Node, error) {
	var val models.Node
	var err error
	var error []string
	if utils.IsReassignmentCondition(tokens) {
		val, err = handleReassignmentCondition(tokens)
		if err != nil {
			return models.NilNode{}, err
		}
		return val, nil
	}
	val, error = parse.Parse(tokens)
	if error != nil {
		return models.NilNode{}, fmt.Errorf("invalid expression: %v", error[0])
	}
	return val, nil
}

func handlePrint(tokens []models.TokenInfo) (int, error) {
	if len(tokens) < 2 {
		return 0, fmt.Errorf("incomplete print statement")
	}

	tokensUsed := 0
	for i := 0; i < len(tokens); i++ {
		if strings.HasPrefix(tokens[i].Token, "SEMICOLON") {
			tokensUsed = i
			break
		}
	}
	if tokensUsed == 0 {
		return 0, fmt.Errorf("no semicolon found after print")
	}
	expr, err := parse.Parse(tokens[1:tokensUsed])
	if err != nil {
		return 0, fmt.Errorf("invalid print expression")
	}
	result := expr.Evaluate()
	fmt.Printf("%v\n", result)

	return tokensUsed + 1, nil
}

func handleIf(tokens []models.TokenInfo) (int, error) {
	conditionStart := -1
	conditionEnd := -1
	ifBodyStart := -1
	ifBodyEnd := -1
	elseBodyStart := -1
	elseBodyEnd := -1
	parenCount := 0
	braceCount := 0
	elseIf := false
	elseIfStart := -1
	elseIfEnd := -1
	for i := 0; i < len(tokens); i++ {
		token := tokens[i].Token
		switch {
		case strings.HasPrefix(token, "LEFT_PAREN"):
			if conditionStart == -1 && parenCount == 0 {
				conditionStart = i
			}
			parenCount++
		case strings.HasPrefix(token, "RIGHT_PAREN"):
			parenCount--
			if parenCount == 0 && conditionEnd == -1 {
				conditionEnd = i
				ifBodyStart = i + 1
				if !strings.HasPrefix(tokens[i+1].Token, "LEFT_BRACE") {
					for j := i + 1; j < len(tokens); j++ {
						if strings.HasPrefix(tokens[j].Token, "SEMICOLON") {
							ifBodyEnd = j
							break
						}
					}
				}
			}
		case strings.HasPrefix(token, "LEFT_BRACE"):
			braceCount++
		case strings.HasPrefix(token, "RIGHT_BRACE"):
			braceCount--
			if braceCount == 0 {
				if ifBodyEnd != -1 && elseBodyEnd == -1 && elseIf == false {
					elseBodyEnd = i
				}
				if ifBodyEnd != -1 && elseBodyEnd == -1 && elseIf == true {
					elseIfEnd = i
					err := Interprete(tokens[elseIfStart:elseIfEnd])
					if err != nil {
						return 0, fmt.Errorf("invalid else if : %v", err.Error())
					}
				}
				if ifBodyEnd == -1 {
					ifBodyEnd = i
				}
			}
		case strings.HasPrefix(token, "ELSE"):
			if braceCount == 0 && elseBodyStart == -1 {
				ifBodyEnd = i - 1
				if strings.HasPrefix(tokens[i+1].Token, "LEFT_BRACE") {
					elseBodyStart = i + 2
				} else if strings.HasPrefix(tokens[i+1].Token, "IF") && elseIf == false {
					elseIf = true
				} else {
					elseBodyStart = i + 1
					for j := i + 1; j < len(tokens); j++ {
						if strings.HasPrefix(tokens[j].Token, "SEMICOLON") {
							elseBodyEnd = j
							break
						}

					}
				}
			}
		default:

		}
	}

	if conditionStart == -1 || conditionEnd == -1 || ifBodyStart == -1 || ifBodyEnd == -1 {
		return 0, fmt.Errorf("malformed if statement")
	}
	condition, err := handleExpression(tokens[conditionStart+1 : conditionEnd])
	if err != nil {
		return 0, fmt.Errorf("invalid if condition: %v", err.Error())
	}
	if condition.Evaluate().(bool) {
		err := Interprete(tokens[ifBodyStart : ifBodyEnd+1])
		if err != nil {
			return 0, fmt.Errorf("invalid if body: %v", err.Error())
		}
	} else {
		if elseBodyEnd != -1 && elseBodyStart != -1 {
			err := Interprete(tokens[elseBodyStart : elseBodyEnd+1])
			if err != nil {
				return 0, fmt.Errorf("invalid else body: %v", err.Error())
			}
		}
	}
	if elseBodyEnd == -1 || elseBodyStart == -1 {
		return ifBodyEnd + 1, nil
	} else {
		return elseBodyEnd + 1, nil
	}
}
