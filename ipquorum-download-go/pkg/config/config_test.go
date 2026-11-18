package config

import "testing"

func validConfig() *Config {
	cfg := NewConfig()
	cfg.APIEndpoint = "10.0.0.1"
	cfg.Username = "superuser"
	cfg.Password = "secret"
	cfg.PartnerSystem = "cluster02"
	return cfg
}

func TestNewConfigDefaultPort(t *testing.T) {
	if got := NewConfig().APIPort; got != 7443 {
		t.Errorf("default APIPort = %d, want 7443", got)
	}
}

func TestValidateAPIPort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"default", 7443, false},
		{"custom", 8443, false},
		{"min", 1, false},
		{"max", 65535, false},
		{"zero", 0, true},
		{"negative", -1, true},
		{"too large", 65536, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			cfg.APIPort = tt.port
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() with port %d: err = %v, wantErr %v", tt.port, err, tt.wantErr)
			}
		})
	}
}
