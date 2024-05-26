package gosyms

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Заглушка-ограничение функционала согласно требованиям к MVP
func mvpLimitFunctionality(expr string) error {
	unsupportedFunctions := []string{"tg", "ctg", "ln", "log"}
	unsupportedOperations := []string{"/"}

	for _, function := range unsupportedFunctions {
		if strings.Contains(expr, function) {
			return errors.New(fmt.Sprintf("Ошибка: Функция '%s' в настоящее время не поддерживается.", function))
		}
	}

	for _, operation := range unsupportedOperations {
		if strings.Contains(expr, operation) {
			return errors.New(fmt.Sprintf("Ошибка: Операция '%s' в настоящее время не поддерживается.", operation))
		}
	}
	// Если проверка пройдена, возвращаем nil для указания на отсутствие ошибки
	return nil
}

func createTermsSimp(expr string) []string {
	// Проверяем вложенность переменных скобок внутри cos{} или sin{}
	var level int
	// Разделить строку на термы
	var terms []string
	currentTerm := ""
	checkTermsInMinus := ""
	for _, char := range expr {

		if char == '{' || char == '(' {
			level++
		} else if char == '}' || char == ')' {
			// Если уровень вложенности больше нуля, уменьшаем его
			level--
		}
		if (char == '+' || char == '-') && level == 0 {
			if checkTermsInMinus != "*" {
				if currentTerm != "" {
					terms = append(terms, currentTerm)
				}
				currentTerm = string(char)
			} else {
				currentTerm += string(char)
			}

		} else {
			currentTerm += string(char)
		}
		checkTermsInMinus = string(char)
	}
	if currentTerm != "" {
		terms = append(terms, currentTerm)
	}
	return terms
}

func splitTerm(term string) []string {

	var parts []string
	var part strings.Builder
	nestedLevel := 0

	for _, char := range term {
		if char == '{' {
			nestedLevel++
		} else if char == '}' {
			nestedLevel--
		}
		if char == '*' && nestedLevel == 0 {
			parts = append(parts, part.String())
			part.Reset()
		} else {
			part.WriteRune(char)
		}
	}

	if part.Len() > 0 {
		parts = append(parts, part.String())
	}
	return parts
}

// ТЕСТИРУЕМ СИНУСЫ,КОСИНУСЫ И ОСТАЛЬНОЕ
func parseTerm(term string) (float64, string) {
	var coefficient float64
	var variable []string

	if strings.Contains(term, "*") {
		// Разбиваем терм на части
		parts := splitTerm(term)

		// Проверяем каждую часть
		for _, part := range parts {
			if num, err := strconv.ParseFloat(part, 64); err == nil {
				// Если удалось преобразовать в число, это коэффициент
				coefficient = num
			} else {
				// Иначе это переменная
				if strings.HasPrefix(part, "sin{") || strings.HasPrefix(part, "cos{") ||
					strings.HasPrefix(part, "EXP{") ||
					strings.HasPrefix(part, "tg{") || strings.HasPrefix(part, "ctg{") {
					// Обработка функций
					variable = append(variable, part)
				} else {
					variable = append(variable, part)
				}
			}
		}
		if coefficient == 0 {
			coefficient = 1
		}
		variableNew := multiplyVariables(variable)
		// Если есть переменная и коэффициент не равен 0, возвращаем коэффициент и переменную в степени
		if variableNew != "" && coefficient != 0 {
			return coefficient, fmt.Sprintf("%s", variableNew)
		}
		return coefficient, variableNew
	} else {
		varOne := ""
		if num, err := strconv.ParseFloat(term, 64); err == nil {
			// Если удалось преобразовать в число, это коэффициент
			coefficient = num
		} else {
			// Иначе это переменная
			varOne = term
		}
		if coefficient == 0 {
			coefficient = 1
		}
		return coefficient, varOne
	}
}

func checkNumber(str string) (bool, float64) {
	if num, err := strconv.ParseFloat(str, 64); err == nil {
		// Если удалось преобразовать строку в число, то это коэффициент и возвращаем True
		return true, num
	} else {
		return false, 0
	}
}

func simplifyPow(term string) string {
	parts := strings.Split(term, "^")
	checkNumbBase, numbBase := checkNumber(parts[0])
	checkNumbPow, numbPow := checkNumber(parts[1])
	if strings.Contains(term, "^") && checkNumbBase && checkNumbPow {
		base := numbBase
		pow := numbPow
		//fmt.Println("Base: ", base)
		//fmt.Println("pow: ", pow)
		result := math.Pow(base, pow)
		return strconv.FormatFloat(result, 'f', -1, 64)
	}
	return term
}

// Разбиваем выражение по знаку ^, игнорируя те случаи, когда знак находится внутри фигурных скобок
func splitByCaretOutsideBraces(expr string) []string {
	var parts []string
	var currentPart []rune
	braceLevel := 0

	for _, char := range expr {
		if char == '{' {
			braceLevel++
		} else if char == '}' {
			braceLevel--
		}

		if char == '^' && braceLevel == 0 {
			if len(currentPart) > 0 {
				parts = append(parts, string(currentPart))
				currentPart = []rune{}
			}
		} else {
			currentPart = append(currentPart, char)
		}
	}

	if len(currentPart) > 0 {
		parts = append(parts, string(currentPart))
	}

	return parts
}

func multiplyVariables(Variable []string) string {
	if Variable == nil {
		return ""
	}

	terms := make(map[string]int)

	for _, part := range Variable {
		// Разбиваем на переменную и степень если не внутри фигурных скобок
		termParts := splitByCaretOutsideBraces(part)

		variable := termParts[0]
		powerStr := "1" // По умолчанию степень 1, если не указана

		if len(termParts) > 1 {
			powerStr = termParts[1]
		}
		power, err := strconv.Atoi(powerStr)
		if err != nil {
			fmt.Println("Ошибка преобразования степени:", err)
			return ""
		}

		// Добавляем переменную и степень в мап
		if existingPower, ok := terms[variable]; ok {
			terms[variable] = existingPower + power
		} else {
			terms[variable] = power
		}
	}

	// Формируем итоговую строку
	var result []string
	for variable, power := range terms {
		if power == 1 {
			result = append(result, variable)
		} else {
			result = append(result, fmt.Sprintf("%s^%d", variable, power))
		}
	}

	finalResult := strings.Join(result, "*")
	return finalResult
}

func simplifyTerm(term string) string {

	// Для корректной обработки выражения в том случае, если оно начинается с "-x..."
	term = strings.ReplaceAll(term, "+x", "+1*x")
	term = strings.ReplaceAll(term, "-x", "-1*x")

	term = strings.ReplaceAll(term, "-sin", "-1*sin")
	term = strings.ReplaceAll(term, "-cos", "-1*cos")
	term = strings.ReplaceAll(term, "+sin", "+1*sin")
	term = strings.ReplaceAll(term, "+cos", "+1*cos")

	term = strings.ReplaceAll(term, "-tg", "-1*tg")
	term = strings.ReplaceAll(term, "+tg", "+1*tg")
	term = strings.ReplaceAll(term, "-ctg", "-1*ctg")
	term = strings.ReplaceAll(term, "+ctg", "+1*ctg")

	term = strings.ReplaceAll(term, "-EXP", "-1*EXP")
	term = strings.ReplaceAll(term, "+EXP", "+1*EXP")

	// Проверяем, есть ли скобки в терме
	if strings.Contains(term, "(") && strings.Contains(term, ")") {
		// Разбиваем терм на две части: до скобки и после скобки
		parts := strings.Split(term, "(")
		coeff, varname := parseTerm(parts[0])
		if coeff != 1 {
			term = fmt.Sprintf("%.2f", coeff) + "*" + varname
		} else {
			term = varname
		}

		innerTerm := strings.TrimRight(parts[1], ")")
		innerTerm = strings.TrimSpace(innerTerm)

		// Для внутреннего терма обрабатываем его отдельно
		innerSimplified := simplifyTerm(innerTerm)

		// Собираем упрощенный терм с учетом внутреннего упрощенного терма
		if term != "" {
			return term + "*" + innerSimplified
		}
		return innerSimplified
	}
	///////////////////////////////////////
	// Разделить терм на коэффициенты и переменные
	if strings.Contains(term, "*") {
		parts := strings.Split(term, "*")

		// Проверяем есть ли в списке выражения в виде 5^2 которые можно вычислить
		for i, part := range parts {
			if strings.Contains(part, "^") {
				parts[i] = simplifyPow(part)
			}
		}
		var coefficients []float64
		var variables []string
		for _, part := range parts {
			num, err := strconv.ParseFloat(part, 64)
			if err == nil {
				coefficients = append(coefficients, num)
			} else {
				variables = append(variables, part)
			}
		}

		// Перемножить коэффициенты
		totalCoefficient := 1.0
		for _, coeff := range coefficients {
			totalCoefficient *= coeff
		}

		// Собрать упрощенное выражение
		var simplifiedExpr strings.Builder
		if totalCoefficient != 0 {
			if totalCoefficient != 1 {
				simplifiedExpr.WriteString(strconv.FormatFloat(totalCoefficient, 'f', -1, 64))
				simplifiedExpr.WriteString("*") // Добавляем знак умножения
			}
			simplifiedExpr.WriteString(strings.Join(variables, "*"))
		}
		return simplifiedExpr.String()
	} else {

		// Проверяем есть ли в списке выражения в виде x, cos(x), sin(x) и так далее
		switch term {
		case "x":
			term = strings.ReplaceAll(term, "x", "1*x")
		case "sin":
			term = strings.ReplaceAll(term, "sin", "1*sin")
		case "cos":
			term = strings.ReplaceAll(term, "cos", "1*cos")
		case "tg":
			term = strings.ReplaceAll(term, "tg", "1*tg")
		case "ctg":
			term = strings.ReplaceAll(term, "ctg", "1*ctg")
		case "EXP":
			term = strings.ReplaceAll(term, "EXP", "1*EXP")
		}
		isTermTrue := false

		if len(term) > 2 {
			if term[1] == 'x' ||
				(term[1] == 's' && term[2] == 'i' && term[3] == 'n') ||
				(term[1] == 'c' && term[2] == 'o' && term[3] == 's') ||
				(term[1] == 't' && term[2] == 'g' ||
					(term[1] == 'c' && term[2] == 't' && term[3] == 'g') ||
					(term[1] == 'E' && term[2] == 'X' && term[3] == 'P')) {
				isTermTrue = true
			}
		}

		// УПРОЩАЕМ МЕСТА С +-X в +1*X
		if (term[0] == '+' || term[0] == '-') && isTermTrue && len(term) > 1 {
			var simplifiedExpr strings.Builder
			simplifiedExpr.WriteByte(term[0])
			simplifiedExpr.WriteString("1*" + term[1:])
			return simplifiedExpr.String()
		} else {
			return term
		}

	}
}

func simplify(expr string) string {
	// Проводим чистку знаков
	expr = deleteCoefOne(expr)
	// Разделяем строку на термы с помощью функции CreateTerms(expr)
	terms := createTermsSimp(expr)

	// Создаем карту для хранения коэффициентов для каждого терма
	coefficients := make(map[string]float64)

	for _, term := range terms {
		// Парсим коэффициент и переменную из терма
		if strings.Contains(term, "{") {
			var textInFig string
			ExtractContent := extractContentInBraces(term)
			textInFig = simplify(ExtractContent)
			term = strings.ReplaceAll(term, ExtractContent, textInFig)
		}
		if strings.Contains(term, "#") {
			term = strings.ReplaceAll(term, "#", "*")
		}
		if term == "+0" || term == "-0" || term == "0" {
			term = ""
		} else {
			term = simplifyTerm(term)
			// Разделяем элементы по карте в виде x:5
			coef, variable := parseTerm(term)
			// Добавляем коэффициент к существующему значению в карте
			coefficients[variable] += coef
		}
	}

	// Собираем упрощенное выражение
	var simplifiedExpr strings.Builder
	for variable, coefficient := range coefficients {
		if coefficient != 0 {
			// В нашем случае из-за того, что в +coef + не учитывается, то приходится добавлять
			if simplifiedExpr.Len() > 0 && coefficient > 0 {
				simplifiedExpr.WriteString("+")
			}
			if variable == "" {
				simplifiedExpr.WriteString(strconv.FormatFloat(coefficient, 'f', -1, 64))
				simplifiedExpr.WriteString(variable)
			} else {
				simplifiedExpr.WriteString(strconv.FormatFloat(coefficient, 'f', -1, 64))
				simplifiedExpr.WriteString("*")
				simplifiedExpr.WriteString(variable)
			}
		}
	}

	var resultStr string

	if len(simplifiedExpr.String()) != 1 {
		resultStr = strings.ReplaceAll(simplifiedExpr.String(), "x", "x")
	} else {
		resultStr = simplifiedExpr.String()
	}

	// Удаляем коэффициент если он равен 1
	resultStr = deleteCoefOne(resultStr)

	// Если итоговая строка пустая, то есть все сократилось, ответ = 0
	if resultStr == "" {
		resultStr = "0"
	}
	return resultStr
}

func extractContentInBraces(expr string) string {
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

func deleteCoefOne(resultStr string) string {
	if resultStr == "" {
		return resultStr
	}

	for i := 0; i < 3; i++ {

		resultStr = replaceAll(resultStr)

		partsStr := splitIgnoringBracesSimp(resultStr, '+')

		resultStr = strings.Join(partsStr, "+")
		resultStr = replaceAll(resultStr)
	}
	return resultStr
}

func clearOne(str string) string {
	re := regexp.MustCompile(`(?:(\+|-|\(|^))1\*`)
	return re.ReplaceAllString(str, "$1")
}

func splitIgnoringBracesSimp(expr string, delimiter rune) []string {
	var parts []string
	var currentPart []rune
	var braceLevel, parenLevel int

	for _, char := range expr {
		switch char {
		case '{':
			braceLevel++
		case '}':
			braceLevel--
		case '(':
			parenLevel++
		case ')':
			parenLevel--
		}

		if char == delimiter && braceLevel == 0 && parenLevel == 0 {
			parts = append(parts, string(currentPart))
			currentPart = nil
		} else {
			currentPart = append(currentPart, char)
		}
	}
	if len(currentPart) > 0 {
		parts = append(parts, string(currentPart))
	}
	return parts
}

func replaceAll(resultStr string) string {
	//resultStr = strings.ReplaceAll(resultStr, "1x", "x")
	//resultStr = strings.ReplaceAll(resultStr, "-1x", "-x")

	resultStr = strings.ReplaceAll(resultStr, "+1*", "+")
	resultStr = strings.ReplaceAll(resultStr, "-1*", "-")

	resultStr = strings.ReplaceAll(resultStr, "-*", "*")
	resultStr = strings.ReplaceAll(resultStr, "+*", "*")
	resultStr = strings.ReplaceAll(resultStr, "**", "*")
	resultStr = strings.ReplaceAll(resultStr, "*+", "*")

	resultStr = strings.ReplaceAll(resultStr, "-+", "-")
	resultStr = strings.ReplaceAll(resultStr, "++", "+")
	resultStr = strings.ReplaceAll(resultStr, "+-", "-")
	resultStr = strings.ReplaceAll(resultStr, "--", "+")
	return resultStr
}

// Основная программа упрощения
func SimplifyExpr(expr string) string {
	// Проверяем на наличие функций недоступных в данной версии
	err := mvpLimitFunctionality(expr)

	if err != nil {
		fmt.Println(err)
		return ""
	}

	expr = strings.ReplaceAll(expr, " ", "")

	// Раскрываем скобки
	expr = evaluateExpression(expr)
	expr = simplify(expr)
	expr = replaceFigBracketsBack(expr)
	expr = clearOne(expr)

	return expr
}
