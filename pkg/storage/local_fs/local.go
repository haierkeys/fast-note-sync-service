package local_fs

type Config struct {
	IsEnabled      bool   `yaml:"is-enable"`
	HttpfsIsEnable bool   `yaml:"httpfs-is-enable"`
	IsUserEnabled  bool   `yaml:"is-user-enable"`
	SavePath       string `yaml:"save-path"`
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
