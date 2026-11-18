package razorpay

// Identifier formats for ids that the Razorpay SDK places in a REST path.
// The SDK builds paths with fmt.Sprintf and does not escape them, so an id
// containing "/", "?" or "#" can reach a different endpoint. Each pattern is
// anchored and admits a single path segment only. Handlers also PathEscape
// the id before passing it to the SDK, as defence in depth.
const (
	orderIDPattern       = "^order_[A-Za-z0-9]+$"
	paymentLinkIDPattern = "^plink_[A-Za-z0-9]+$"
	payoutIDPattern      = "^pout_[A-Za-z0-9]+$"
)
