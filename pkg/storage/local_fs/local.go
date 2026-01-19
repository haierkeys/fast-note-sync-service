package local_fs

type Config struct {
	IsEnabled      bool   `yaml:"is-enable" default:"true"`
	HttpfsIsEnable bool   `yaml:"httpfs-is-enable" default:"true"`
	SavePath       string `yaml:"save-path" default:"storage/uploads"`
	CustomPath     string `yaml:"custom-path"`
}

type LocalFS struct {
	IsCheckSave bool
	Config      *Config
}

func NewClient(cf map[string]any) (*LocalFS, error) {

	var IsEnabled bool
	switch t := cf["IsEnabled"].(type) {
	case int64:
		if t == 0 {
			IsEnabled = false
		} else {
			IsEnabled = true
		}
	case bool:
		IsEnabled = t
	}

	var HttpfsIsEnable bool
	switch t := cf["HttpfsIsEnable"].(type) {
	case int64:
		if t == 0 {
			HttpfsIsEnable = false
		} else {
			HttpfsIsEnable = true
		}
	case bool:
		HttpfsIsEnable = t
	}

	conf := &Config{
		IsEnabled:      IsEnabled,
		CustomPath:     cf["CustomPath"].(string),
		HttpfsIsEnable: HttpfsIsEnable,
		SavePath:       cf["SavePath"].(string),
	}
	return &LocalFS{
		Config: conf,
	}, nil
}
