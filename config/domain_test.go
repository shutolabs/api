package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockConfigLoader is a mock implementation of ConfigLoader
type MockConfigLoader struct {
    mock.Mock
}

func (m *MockConfigLoader) ReadConfig(path string) ([]byte, error) {
    args := m.Called(path)
    return args.Get(0).([]byte), args.Error(1)
}

func TestGetDomainConfig_Success(t *testing.T) {
    // Arrange
    mockLoader := new(MockConfigLoader)
    validYaml := `
domains:
  example.com:
    rclone:
      remote: "remote1"
      flags:
        - "--flag1"
        - "--flag2"
`
    mockLoader.On("ReadConfig", "config/domains.yaml").Return([]byte(validYaml), nil)
    
    manager := NewDomainConfigManager(mockLoader, "config/domains.yaml")

    // Act
    config, err := manager.GetDomainConfig("example.com")

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "remote1", config.Rclone.Remote)
    assert.Equal(t, []string{"--flag1", "--flag2"}, config.Rclone.Flags)
    mockLoader.AssertExpectations(t)
}

func TestGetDomainConfig_DomainNotFound(t *testing.T) {
    // Arrange
    mockLoader := new(MockConfigLoader)
    validYaml := `
domains:
  example.com:
    rclone:
      remote: "remote1"
      flags: []
`
    mockLoader.On("ReadConfig", "config/domains.yaml").Return([]byte(validYaml), nil)
    
    manager := NewDomainConfigManager(mockLoader, "config/domains.yaml")

    // Act
    _, err := manager.GetDomainConfig("nonexistent.com")

    // Assert
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "domain config not found for: nonexistent.com")
    mockLoader.AssertExpectations(t)
}

func TestGetDomainConfig_InvalidYAML(t *testing.T) {
    // Arrange
    mockLoader := new(MockConfigLoader)
    invalidYaml := `invalid: yaml: content`
    mockLoader.On("ReadConfig", "config/domains.yaml").Return([]byte(invalidYaml), nil)
    
    manager := NewDomainConfigManager(mockLoader, "config/domains.yaml")

    // Act
    _, err := manager.GetDomainConfig("example.com")

    // Assert
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "error parsing config file")
    mockLoader.AssertExpectations(t)
} 