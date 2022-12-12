package fiatshamir

import "hash"

type Settings struct {
	Transcript         Transcript
	Prefix             string
	BaseChallenges     [][]byte
	Hash               hash.Hash
	TranscriptProvided bool
}

func WithTranscript(transcript Transcript, prefix string, baseChallenges ...[]byte) Settings {
	return Settings{
		Transcript:         transcript,
		Prefix:             prefix,
		TranscriptProvided: true,
	}
}

func WithBaseChallenge(hash hash.Hash, baseChallenges [][]byte) Settings {
	return Settings{
		BaseChallenges:     baseChallenges,
		Hash:               hash,
		TranscriptProvided: false,
	}
}
