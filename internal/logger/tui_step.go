package logger

// tuiStepLogger implements StepLogger for TUI mode.
type tuiStepLogger struct {
	logger *tuiLogger
}

func (s *tuiStepLogger) Done() {
	s.logger.send(stepDoneMsg{})
}

func (s *tuiStepLogger) Fail(err error) {
	s.logger.send(stepFailMsg{err: err})
}

// noopStepLogger is returned when log level is below threshold.
type noopStepLogger struct{}

func (s *noopStepLogger) Done()        {}
func (s *noopStepLogger) Fail(_ error) {}
