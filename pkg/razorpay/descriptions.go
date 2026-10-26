package razorpay

// Shared description text for amount and currency parameters. The
// {{currency}}, {{minor_unit}} and {{symbol}} placeholders are filled from
// i18nify for the active region when the toolsets are built (see
// region.Profile.Localize).
const (
	// amountUnitDesc completes a sentence such as "Payment amount ...".
	amountUnitDesc = "in the smallest unit of {{currency}} " +
		"({{minor_unit}} decimal places): 100 = {{symbol}}1"

	// amountExampleDesc gives a worked conversion for amountUnitDesc.
	amountExampleDesc = "Example: for {{symbol}}295, use 29500"

	// currencyCodeDesc describes a three-letter currency code parameter.
	currencyCodeDesc = "Three-letter ISO currency code (e.g., {{currency}})"
)
