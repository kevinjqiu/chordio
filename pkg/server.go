package pkg

type Server struct {
	config Config
}

func (s *Server) Serve() error {
	return nil
}

func NewServer(config Config) (*Server, error) {
	s := Server{config: config}

	return &s, nil
}
