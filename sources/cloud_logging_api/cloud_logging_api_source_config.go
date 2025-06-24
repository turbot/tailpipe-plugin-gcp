package cloud_logging_api

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

type CloudLoggingAPISourceConfig struct {
	// required to allow partial decoding
	Remain   hcl.Body `hcl:",remain" json:"-"`
	LogTypes []string `hcl:"log_types,optional" json:"log_types"`
}

func (a *CloudLoggingAPISourceConfig) Validate() error {
    validLogTypes := []string{"requests"} // Currently, only "requests" is supported, but keeping this a list for future expansion

    for _, logType := range a.LogTypes {
        isValid := false
        for _, validType := range validLogTypes {
            if logType == validType {
                isValid = true
                break
            }
        }

        if !isValid {
            return fmt.Errorf("invalid log type %q, valid log types are %s", logType, strings.Join(validLogTypes, ", "))
        }
    }
    return nil
}

func (a *CloudLoggingAPISourceConfig) Identifier() string {
	return CloudLoggingAPISourceIdentifier
}
