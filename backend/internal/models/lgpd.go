package models

import (
	"strings"
	"unicode"
)

// MaskName applies LGPD masking to a name
// Example: "Joao Silva" -> "Jo** Si***"
func MaskName(name string) string {
	if name == "" {
		return ""
	}

	words := strings.Fields(name)
	maskedWords := make([]string, len(words))

	for i, word := range words {
		maskedWords[i] = maskWord(word)
	}

	return strings.Join(maskedWords, " ")
}

// maskWord masks a single word, keeping only the first 2 characters visible
func maskWord(word string) string {
	if word == "" {
		return ""
	}

	runes := []rune(word)
	length := len(runes)

	if length <= 2 {
		// Very short words: mask all but first character
		if length == 1 {
			return string(runes[0])
		}
		return string(runes[0]) + "*"
	}

	// Keep first 2 characters, mask the rest
	visiblePart := string(runes[:2])
	maskedPart := strings.Repeat("*", length-2)

	return visiblePart + maskedPart
}

// MaskEmail applies LGPD masking to an email
// Example: "joao.silva@email.com" -> "jo*****a@em***.com"
func MaskEmail(email string) string {
	if email == "" {
		return ""
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return maskWord(email)
	}

	localPart := parts[0]
	domainPart := parts[1]

	// Mask local part
	maskedLocal := maskEmailPart(localPart)

	// Mask domain (keep extension)
	domainParts := strings.Split(domainPart, ".")
	if len(domainParts) >= 2 {
		extension := domainParts[len(domainParts)-1]
		domain := strings.Join(domainParts[:len(domainParts)-1], ".")
		maskedDomain := maskWord(domain) + "." + extension
		return maskedLocal + "@" + maskedDomain
	}

	return maskedLocal + "@" + maskWord(domainPart)
}

// maskEmailPart masks an email local part, keeping first and last characters
func maskEmailPart(part string) string {
	if part == "" {
		return ""
	}

	runes := []rune(part)
	length := len(runes)

	if length <= 2 {
		return string(runes[0]) + "*"
	}

	if length <= 4 {
		return string(runes[0]) + strings.Repeat("*", length-2) + string(runes[length-1])
	}

	return string(runes[:2]) + strings.Repeat("*", length-3) + string(runes[length-1])
}

// MaskProntuario partially masks a patient record number
// Example: "PRO12345" -> "PRO1****"
func MaskProntuario(prontuario string) string {
	if prontuario == "" {
		return ""
	}

	runes := []rune(prontuario)
	length := len(runes)

	if length <= 4 {
		return prontuario
	}

	// Keep first 4 characters, mask the rest
	return string(runes[:4]) + strings.Repeat("*", length-4)
}

// MaskCPF masks a CPF number
// Example: "123.456.789-10" -> "***.***.***-10"
func MaskCPF(cpf string) string {
	if cpf == "" {
		return ""
	}

	// Remove formatting
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, cpf)

	if len(cleaned) != 11 {
		return maskWord(cpf)
	}

	// Keep only last 2 digits
	return "***.***.***-" + cleaned[9:]
}

// SanitizeForLog removes sensitive data from strings for logging purposes
func SanitizeForLog(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	sensitiveFields := map[string]bool{
		"nome_paciente":  true,
		"nome":           true,
		"email":          true,
		"cpf":            true,
		"prontuario":     true,
		"password":       true,
		"password_hash":  true,
		"dados_completos": true,
	}

	for key, value := range data {
		if sensitiveFields[key] {
			switch v := value.(type) {
			case string:
				if key == "email" {
					sanitized[key] = MaskEmail(v)
				} else if key == "cpf" {
					sanitized[key] = MaskCPF(v)
				} else if key == "prontuario" {
					sanitized[key] = MaskProntuario(v)
				} else if key == "password" || key == "password_hash" {
					sanitized[key] = "[REDACTED]"
				} else {
					sanitized[key] = MaskName(v)
				}
			default:
				sanitized[key] = "[REDACTED]"
			}
		} else {
			sanitized[key] = value
		}
	}

	return sanitized
}
