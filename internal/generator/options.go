package generator

type ProjectOptions struct {
	ProjectName string
	ModuleName  string
	ServiceName string
	DB          string // "postgres" or "none"
	NoCache     bool
	NoDocker    bool
	NoCI        bool
	OutputDir   string
}
