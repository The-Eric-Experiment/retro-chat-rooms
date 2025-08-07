package profanity

import "testing"

func TestBasicLeetSpeak(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"H3ll0 W0rld!", "hello worldi"},
		{"Th1s 1s 4 t3st", "this is a test"},
		{"l0v3 pr0gr4mm1ng", "love programming"},
		{"5up3r s3cr3t", "super secret"},
		{"numb3r5 4nd l3tt3r5", "numbers and letters"},
	}

	for _, test := range tests {
		result := normalizeText(test.input)
		if result != test.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestBasicLeetSpeakPreserveCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"H3ll0 W0rld!", "Hello Worldi"},
		{"Th1s 1s 4 t3st", "This is a test"},
		{"l0v3 pr0gr4mm1ng", "love programming"},
		{"5up3r s3cr3t", "super secret"},
		{"numb3r5 4nd l3tt3r5", "numbers and letters"},
	}

	for _, test := range tests {
		result := normalizeTextPreserveCase(test.input)
		if result != test.expected {
			t.Errorf("normalizeTextPreserveCase(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestSymbolSubstitution(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"$p3c14l ch4r4ct3r5", "special characters"},
		{"@dm1n u53r", "admin user"},
		{"h@ck3r $k1ll5", "hacker skills"},
		{"!mp0rt4nt m355@g3", "important message"},
		{"u53r|1d", "userlid"},
	}

	for _, test := range tests {
		result := normalizeText(test.input)
		if result != test.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestSymbolSubstitutionPreserveCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"$p3c14l Ch4r4ct3r5", "special Characters"},
		{"@dm1n u53r", "admin user"},
		{"H@ck3r $k1ll5", "Hacker skills"},
		{"!mp0rt4nt M355@g3", "important Message"},
		{"u53r|1d", "userlid"},
	}

	for _, test := range tests {
		result := normalizeTextPreserveCase(test.input)
		if result != test.expected {
			t.Errorf("normalizeTextPreserveCase(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestGreekCharacters(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"αβγδε", "abyde"},
		{"ήάτ", "nat"},
		{"Ωhάτ", "what"},
		{"βέτα", "beta"},
		{"γάμμα", "yamma"},
		{"Γρεεκ", "ypeek"},
		{"ΑΒΓ", "aby"},
	}

	for _, test := range tests {
		result := normalizeText(test.input)
		if result != test.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestGreekCharactersPreserveCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"αβγδε", "abyde"},
		{"ήάτ", "nat"},
		{"Ωhάτ", "What"},
		{"βέτα", "beta"},
		{"γάμμα", "yamma"},
		{"Γρεεκ", "Ypeek"},
		{"ΑΒΓ", "ABY"},
	}

	for _, test := range tests {
		result := normalizeTextPreserveCase(test.input)
		if result != test.expected {
			t.Errorf("normalizeTextPreserveCase(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestCyrillicCharacters(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Русский", "pyccknn"},
		{"Н3ЬЬ０", "hebbo"},
		{"аеор", "aeop"},
		{"РУСС", "pycc"},
	}

	for _, test := range tests {
		result := normalizeText(test.input)
		if result != test.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestCyrillicCharactersPreserveCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Русский", "PycckNN"},
		{"Н3ЬЬ０", "HeBBO"},
		{"аеор", "aeop"},
		{"РУСС", "PYCC"},
	}

	for _, test := range tests {
		result := normalizeTextPreserveCase(test.input)
		if result != test.expected {
			t.Errorf("normalizeTextPreserveCase(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestUkrainianCharacters(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Gґέέκ", "greek"},
		{"ґрек", "rpek"},
		{"Ґреція", "rpeuir"},
	}

	for _, test := range tests {
		result := normalizeText(test.input)
		if result != test.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestUkrainianCharactersPreserveCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Gґέέκ", "Greek"},
		{"ґрек", "rpek"},
		{"Ґреція", "RpeuiR"},
	}

	for _, test := range tests {
		result := normalizeTextPreserveCase(test.input)
		if result != test.expected {
			t.Errorf("normalizeTextPreserveCase(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestFullWidthCharacters(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ＦＵＬＬ－ＷＩＤＴＨ", "full-width"}, // Note: dash might not be mapped
		{"ａｂｃｄｅｆ", "abcdef"},
		{"ＡＢＣＤＥＦ", "abcdef"},
		{"１２３４５", "izeas"},
		{"０１２３４５６７８９", "oizeasgtbg"},
	}

	for _, test := range tests {
		result := normalizeText(test.input)
		if result != test.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestFullWidthCharactersPreserveCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ＦＵＬＬ－ＷＩＤＴＨ", "FULL-WIDTH"}, // Note: dash might not be mapped
		{"ａｂｃｄｅｆ", "abcdef"},
		{"ＡＢＣＤＥＦ", "ABCDEF"},
		{"１２３４５", "IZEAS"},
		{"０１２３４５６７８９", "OIZEASGTBg"},
	}

	for _, test := range tests {
		result := normalizeTextPreserveCase(test.input)
		if result != test.expected {
			t.Errorf("normalizeTextPreserveCase(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestMultiCharacterPatterns(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"tr1ck3d |<3y w0rd5 |-|3r3", "tricked key words here"},
		{"h3ll() w()rld", "hello world"},
		{"t3st [] patt3rns", "test o patterns"},
		{"us3 {} br4c3s", "use o braces"},
		{"m@k3 |\\/| patt3rns", "make m patterns"},
		{"us3 |\\| ch4rs", "use n chars"},
	}

	for _, test := range tests {
		result := normalizeText(test.input)
		if result != test.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestSuperscriptSubscript(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"5up3r5cr1pt⁴", "superscript4"},
		{"sub₅cr1pt₂", "sub5cript2"},
		{"x²⁺y²", "x2+y2"},
		{"H₂O", "h2o"},
		{"E=mc²", "e=mc2"},
	}

	for _, test := range tests {
		result := normalizeText(test.input)
		if result != test.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestMathematicalSymbols(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"∀∃∈∇∩∪∫∮", "aeevnuff"},
		{"∑∏√∝∞∠∧∨", "epva8lav"},
		{"M4th 5ymb0l5", "math symbols"},
	}

	for _, test := range tests {
		result := normalizeText(test.input)
		if result != test.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestLatinVariants(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"àëïöü", "aeiou"},
		{"ñice châractërs", "nice characters"},
		{"café résumé naïve", "cafe resume naive"},
		{"Zürich", "zurich"},
		{"piñata", "pinata"},
	}

	for _, test := range tests {
		result := normalizeText(test.input)
		if result != test.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestCasePreservation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"H3LL0 W0RLD", "HeLLo WoRLD"}, // This might not work perfectly due to string replacements
		{"Th1s Is A T3st", "This Is A Test"},
		{"ＦＵＬＬ－ＷＩＤＴＨ", "FULL-WIDTH"},
		{"Ωήάτ Άβούτ", "Wnat About"},
		{"MixedCa$e", "MixedCase"},
	}

	for _, test := range tests {
		result := normalizeTextPreserveCase(test.input)
		if result != test.expected {
			t.Errorf("normalizeTextPreserveCase(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestCzechSlovakCharacters(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Cγґίllίč", "cyrillic"},
		{"βέγоήď", "beyond"},
		{"český", "cesky"},
		{"ďalší", "dalsi"},
		{"Žeň", "zen"},
		{"řeka", "reka"},
	}

	for _, test := range tests {
		result := normalizeText(test.input)
		if result != test.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestComplexMixedInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"H3LL０ Ｗ０ＲＬＤ!", "hello worldi"},
		{"M1x3d: Н3ＬＬ０ ｗ０ＲＬＤ αήď βέγоήď!", "mixed: hello world and beyondi"}, // Some chars might not be mapped
		{"$p3c14l |<3y 4nd Gґέέκ ch4r5", "special key and greek chars"},
		{"Ωήάτ άβούτ ∀∃∈ m@th?", "wnat about aee math?"},
	}

	for _, test := range tests {
		result := normalizeText(test.input)
		if result != test.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestEmptyAndEdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"   ", "   "},
		{"123", "ize"},
		{"normal text", "normal text"},
		{"NORMAL TEXT", "normal text"},
	}

	for _, test := range tests {
		result := normalizeText(test.input)
		if result != test.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestStringReplacementOrder(t *testing.T) {
	// Test that string replacements happen before character replacements
	tests := []struct {
		input    string
		expected string
	}{
		{"|<ey", "key"},            // |< should become k, not l<
		{"|-|ello", "hello"},       // |-| should become h, not l-l
		{"|\\/|essage", "message"}, // |\\/| should become m
		{"|\\|ews", "news"},        // |\\| should become n
	}

	for _, test := range tests {
		result := normalizeText(test.input)
		if result != test.expected {
			t.Errorf("normalizeText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func BenchmarkNormalizeText(b *testing.B) {
	testString := "H3LL０ Ｗ０ＲＬＤ! Ωήάτ άβούτ Gґέέκ άήď βέγоήď $p3c14l ch4r4ct3r5?"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizeText(testString)
	}
}

func BenchmarkNormalizeTextPreserveCase(b *testing.B) {
	testString := "H3LL０ Ｗ０ＲＬＤ! Ωήάτ άβούτ Gґέέκ άήď βέγоήď $p3c14l ch4r4ct3r5?"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizeTextPreserveCase(testString)
	}
}
