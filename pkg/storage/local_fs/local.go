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

	var IsUserEnabled bool
	switch t := cf["IsUserEnabled"].(type) {
	case int64:
		if t == 0 {
			IsUserEnabled = false
		} else {
			IsUserEnabled = true
		}
	case bool:
		IsUserEnabled = t
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
		IsUserEnabled:  IsUserEnabled,
		HttpfsIsEnable: HttpfsIsEnable,
		SavePath:       cf["SavePath"].(string),
	}
	return &LocalFS{
		Config: conf,
	}, nil
}
