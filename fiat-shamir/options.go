package fiatshamir

type transcriptConfig struct {
	withDomainSeparation bool
	fixedChallenges      []string
}

// TranscriptOption allows modifying the [Transcript] operation.
type TranscriptOption func(tc *transcriptConfig) error

// WithDomainSeparation adds domain separation string `string:` when hashing
// challenge name as defined in RCF 9380.
func WithDomainSeparation() TranscriptOption {
	return func(tc *transcriptConfig) error {
		tc.withDomainSeparation = true
		return nil
	}
}

// WithStaticChallenges fixes the allowed challenges. Otherwise challenges are
// appended when bound.
func WithStaticChallenges(challenges ...string) TranscriptOption {
	return func(tc *transcriptConfig) error {
		tc.fixedChallenges = challenges
		return nil
	}
}

func newConfig(opts ...TranscriptOption) *transcriptConfig {
	tc := &transcriptConfig{}
	for _, opt := range opts {
		opt(tc)
	}
	return tc
}
