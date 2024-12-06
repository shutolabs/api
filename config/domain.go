package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type RcloneConfig struct {
	Remote string
	Flags []string
}

// DomainConfig represents configuration for a specific domain
type DomainConfig struct {
	Rclone RcloneConfig
}

type DomainsConfig struct {
	Domains map[string]DomainConfig `yaml:"domains"`
}

// ConfigLoader defines the interface for loading configuration
type ConfigLoader interface {
	ReadConfig(path string) ([]byte, error)
}

// FileConfigLoader implements ConfigLoader using the file system
type FileConfigLoader struct{}

func (f *FileConfigLoader) ReadConfig(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// DomainConfigManager defines the interface for domain configuration operations
type DomainConfigManager interface {
	GetDomainConfig(domain string) (DomainConfig, error)
}

// domainConfigManagerImpl implements DomainConfigManager
type domainConfigManagerImpl struct {
	loader     ConfigLoader
	configPath string
}

// NewDomainConfigManager creates a new DomainConfigManager instance
func NewDomainConfigManager(loader ConfigLoader, configPath string) DomainConfigManager {
	return &domainConfigManagerImpl{
		loader:     loader,
		configPath: configPath,
	}
}

type MockDomainConfigManager struct {
	GetDomainConfigFunc func(domain string) (DomainConfig, error)
}

func (m *MockDomainConfigManager) GetDomainConfig(domain string) (DomainConfig, error) {
	return m.GetDomainConfigFunc(domain)
}

// loadDomainsConfig reads and parses the domains.yaml file
func (m *domainConfigManagerImpl) loadDomainsConfig() (DomainsConfig, error) {
	var config DomainsConfig
	
	data, err := m.loader.ReadConfig(m.configPath)
	if err != nil {
		return config, fmt.Errorf("error reading config file: %w", err)
	}

	// Replace environment variables in the yaml content
	content := os.ExpandEnv(string(data))

	err = yaml.Unmarshal([]byte(content), &config)
	if err != nil {
		return config, fmt.Errorf("error parsing config file: %w", err)
	}

	return config, nil
}

// GetDomainConfig retrieves configuration for a specific domain
func (m *domainConfigManagerImpl) GetDomainConfig(domain string) (DomainConfig, error) {
	config, err := m.loadDomainsConfig()
	if err != nil {
		return DomainConfig{}, err
	}

	if domainConfig, exists := config.Domains[domain]; exists {
		return domainConfig, nil
	}
	
	return DomainConfig{}, fmt.Errorf("domain config not found for: %s", domain)
} 