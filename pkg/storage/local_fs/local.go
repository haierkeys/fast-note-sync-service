package local_fs

type Config struct {
	CustomPath string `yaml:"custom-path"`
	SavePath   string `yaml:"save-path"`
}

type LocalFS struct {
	IsCheckSave bool
	Config      *Config
}

func NewClient(conf *Config) (*LocalFS, error) {
	return &LocalFS{
		Config: conf,
	}, nil
}
