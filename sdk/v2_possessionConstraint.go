// DO NOT EDIT LOCAL SDK - USE v3 okta-sdk-golang FOR API CALLS THAT DO NOT EXIST IN LOCAL SDK
package sdk

type PossessionConstraint struct {
	AuthenticationMethods         []AuthenticationMethodObject `json:"authenticationMethods,omitempty"`
	ExcludedAuthenticationMethods []AuthenticationMethodObject `json:"excludedAuthenticationMethods,omitempty"`
	Methods                       []string                     `json:"methods,omitempty"`
	ReauthenticateIn              string                       `json:"reauthenticateIn,omitempty"`
	Types                         []string                     `json:"types,omitempty"`
	Required                      bool                         `json:"required"`
	DeviceBound                   string                       `json:"deviceBound,omitempty"`
	HardwareProtection            string                       `json:"hardwareProtection,omitempty"`
	PhishingResistant             string                       `json:"phishingResistant,omitempty"`
	UserPresence                  string                       `json:"userPresence,omitempty"`
	UserVerification              string                       `json:"userVerification,omitempty"`
	UserVerificationMethods       []string                     `json:"userVerificationMethods,omitempty"`
}

func NewPossessionConstraint() *PossessionConstraint {
	return &PossessionConstraint{}
}

func (a *PossessionConstraint) IsPolicyInstance() bool {
	return true
}
