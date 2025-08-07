package profanity

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Character mapping table - maps lookalike characters to their ASCII equivalents
var characterMap = map[rune]rune{
	// Leet speak / number substitutions
	'0': 'o', '1': 'i', '2': 'z', '3': 'e', '4': 'a', '5': 's', '6': 'g', '7': 't', '8': 'b', '9': 'g',
	'@': 'a', '$': 's', '+': 't', '!': 'i', '|': 'l',

	// Unicode lookalikes - Latin variants
	'à': 'a', 'á': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'å': 'a', 'ā': 'a', 'ă': 'a', 'ą': 'a',
	'è': 'e', 'é': 'e', 'ê': 'e', 'ë': 'e', 'ē': 'e', 'ĕ': 'e', 'ė': 'e', 'ę': 'e', 'ě': 'e',
	'ì': 'i', 'í': 'i', 'î': 'i', 'ï': 'i', 'ĩ': 'i', 'ī': 'i', 'ĭ': 'i', 'į': 'i', 'ı': 'i',
	'ò': 'o', 'ó': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o', 'ø': 'o', 'ō': 'o', 'ŏ': 'o', 'ő': 'o',
	'ù': 'u', 'ú': 'u', 'û': 'u', 'ü': 'u', 'ũ': 'u', 'ū': 'u', 'ŭ': 'u', 'ů': 'u', 'ű': 'u', 'ų': 'u',
	'ý': 'y', 'ÿ': 'y', 'ỳ': 'y', 'ỹ': 'y', 'ȳ': 'y', 'ŷ': 'y',
	'ç': 'c', 'ć': 'c', 'ĉ': 'c', 'ċ': 'c', 'č': 'c',
	'ñ': 'n', 'ń': 'n', 'ņ': 'n', 'ň': 'n', 'ŋ': 'n',
	// Czech/Slovak characters
	'ď': 'd', 'ľ': 'l', 'ř': 'r', 'š': 's', 'ť': 't', 'ž': 'z',

	// Uppercase variants
	'À': 'A', 'Á': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A', 'Å': 'A', 'Ā': 'A', 'Ă': 'A', 'Ą': 'A',
	'È': 'E', 'É': 'E', 'Ê': 'E', 'Ë': 'E', 'Ē': 'E', 'Ĕ': 'E', 'Ė': 'E', 'Ę': 'E', 'Ě': 'E',
	'Ì': 'I', 'Í': 'I', 'Î': 'I', 'Ï': 'I', 'Ĩ': 'I', 'Ī': 'I', 'Ĭ': 'I', 'Į': 'I',
	'Ò': 'O', 'Ó': 'O', 'Ô': 'O', 'Õ': 'O', 'Ö': 'O', 'Ø': 'O', 'Ō': 'O', 'Ŏ': 'O', 'Ő': 'O',
	'Ù': 'U', 'Ú': 'U', 'Û': 'U', 'Ü': 'U', 'Ũ': 'U', 'Ū': 'U', 'Ŭ': 'U', 'Ů': 'U', 'Ű': 'U', 'Ų': 'U',
	'Ý': 'Y', 'Ÿ': 'Y', 'Ỳ': 'Y', 'Ỹ': 'Y', 'Ȳ': 'Y', 'Ŷ': 'Y',
	'Ç': 'C', 'Ć': 'C', 'Ĉ': 'C', 'Ċ': 'C', 'Č': 'C',
	'Ñ': 'N', 'Ń': 'N', 'Ņ': 'N', 'Ň': 'N', 'Ŋ': 'N',
	// Czech/Slovak uppercase
	'Ď': 'D', 'Ľ': 'L', 'Ř': 'R', 'Š': 'S', 'Ť': 'T', 'Ž': 'Z',

	// Cyrillic lookalikes
	'а': 'a', 'е': 'e', 'о': 'o', 'р': 'p', 'с': 'c', 'х': 'x', 'у': 'y', 'н': 'h', 'к': 'k', 'м': 'm', 'т': 't', 'і': 'i', 'ѕ': 's', 'ї': 'i',
	'А': 'A', 'Е': 'E', 'О': 'O', 'Р': 'P', 'С': 'C', 'Х': 'X', 'У': 'Y', 'Н': 'H', 'К': 'K', 'М': 'M', 'Т': 'T', 'І': 'I', 'Ѕ': 'S', 'Ї': 'I',
	'ґ': 'r', 'Ґ': 'R', 'Ь': 'b', 'и': 'N', 'й': 'N',
	// Ukranian
	'ц': 'u', 'я': 'R',

	// Uppercase Greek - CORRECTED
	'Α': 'A', 'Β': 'B', 'Δ': 'D', 'Ε': 'E', 'Ζ': 'Z', 'Η': 'H', 'Θ': 'O',
	'Ι': 'I', 'Κ': 'K', 'Λ': 'L', 'Μ': 'M', 'Ν': 'N', 'Ξ': 'X', 'Ο': 'O', 'Π': 'P',
	'Ρ': 'P', 'Σ': 'S', 'Τ': 'T', 'Υ': 'Y', 'Φ': 'F', 'Χ': 'X', 'Ψ': 'Y', 'Ω': 'W',
	// FIXED: Γ should be 'Y' not 'G'
	'Γ': 'Y',

	// Greek lookalikes - CORRECTED
	'α': 'a', 'β': 'b', 'δ': 'd', 'ε': 'e', 'ζ': 'z', 'η': 'h', 'θ': 'o',
	'ι': 'i', 'κ': 'k', 'λ': 'l', 'μ': 'm', 'ν': 'n', 'ξ': 'x', 'ο': 'o', 'π': 'p',
	'ρ': 'p', 'σ': 's', 'ς': 's', 'τ': 't', 'υ': 'u', 'φ': 'f', 'χ': 'x', 'ψ': 'y', 'ω': 'w',
	// FIXED: γ should be 'y' not 'g'
	'γ': 'y',

	// Greek with accents
	'ά': 'a', 'έ': 'e', 'ή': 'n', 'ί': 'i', 'ό': 'o', 'ύ': 'u', 'ώ': 'o',
	'ΐ': 'i', 'ΰ': 'u',

	// Uppercase Greek with accents
	'Ώ': 'W', 'Ά': 'A', 'Έ': 'E', 'Ή': 'H', 'Ί': 'I', 'Ό': 'O', 'Ύ': 'U',
	'Ϊ': 'I', 'Ϋ': 'Y',

	// Mathematical symbols and other Unicode tricks
	'∀': 'A', '∃': 'E', '∈': 'E', '∇': 'V', '∩': 'n', '∪': 'U', '∫': 'f', '∮': 'f',
	'∑': 'E', '∏': 'P', '√': 'V', '∝': 'a', '∞': '8', '∠': 'L', '∧': 'A', '∨': 'V',

	// Full-width characters
	'Ａ': 'A', 'Ｂ': 'B', 'Ｃ': 'C', 'Ｄ': 'D', 'Ｅ': 'E', 'Ｆ': 'F', 'Ｇ': 'G', 'Ｈ': 'H',
	'Ｉ': 'I', 'Ｊ': 'J', 'Ｋ': 'K', 'Ｌ': 'L', 'Ｍ': 'M', 'Ｎ': 'N', 'Ｏ': 'O', 'Ｐ': 'P',
	'Ｑ': 'Q', 'Ｒ': 'R', 'Ｓ': 'S', 'Ｔ': 'T', 'Ｕ': 'U', 'Ｖ': 'V', 'Ｗ': 'W', 'Ｘ': 'X',
	'Ｙ': 'Y', 'Ｚ': 'Z',
	'ａ': 'a', 'ｂ': 'b', 'ｃ': 'c', 'ｄ': 'd', 'ｅ': 'e', 'ｆ': 'f', 'ｇ': 'g', 'ｈ': 'h',
	'ｉ': 'i', 'ｊ': 'j', 'ｋ': 'k', 'ｌ': 'l', 'ｍ': 'm', 'ｎ': 'n', 'ｏ': 'o', 'ｐ': 'p',
	'ｑ': 'q', 'ｒ': 'r', 'ｓ': 's', 'ｔ': 't', 'ｕ': 'u', 'ｖ': 'v', 'ｗ': 'w', 'ｘ': 'x',
	'ｙ': 'y', 'ｚ': 'z', '－': '-',

	// Full-width numbers (0-9)
	'０': 'O', '１': 'I', '２': 'Z', '３': 'E', '４': 'A', '５': 'S', '６': 'G', '７': 'T', '８': 'B', '９': 'g',

	// Subscript and superscript
	'⁰': '0', '¹': '1', '²': '2', '³': '3', '⁴': '4', '⁵': '5', '⁶': '6', '⁷': '7', '⁸': '8', '⁹': '9',
	'₀': '0', '₁': '1', '₂': '2', '₃': '3', '₄': '4', '₅': '5', '₆': '6', '₇': '7', '₈': '8', '₉': '9',
	'ₐ': 'a', 'ₑ': 'e', 'ᵢ': 'i', 'ₒ': 'o', 'ᵤ': 'u',
	'ᵃ': 'a', 'ᵇ': 'b', 'ᶜ': 'c', 'ᵈ': 'd', 'ᵉ': 'e', 'ᶠ': 'f', 'ᵍ': 'g', 'ʰ': 'h',
	'ⁱ': 'i', 'ʲ': 'j', 'ᵏ': 'k', 'ˡ': 'l', 'ᵐ': 'm', 'ⁿ': 'n', 'ᵒ': 'o', 'ᵖ': 'p',
	'ʳ': 'r', 'ˢ': 's', 'ᵗ': 't', 'ᵘ': 'u', 'ᵛ': 'v', 'ʷ': 'w', 'ˣ': 'x', 'ʸ': 'y', 'ᶻ': 'z', '⁺': '+',
}

// String replacements for multi-character leet speak
var stringReplacements = map[string]string{
	"()":    "o",
	"[]":    "o",
	"{}":    "o",
	"<>":    "o",
	"|-|":   "h",
	"|<":    "k",
	"|\\/|": "m",
	"|\\|":  "n",
}

// normalizeText converts leet speak and Unicode lookalikes to ASCII (lowercase)
func normalizeText(text string) string {
	if !utf8.ValidString(text) {
		return text
	}

	// STEP 1: Handle string replacements FIRST
	result := text
	for pattern, replacement := range stringReplacements {
		result = strings.ReplaceAll(result, pattern, replacement)
	}

	// STEP 2: Handle character replacements
	var normalized strings.Builder
	for _, char := range result {
		if replacement, exists := characterMap[char]; exists {
			normalized.WriteRune(replacement)
		} else {
			normalized.WriteRune(char)
		}
	}

	return strings.ToLower(normalized.String())
}

// normalizeTextPreserveCase converts characters but preserves original case
func normalizeTextPreserveCase(text string) string {
	if !utf8.ValidString(text) {
		return text
	}

	// STEP 1: Handle string replacements FIRST
	result := text
	for pattern, replacement := range stringReplacements {
		// lowerText := strings.ToLower(result)
		if strings.Contains(result, pattern) {
			result = strings.ReplaceAll(result, pattern, replacement)
		}
	}

	// STEP 2: Handle character replacements while preserving case
	var normalized strings.Builder
	for _, char := range result {
		if replacement, exists := characterMap[char]; exists {
			if unicode.IsUpper(char) {
				normalized.WriteRune(unicode.ToUpper(replacement))
			} else {
				normalized.WriteRune(replacement)
			}
		} else {
			normalized.WriteRune(char)
		}
	}

	return normalized.String()
}
