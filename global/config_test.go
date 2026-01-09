package global

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfigSave(t *testing.T) {
	// 1. 创建临时配置文件
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")

	initialConfig := config{
		WebGUI: webGUI{
			FontSet: "InitialFont",
		},
	}
	data, err := yaml.Marshal(initialConfig)
	if err != nil {
		t.Fatalf("Failed to marshal initial config: %v", err)
	}

	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("Failed to write initial config file: %v", err)
	}

	// 2. 加载配置
	absPath, _ := filepath.Abs(tmpFile)
	_, err = ConfigLoad(absPath)
	if err != nil {
		t.Fatalf("ConfigLoad failed: %v", err)
	}

	// 3. 修改配置并保存
	Config.WebGUI.FontSet = "UpdatedFont"
	if err := Config.Save(); err != nil {
		t.Fatalf("Config.Save error: %v, file: %s", err, Config.File)
	}

	// 4. 验证文件内容
	updatedData, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read updated config file: %v", err)
	}

	var updatedConfig config
	if err := yaml.Unmarshal(updatedData, &updatedConfig); err != nil {
		t.Fatalf("Failed to unmarshal updated config: %v", err)
	}

	if updatedConfig.WebGUI.FontSet != "UpdatedFont" {
		t.Errorf("Expected FontSet UpdatedFont, got %s", updatedConfig.WebGUI.FontSet)
	}
}
