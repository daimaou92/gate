package gate

type ParamIn int

const (
	PARAM_IN_INVALID ParamIn = iota
	PARAM_IN_PATH
	PARAM_IN_QUERY
)

func (pi ParamIn) String() string {
	switch pi {
	case PARAM_IN_PATH:
		return "in"
	case PARAM_IN_QUERY:
		return "query"
	}
	return ""
}

type Param interface {
	name() string
	in() ParamIn
	required() bool
}

type Params interface {
	Params() []Param
}
