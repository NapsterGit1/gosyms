package gosyms

import (
	"errors"
	"regexp"
	"strings"
)

func validateExpression(expr string) error {
	// Проверка на соответствие открывающих и закрывающих скобок
	if !checkBracketsBalance(expr) {
		return errors.New("Некорректное использование скобок")
	}

	if !checkPowDuplicate(expr) {
		return errors.New("Степенная башня!")
	}

	if !checkBracketsContent(expr) {
		return errors.New("Пустые скобки")
	}

	if expr == "" {
		return errors.New("Введите выражение. Строка не может быть пустой")
	}

	// Проверка на наличие недопустимых символов
	if !checkValidCharacters(expr) {
		return errors.New("Недопустимые символы в выражении")
	}

	// Проверка на повторение переменной x
	if checkRepeatedVariable(expr) {
		return errors.New("Повторение переменной x")
	}

	// Проверка на повторение символов
	if checkNoConsecutiveOperators(expr) {
		return errors.New("Два и более оператора подряд!")
	}

	if !checkPowBrackets(expr) {
		return errors.New("Возведение скобки или функции в степень не поддерживается.")
	}

	if !checkValidExponent(expr) {
		return errors.New("Некорректная степень")
	}

	if !checkFirstSymbol(expr) {
		return errors.New("Некорректное положение операторов")
	}

	// Если все проверки пройдены успешно, возвращаем nil
	return nil
}

// Проверка корректного использования скобок
func checkBracketsBalance(expr string) bool {
	var stack []rune

	for _, char := range expr {
		if char == '(' {
			stack = append(stack, char)
		} else if char == ')' {
			if len(stack) == 0 || stack[len(stack)-1] != '(' {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}

	return len(stack) == 0
}

func checkFirstSymbol(expr string) bool {
	if (expr[0] == '*' || expr[0] == '+' || expr[0] == '^') || (expr[len(expr)-1] == '*' ||
		expr[len(expr)-1] == '+' || expr[len(expr)-1] == '-' || expr[len(expr)-1] == '^') ||
		(len(expr) == 1 && expr[0] == '-') {
		return false
	}
	return true
}

func checkValidCharacters(expr string) bool {
	// Регулярное выражение для проверки наличия недопустимых символов и слова "sin(", "cos(", "EXP(" `

	regex := `^[0-9+\-*^(){}xsincosEXP]+$`
	//regex := `^[0-9+\-*^()x]+$|\bsin\(|\bcos\(|\bEXP\(+$`
	//regex := `[^0-9+\-*^()x{}]|(?:(?:sin|cos|EXP)[({][^})]+[)}])`
	//regex := `[^0-9+\-*^()x{}]|(?:(?:sin|cos|EXP)[({][^})]+[)}])|sin{[^{}]*}|EXP{[^{}]*}`

	// Компилируем регулярное выражение
	re := regexp.MustCompile(regex)

	// Проверяем, содержит ли выражение недопустимые символы или слова "sin(", "cos(", "EXP("
	return re.MatchString(expr)
}

// Проверка наличия хотя бы одного символа между скобками
func checkBracketsContent(expr string) bool {
	if strings.Contains(expr, "(") && strings.Contains(expr, ")") {
		// Находим индекс открывающей скобки
		openIndex := strings.Index(expr, "(")
		// Находим индекс закрывающей скобки
		closeIndex := strings.Index(expr, ")")
		// Проверяем, что индексы скобок найдены и между ними есть хотя бы один символ
		return openIndex != -1 && closeIndex != -1 && closeIndex-openIndex > 1
	}
	return true
}

// Проверка на повторение переменной x
func checkRepeatedVariable(expr string) bool {
	prevWasVariable := false

	for _, char := range expr {
		if char == 'x' {
			if prevWasVariable {
				return true
			}
			prevWasVariable = true
		} else {
			prevWasVariable = false
		}
	}

	return false
}

func checkPowDuplicate(expr string) bool {
	// Регулярное выражение для поиска шаблона "число^число^"
	regex := `(x|\d+)\^\d+\^`

	// Компилируем регулярное выражение
	re := regexp.MustCompile(regex)

	// Проверяем, совпадает ли строка с шаблоном
	if !re.MatchString(expr) {
		return true // Если нет такого паттерна, возвращаем true
	}

	// Если найден паттерн "число^число^", возвращаем false
	return false
}

func checkNoConsecutiveOperators(expr string) bool {
	// Регулярное выражение для поиска двух и более операторов подряд
	regex := `[\+\-\*^]{2,}`

	// Компилируем регулярное выражение
	re := regexp.MustCompile(regex)

	// Проверяем, содержит ли строка два и более операторов подряд
	return re.MatchString(expr)
}

func checkPowBrackets(expr string) bool {
	if strings.Contains(expr, ")^") {
		return false
	}
	return true
}

func checkValidExponent(expr string) bool {
	// Проверяем, есть ли в выражении знак '^' (степень)
	if strings.Contains(expr, "^") {
		// Регулярное выражение для поиска допустимых степеней
		regex := `\^0$|\^[1-9]`

		// Компилируем регулярное выражение
		re := regexp.MustCompile(regex)

		// Проверяем, содержит ли строка допустимые степени
		return re.MatchString(expr)
	}
	// Если степени нет, то возвращаем true
	return true
}
