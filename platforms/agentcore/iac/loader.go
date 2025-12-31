// Package iac provides shared infrastructure-as-code configuration for AgentCore deployments.
package iac

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadStackConfigFromFile loads a StackConfig from a JSON or YAML file.
// The file format is auto-detected from the extension.
func LoadStackConfigFromFile(path string) (*StackConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return LoadStackConfigFromJSON(data)
	case ".yaml", ".yml":
		return LoadStackConfigFromYAML(data)
	default:
		return nil, fmt.Errorf("unsupported file format: %s (use .json, .yaml, or .yml)", ext)
	}
}

// LoadStackConfigFromJSON parses a StackConfig from JSON data.
func LoadStackConfigFromJSON(data []byte) (*StackConfig, error) {
	var config StackConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	config.ApplyDefaults()
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// LoadStackConfigFromYAML parses a StackConfig from YAML data.
func LoadStackConfigFromYAML(data []byte) (*StackConfig, error) {
	var config StackConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	config.ApplyDefaults()
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// JSONConfigExample returns an example JSON configuration.
func JSONConfigExample() string {
	return `{
  "stackName": "my-agent-stack",
  "description": "My AgentCore deployment",
  "agents": [
    {
      "name": "primary-agent",
      "description": "Primary processing agent",
      "containerImage": "123456789.dkr.ecr.us-east-1.amazonaws.com/my-agent:latest",
      "memoryMB": 1024,
      "timeoutSeconds": 300,
      "isDefault": true,
      "environment": {
        "LOG_LEVEL": "info"
      },
      "secretsARNs": [
        "arn:aws:secretsmanager:us-east-1:123456789:secret:api-keys"
      ]
    },
    {
      "name": "secondary-agent",
      "description": "Secondary validation agent",
      "containerImage": "123456789.dkr.ecr.us-east-1.amazonaws.com/validator:latest",
      "memoryMB": 512,
      "timeoutSeconds": 60
    }
  ],
  "vpc": {
    "createVPC": true,
    "vpcCidr": "10.0.0.0/16",
    "maxAZs": 2,
    "enableVPCEndpoints": true
  },
  "observability": {
    "provider": "opik",
    "project": "my-agent-stack",
    "enableCloudWatchLogs": true,
    "logRetentionDays": 30
  },
  "iam": {
    "enableBedrockAccess": true
  },
  "tags": {
    "Environment": "production",
    "Team": "ai-platform"
  },
  "removalPolicy": "destroy"
}`
}

// YAMLConfigExample returns an example YAML configuration.
func YAMLConfigExample() string {
	return `stackName: my-agent-stack
description: My AgentCore deployment

agents:
  - name: primary-agent
    description: Primary processing agent
    containerImage: 123456789.dkr.ecr.us-east-1.amazonaws.com/my-agent:latest
    memoryMB: 1024
    timeoutSeconds: 300
    isDefault: true
    environment:
      LOG_LEVEL: info
    secretsARNs:
      - arn:aws:secretsmanager:us-east-1:123456789:secret:api-keys

  - name: secondary-agent
    description: Secondary validation agent
    containerImage: 123456789.dkr.ecr.us-east-1.amazonaws.com/validator:latest
    memoryMB: 512
    timeoutSeconds: 60

vpc:
  createVPC: true
  vpcCidr: 10.0.0.0/16
  maxAZs: 2
  enableVPCEndpoints: true

observability:
  provider: opik
  project: my-agent-stack
  enableCloudWatchLogs: true
  logRetentionDays: 30

iam:
  enableBedrockAccess: true

tags:
  Environment: production
  Team: ai-platform

removalPolicy: destroy
`
}

// WriteExampleConfig writes an example configuration file.
func WriteExampleConfig(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	var content string

	switch ext {
	case ".json":
		content = JSONConfigExample()
	case ".yaml", ".yml":
		content = YAMLConfigExample()
	default:
		return fmt.Errorf("unsupported file format: %s (use .json, .yaml, or .yml)", ext)
	}

	return os.WriteFile(path, []byte(content), 0600)
}
