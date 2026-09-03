package format

import (
	"bytes"
	"strings"
)

// Format formats Necto source code according to standard canonical style.
func Format(source string) (string, error) {
	lines := strings.Split(source, "\n")
	var formattedLines []string
	indent := 0
	inMultiComment := false

	for _, rawLine := range lines {
		trimmed := strings.TrimSpace(rawLine)

		if trimmed == "" {
			// Не допускаем более двух пустых строк подряд
			if len(formattedLines) > 0 && formattedLines[len(formattedLines)-1] != "" {
				formattedLines = append(formattedLines, "")
			}
			continue
		}

		// Проверка многострочных комментариев
		if inMultiComment {
			formattedLines = append(formattedLines, strings.Repeat("    ", indent)+trimmed)
			if strings.Contains(trimmed, "*/") {
				inMultiComment = false
			}
			continue
		}

		if strings.HasPrefix(trimmed, "/*") {
			formattedLines = append(formattedLines, strings.Repeat("    ", indent)+trimmed)
			if !strings.Contains(trimmed, "*/") {
				inMultiComment = true
			}
			continue
		}

		// Если строка начинается с закрывающей фигурной скобки, уменьшаем отступ ДО добавления
		closeBracesAtStart := 0
		for _, ch := range trimmed {
			if ch == '}' {
				closeBracesAtStart++
			} else if ch != ' ' && ch != '\t' {
				break
			}
		}

		currIndent := indent - closeBracesAtStart
		if currIndent < 0 {
			currIndent = 0
		}

		// Форматируем токены и пробелы внутри строки (если это не чистый комментарий)
		formattedLine := formatLineTokens(trimmed)

		// Добавляем выровненную строку
		formattedLines = append(formattedLines, strings.Repeat("    ", currIndent)+formattedLine)

		// Подсчитываем изменение уровня отступов на основе всех скобок в строке
		openCount := countNonCommentBraces(trimmed, '{')
		closeCount := countNonCommentBraces(trimmed, '}')
		indent += openCount - closeCount
		if indent < 0 {
			indent = 0
		}
	}

	// Удаляем лишние завершающие пустые строки и гарантируем один перенос
	for len(formattedLines) > 0 && formattedLines[len(formattedLines)-1] == "" {
		formattedLines = formattedLines[:len(formattedLines)-1]
	}

	return strings.Join(formattedLines, "\n") + "\n", nil
}

func countNonCommentBraces(line string, brace rune) int {
	count := 0
	inString := false
	var strChar rune
	runes := []rune(line)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if inString {
			if r == strChar && (i == 0 || runes[i-1] != '\\') {
				inString = false
			}
			continue
		}

		if r == '"' || r == '\'' {
			inString = true
			strChar = r
			continue
		}

		// Если начался комментарий //, дальше не считаем
		if r == '/' && i+1 < len(runes) && runes[i+1] == '/' {
			break
		}

		if r == brace {
			count++
		}
	}

	return count
}

func formatLineTokens(line string) string {
	// Если строка целиком является комментарием, возвращаем как есть
	if strings.HasPrefix(line, "//") {
		return line
	}

	var out bytes.Buffer
	runes := []rune(line)
	inString := false
	var strChar rune

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if inString {
			out.WriteRune(r)
			if r == strChar && (i == 0 || runes[i-1] != '\\') {
				inString = false
			}
			continue
		}

		if r == '"' || r == '\'' {
			inString = true
			strChar = r
			out.WriteRune(r)
			continue
		}

		// Начало комментария
		if r == '/' && i+1 < len(runes) && runes[i+1] == '/' {
			// Гарантируем пробел перед комментарием, если перед ним был код
			if out.Len() > 0 && !strings.HasSuffix(out.String(), " ") {
				out.WriteString(" ")
			}
			out.WriteString(string(runes[i:]))
			break
		}

		// Стрелка возврата типа ->
		if r == '-' && i+1 < len(runes) && runes[i+1] == '>' {
			if out.Len() > 0 && !strings.HasSuffix(out.String(), " ") {
				out.WriteRune(' ')
			}
			out.WriteString("-> ")
			i++
			continue
		}

		// Открывающая фигурная скобка: пробел перед {
		if r == '{' {
			if out.Len() > 0 && !strings.HasSuffix(out.String(), " ") {
				out.WriteRune(' ')
			}
			out.WriteRune('{')
			continue
		}

		// Запятые: всегда пробел после запятой
		if r == ',' {
			out.WriteRune(',')
			if i+1 < len(runes) && runes[i+1] != ' ' && runes[i+1] != '\t' && runes[i+1] != '\n' {
				out.WriteRune(' ')
			}
			continue
		}

		// Двоеточие: пробел после двоеточия (кроме '::')
		if r == ':' {
			out.WriteRune(':')
			if i+1 < len(runes) && runes[i+1] != ' ' && runes[i+1] != ':' && runes[i+1] != '\t' {
				out.WriteRune(' ')
			}
			continue
		}

		out.WriteRune(r)
	}

	// Схлопываем лишние внутренние пробелы (вне строк)
	return cleanExcessSpaces(out.String())
}

func cleanExcessSpaces(line string) string {
	var res bytes.Buffer
	inString := false
	var strChar rune
	runes := []rune(line)
	spaceCount := 0

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if inString {
			res.WriteRune(r)
			if r == strChar && (i == 0 || runes[i-1] != '\\') {
				inString = false
			}
			continue
		}

		if r == '"' || r == '\'' {
			inString = true
			strChar = r
			res.WriteRune(r)
			continue
		}

		if r == '/' && i+1 < len(runes) && runes[i+1] == '/' {
			res.WriteString(string(runes[i:]))
			break
		}

		if r == ' ' || r == '\t' {
			spaceCount++
			if spaceCount == 1 {
				res.WriteRune(' ')
			}
		} else {
			spaceCount = 0
			res.WriteRune(r)
		}
	}

	resStr := strings.TrimSpace(res.String())
	resStr = strings.ReplaceAll(resStr, "}else", "} else")
	resStr = strings.ReplaceAll(resStr, "} else{", "} else {")
	return resStr
}
