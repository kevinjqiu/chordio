package chordio

type Server struct {
	m    Rank
	bind string

	local  Node
	finger FingerTable
}

func (s *Server) Serve() error {
	return nil
}

func NewServer(config Config) (*Server, error) {
	s := Server{
		m:    config.M,
		bind: config.Bind,

		local: newNode(config.Bind, config.M),
		finger: newFingerTable(),
	}

	return &s, nil
}
