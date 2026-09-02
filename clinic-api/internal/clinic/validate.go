package clinic

import "strings"

// normalizeDocument strips every non-digit character from doc.
func normalizeDocument(doc string) string {
	var b strings.Builder
	for _, r := range doc {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isValidDocumentLength(normalized string) bool {
	return len(normalized) == 11 || len(normalized) == 14
}

func validateCreate(in CreateInput) (normalizedDocument string, err error) {
	fields := map[string]string{}

	normalized := normalizeDocument(in.Document)
	if in.Document == "" || !isValidDocumentLength(normalized) {
		fields["document"] = "must be 11 or 14 digits"
	}
	if strings.TrimSpace(in.LegalName) == "" {
		fields["legal_name"] = "must not be empty"
	}
	if strings.TrimSpace(in.TradeName) == "" {
		fields["trade_name"] = "must not be empty"
	}
	if msg, invalid := validateBanking(in.Banking); invalid {
		fields["banking"] = msg
	}

	if len(fields) > 0 {
		return "", &ValidationError{Fields: fields}
	}
	return normalized, nil
}

func validateUpdate(in UpdateInput) error {
	fields := map[string]string{}

	if in.LegalName != nil && strings.TrimSpace(*in.LegalName) == "" {
		fields["legal_name"] = "must not be empty"
	}
	if in.TradeName != nil && strings.TrimSpace(*in.TradeName) == "" {
		fields["trade_name"] = "must not be empty"
	}
	if msg, invalid := validateBanking(in.Banking); invalid {
		fields["banking"] = msg
	}

	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}

// validateBanking enforces the all-or-nothing rule: if Banking is
// provided, bank/agency/account must all be non-empty.
func validateBanking(b *Banking) (message string, invalid bool) {
	if b == nil {
		return "", false
	}
	if strings.TrimSpace(b.Bank) == "" || strings.TrimSpace(b.Agency) == "" || strings.TrimSpace(b.Account) == "" {
		return "bank, agency and account must all be present when banking is provided", true
	}
	return "", false
}
