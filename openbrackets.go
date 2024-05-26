package gosyms

import (
	"fmt"
	"strconv"
	"strings"
)

func evaluateExpression(expr string) string {
	expr = replaceClosingBrackets(expr)
	expr = strings.ReplaceAll(expr, "+(", "+1*(")
	expr = strings.ReplaceAll(expr, "-(", "-1*(")
	if expr[0] == '(' {
		expr = "+" + expr
	}
	if expr[0] == 'x' || expr[0] == 's' || expr[0] == 'c' ||
		expr[0] == 'E' {
		expr = "1*" + expr
	}
	// Находим индексы скобок
	var openBracketIndex, closeBracketIndex int
	for i, char := range expr {
		if char == '(' {
			openBracketIndex = i
		} else if char == ')' {
			closeBracketIndex = i
			break
		}
	}

	// Раскрываем скобки если они есть
	if openBracketIndex != 0 && closeBracketIndex != 0 {
		subExpr := expr[openBracketIndex+1 : closeBracketIndex]

		var subExprParts []string
		// РАСКРЫВАЕМ СКОБКИ ЕСЛИ УМНОЖЕНИЕ
		if openBracketIndex > 0 && expr[openBracketIndex-1] == '*' {
			// Разбиваем выражение внутри скобок на элементы по знакам + и -
			subExpr = strings.ReplaceAll(subExpr, "-", "+-")
			subExprParts = splitExpr(subExpr)

			// Преобразуем знак перед * в число
			targetIndex := 0
			if openBracketIndex >= 0 && openBracketIndex < len(expr) {
				targetIndex = findNearestLeft(expr, openBracketIndex)
			} else {
				fmt.Println("Неверный индекс")
			}

			if targetIndex == -1 {
				targetIndex = 0
			}
			signValue, err := strconv.Atoi(expr[targetIndex : openBracketIndex-1])
			if err != nil {
				// Если не удалось преобразовать, значит это переменная, добавим множитель к каждому элементу в скобках
				for i, part := range subExprParts {
					subExprParts[i] = part + "*" + expr[targetIndex:openBracketIndex-1]
				}
			} else {
				// Умножаем каждый элемент на значение перед *
				for i, part := range subExprParts {
					// ЕСЛИ X НАХОДИТСЯ ВНУТРИ СКОБОК
					// Если часть содержит переменную 'x', умножаем только на коэффициент
					if strings.Contains(part, "x") {
						subExprParts[i] = strconv.Itoa(signValue) + "*" + part
					} else {
						subExprParts[i] = strconv.Itoa(signValue * atoi(part))
					}
				}
			}
			// Собираем обратно выражение внутри скобок
			subExpr = strings.Join(subExprParts, "+")

			// ЗДЕСЬ МЫ УПРОЩАЕМ ТО ЧТО ВНУТРИ СКОБОК
			subExpr = simplify(subExpr)

			expr = expr[:targetIndex] + "+" + subExpr + expr[closeBracketIndex+1:]
			expr = replaceOperation(expr)
		}

		// Меняем знаки внутри скобки если перед ней стоит знак '-'
		if strings.Contains(subExpr, "(") && strings.Contains(subExpr, ")") {
			// Проверяем знак перед скобкой
			if openBracketIndex > 0 && expr[openBracketIndex-1] == '-' {
				// Меняем знаки внутри скобок
				subExpr = strings.ReplaceAll(subExpr, "+", "$")
				subExpr = strings.ReplaceAll(subExpr, "-", "+")
				subExpr = strings.ReplaceAll(subExpr, "$", "-")
			}
		}
	}

	openBracketIndex--
	closeBracketIndex--

	// Рекурсивно вызываем функцию, пока есть скобки
	if strings.Contains(expr, "(") {
		expr = evaluateExpression(expr)
	}

	return expr
}

// ИЩЕМ СИМВОЛ + ИЛИ - ПЕРЕД КОЭФФИЦИЕНТОМ МНОЖИТЕЛЕМ (вырезаем множитель)
// игнорируя те, которые находятся внутри фигурных скобок
func findNearestLeft(input string, index int) int {
	targets := []rune{'+', '-'}

	runes := []rune(input)
	braceLevel := 0

	for i := index - 1; i >= 0; i-- {
		// Отслеживаем вложенность фигурных скобок
		if runes[i] == '}' {
			braceLevel++
		} else if runes[i] == '{' {
			braceLevel--
		}
		// Если не находимся внутри фигурных скобок
		if braceLevel == 0 {
			for _, target := range targets {
				if runes[i] == target {
					return i
				}
			}
		}
	}
	return -1 // Если символ не найден слева от заданного индекса
}

func atoi(s string) int {
	v, _ := strconv.Atoi(strings.Trim(s, " "))
	return v
}

func replaceOperation(str string) string {
	str = strings.ReplaceAll(str, "--", "+")
	str = strings.ReplaceAll(str, "-+", "-")
	str = strings.ReplaceAll(str, "+-", "-")
	return str
}

// Функция для разбора выражения с учетом вложенных фигурных скобок
func splitExpr(expr string) []string {
	var result []string
	var current strings.Builder
	var insideBraces bool
	var bracesLevel int

	for i, char := range expr {
		switch char {
		case '{':
			if bracesLevel == 0 {
				insideBraces = true
			}
			bracesLevel++
			current.WriteRune(char)
		case '}':
			bracesLevel--
			if bracesLevel == 0 {
				insideBraces = false
			}
			current.WriteRune(char)
		case '+':
			if i != 0 {
				if insideBraces {
					current.WriteRune(char)
				} else {
					result = append(result, current.String())
					current.Reset()
				}
			}
		default:
			current.WriteRune(char)
		}

		// Добавляем последний элемент, если это конец строки
		if i == len(expr)-1 {
			result = append(result, current.String())
		}
	}
	return result
}

func replaceClosingBrackets(expr string) string {
	var level int
	runes := []rune(expr)
	levels := make([]int, len(runes))

	for i := 0; i < len(runes); i++ {
		if runes[i] == '(' {
			level++
			// Проверяем, если функции
			if i >= 3 && (string(runes[i-3:i]) == "sin" || string(runes[i-3:i]) == "cos" ||
				string(runes[i-3:i]) == "EXP" || string(runes[i-3:i]) == "tg" ||
				string(runes[i-3:i]) == "ctg") {
				levels[level] = i
				runes[i] = '{'
			}
		} else if runes[i] == ')' {
			if levels[level] > 0 {
				runes[i] = '}'
				levels[level] = 0
			}
			level--
		}
	}
	return string(runes)
}

func replaceFigBracketsBack(expr string) string {
	expr = strings.ReplaceAll(expr, "{", "(")
	expr = strings.ReplaceAll(expr, "}", ")")
	return expr
}
