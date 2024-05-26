package gosyms

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Разбить выражение на термы по знакам + и -
func createTermsDiff(expr string) []string {
	var terms []string
	currentTerm := ""
	level := 0
	for _, char := range expr {

		if char == '{' {
			level++
		} else if char == '}' {
			level--
		}

		if (char == '+' || char == '-') && level == 0 {
			if currentTerm != "" {
				terms = append(terms, currentTerm)
			}
			currentTerm = string(char)
		} else {
			currentTerm += string(char)
		}
	}
	if currentTerm != "" {
		terms = append(terms, currentTerm)
	}
	return terms
}

// Найти производную терма
func deriveTerm(term string) string {
	term = strings.ReplaceAll(term, " ", "")

	var parts []string
	var coefficient float64
	var variable string

	var ListVariable []string
	var resultTermDiff string

	// Если в терме разделенном +- есть знак *, то есть терм в виде 2*x делим по умножить
	if strings.Contains(term, "*") {
		parts = splitIgnoringBracesDiff(term, '*')
		// Проверяем каждую часть
		for _, part := range parts {
			if num, err := strconv.ParseFloat(part, 64); err == nil {
				// Если удалось преобразовать в число, это коэффициент
				coefficient = num
			} else {
				// Иначе это переменная
				// Обработка функции sin(x) или cos(x)
				if len(parts) < 3 {
					variable = part
				} else {
					// ЛОГИКА ДЛЯ ПРОИЗВОДНОЙ f(x)*g(x)
					ListVariable = append(ListVariable, part)
				}
			}
		}

		if len(ListVariable) != 0 {
			variable = strings.Join(ListVariable, "#")
		}

		// Костыль
		if coefficient == 0 {
			coefficient = 1
		}

	} else {
		if num, err := strconv.ParseFloat(term, 64); err == nil {
			//Если удалось преобразовать в число, это коэффициент
			coefficient = num
		} else {
			//Иначе это переменная
			variable = term
		}
	}

	// Костыль
	if coefficient == 0 {
		coefficient = 1
	}

	// Основная часть с вычислением производных каждого терма
	resultTermDiff = diffMain(coefficient, variable)

	return resultTermDiff
}

func splitIgnoringBracesDiff(expr string, sep rune) []string {
	var result []string
	var current strings.Builder
	var braceLevel int

	for _, char := range expr {
		switch char {
		case '{':
			braceLevel++
			current.WriteRune(char)
		case '}':
			braceLevel--
			current.WriteRune(char)
		case sep:
			if braceLevel == 0 {
				result = append(result, current.String())
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}
	result = append(result, current.String())
	return result
}

// Найти производную выражения
func Diff(expr string) string {

	// Проверяем на наличие функций недоступных в данной версии
	err := mvpLimitFunctionality(expr)

	if err != nil {
		fmt.Println(err)
		return ""
	}

	// Раскрываем, упрощаем, собираем
	expr = simpDiffExpr(expr)

	terms := createTermsDiff(expr)

	var derivedTerms []string
	for _, term := range terms {
		derivedTerms = append(derivedTerms, deriveTerm(term))
	}
	resStr := strings.Join(derivedTerms, "+")
	resStr = replaceClosingBrackets(resStr)

	// Упрощаем полученное выражение
	resStr = SimplifyExpr(resStr)

	return resStr
}

func diffMain(coeff float64, variable string) string {
	resultStr := ""
	// ЕСЛИ У НАС ПРОСТО ПЕРЕМЕННАЯ x и коэффициента нет
	if (variable == "+x" || variable == "x") && coeff == 1 {
		resultStr = "1"
	}
	// ЕСЛИ У НАС ПРОСТО ПЕРЕМЕННАЯ -x и коэффициента нет
	if variable == "-x" && coeff == 1 {
		resultStr = "-1"
	}

	// ЕСЛИ У НАС КОНСТАНТА
	if variable == "" && coeff != 0 {
		coeff = 0
		resultStr = fmt.Sprintf("%.0f", coeff)
	}

	// Производная выражения 21*x и так далее, по сути, основная часть
	if variable == "x" && coeff != 0 {
		resultStr = fmt.Sprintf("%.0f", coeff)
	}

	// Проверка и дифференцирование выражений вида x^n
	if strings.Contains(variable, "^") {
		re := regexp.MustCompile(`(-?)x\^(\d+)`)
		matches := re.FindStringSubmatch(variable)
		if len(matches) > 0 {
			// Получение знака, степени переменной и обработка коэффициента
			sign := matches[1]
			power, _ := strconv.Atoi(matches[2])
			newCoeff := coeff * float64(power)
			newPower := power - 1

			// Формирование новой переменной
			if newPower == 0 {
				resultStr = fmt.Sprintf("%s%.0f", sign, newCoeff)
			} else if newPower == 1 {
				resultStr = fmt.Sprintf("%s%.0f*x", sign, newCoeff)
			} else {
				resultStr = fmt.Sprintf("%s%.0f*x^%d", sign, newCoeff, newPower)
			}
		}
	}

	// Найдем производные произведений
	if !strings.Contains(variable, "#") {
		switch {
		case strings.Contains(variable, "cos"):
			variable = strings.ReplaceAll(variable, "cos", "-sin")
			newCoeff := diffInFig(variable)
			if strings.Contains(newCoeff, "+") || strings.Contains(newCoeff, "-") {
				newCoeff = "(" + newCoeff + ")"
			}
			resultStr = fmt.Sprintf("%.0f*%s*%s", coeff, variable, newCoeff)
		case strings.Contains(variable, "sin"):
			variable = strings.ReplaceAll(variable, "sin", "cos")
			//resultStr = fmt.Sprintf("%.0f*%s", coeff, variable)
			newCoeff := diffInFig(variable)
			if strings.Contains(newCoeff, "+") || strings.Contains(newCoeff, "-") {
				newCoeff = "(" + newCoeff + ")"
			}
			resultStr = fmt.Sprintf("%.0f*%s*%s", coeff, variable, newCoeff)
		case strings.Contains(variable, "EXP"):
			variable = strings.ReplaceAll(variable, "EXP", "EXP")
			//resultStr = fmt.Sprintf("%.0f*%s", coeff, variable)

			newCoeff := diffInFig(variable)

			if strings.Contains(newCoeff, "+") || strings.Contains(newCoeff, "-") {
				newCoeff = "(" + newCoeff + ")"
			}
			resultStr = fmt.Sprintf("%.0f*%s*%s", coeff, variable, newCoeff)
		}
	} else {
		// ОБРАБАТЫВАЕМ СИТУАЦИЮ (f(x)*g(x))' = f(x)'*g(x)+f(x)*g(x)'
		partsVariable := strings.Split(variable, "#")
		resultStr = ""
		var resList []string

		for i := 0; i < len(partsVariable); i++ {
			derivative := diffMain(coeff, partsVariable[i])
			newLess := ""
			newLess = diffInFig(partsVariable[i])
			if strings.Contains(newLess, "+") || strings.Contains(newLess, "-") {
				newLess = "(" + newLess + ")"
			}
			switch i {
			case 0:
				newPartsVariable := make([]string, len(partsVariable))
				copy(newPartsVariable, partsVariable)
				newPartsVariable = newPartsVariable[i+1:]
				resultStr = derivative + "*" + strings.Join(newPartsVariable, "*")
				resList = append(resList, resultStr)
			case len(partsVariable) - 1:
				newPartsVariable := make([]string, len(partsVariable))
				copy(newPartsVariable, partsVariable)
				newPartsVariable = newPartsVariable[:len(partsVariable)-1]
				resultStr = derivative + "*" + strings.Join(newPartsVariable, "*")
				resList = append(resList, resultStr)
			default:
				newPartsVariable := make([]string, len(partsVariable))
				copy(newPartsVariable, partsVariable)
				newPartsVariable = append(newPartsVariable[:i], newPartsVariable[i+1:]...)
				resultStr = derivative + "*" + strings.Join(newPartsVariable, "*")
				resList = append(resList, resultStr)
			}
		}
		LastResult := strings.Join(resList, "+")
		resultStr = LastResult

	}
	return resultStr
}

func diffInFig(variable string) string {
	var textInFig string
	if strings.Contains(variable, "sin{") ||
		strings.Contains(variable, "cos{") ||
		strings.Contains(variable, "EXP{") {
		if strings.Contains(variable, "{") {
			// Регулярное выражение для поиска содержимого в фигурных скобках
			textInFig = extractInBraces(variable)
			textInFig = Diff(textInFig)
		}
	}
	return textInFig
}

func extractInBraces(expr string) string {
	var result []rune
	var braceContent []rune
	var level int

	for _, char := range expr {
		if char == '{' {
			if level > 0 {
				braceContent = append(braceContent, char)
			}
			level++
		} else if char == '}' {
			level--
			if level > 0 {
				braceContent = append(braceContent, char)
			}
			if level == 0 {
				break
			}
		} else if level > 0 {
			braceContent = append(braceContent, char)
		} else {
			result = append(result, char)
		}
	}

	return string(braceContent)
}

// Аналог функции выше без возврата к круглым скобкам
func simpDiffExpr(expr string) string {
	expr = strings.ReplaceAll(expr, "-(", "-1*(")
	expr = strings.ReplaceAll(expr, "+(", "+1*(")
	if expr[0] == '(' {
		expr = "1*" + expr
	}
	if expr[0] == '-' && expr[1] == '(' {
		expr = "-1*" + expr
	}
	// Раскрываем скобки
	expr = evaluateExpression(expr)
	expr = simplify(expr)

	return expr
}
