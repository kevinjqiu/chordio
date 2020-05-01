package pkg

type (
	ChordID uint64
	Rank    uint32 // otherwise known as the m value

	Config struct {
		M    Rank
		Bind string
	}
)
