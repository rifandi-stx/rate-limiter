package ratelimit

// RequestContext carries the request dimensions a rule can slice by.
// Known fields are first-class; anything the library did not anticipate
// goes in Custom so a new dimension needs no library change.
type RequestContext struct {
	UserID     string
	MerchantID string // a.k.a. org
	IP         string
	Path       string
	Headers    map[string]string
	Custom     map[string]string // escape hatch for arbitrary dimensions
}

// Field resolves a keyField/match field name to its value, checking known
// fields first, then custom. Returns ("", false) when absent.
func (c RequestContext) Field(name string) (string, bool) {
	switch name {
	case "userId":
		return c.UserID, c.UserID != ""
	case "merchantId", "orgId":
		return c.MerchantID, c.MerchantID != ""
	case "ip":
		return c.IP, c.IP != ""
	case "path":
		return c.Path, c.Path != ""
	}
	if c.Headers != nil {
		if v, ok := c.Headers[name]; ok && v != "" {
			return v, true
		}
	}
	if c.Custom != nil {
		if v, ok := c.Custom[name]; ok && v != "" {
			return v, true
		}
	}
	return "", false
}
