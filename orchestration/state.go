package orchestration

// State represents a generic workflow state.
// Implementations should embed this and add their own fields.
type State struct {
	// StepName tracks the current step in the workflow.
	StepName string `json:"step_name,omitempty"`

	// Error stores any error encountered during processing.
	Error string `json:"error,omitempty"`

	// Metadata stores arbitrary key-value pairs.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewState creates a new state with the given step name.
func NewState(stepName string) *State {
	return &State{
		StepName: stepName,
		Metadata: make(map[string]interface{}),
	}
}

// SetMetadata sets a metadata value.
func (s *State) SetMetadata(key string, value interface{}) {
	if s.Metadata == nil {
		s.Metadata = make(map[string]interface{})
	}
	s.Metadata[key] = value
}

// GetMetadata gets a metadata value.
func (s *State) GetMetadata(key string) (interface{}, bool) {
	if s.Metadata == nil {
		return nil, false
	}
	v, ok := s.Metadata[key]
	return v, ok
}

// SetError sets the error message.
func (s *State) SetError(err error) {
	if err != nil {
		s.Error = err.Error()
	}
}

// HasError returns true if there is an error.
func (s *State) HasError() bool {
	return s.Error != ""
}

// QualityDecision represents a quality gate decision in a workflow.
type QualityDecision struct {
	// Passed indicates if the quality check passed.
	Passed bool `json:"passed"`

	// Score is the quality score (0-100).
	Score int `json:"score"`

	// Target is the target score.
	Target int `json:"target"`

	// Shortfall is how many points short of the target.
	Shortfall int `json:"shortfall"`

	// Message provides a human-readable explanation.
	Message string `json:"message"`
}

// NewQualityDecision creates a new quality decision.
func NewQualityDecision(score, target int) *QualityDecision {
	qd := &QualityDecision{
		Score:  score,
		Target: target,
		Passed: score >= target,
	}

	if !qd.Passed {
		qd.Shortfall = target - score
	}

	return qd
}
