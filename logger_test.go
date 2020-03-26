package dodod

import "testing"

func TestDefaultLogger_Debugf(t *testing.T) {
	DefaultLogger.Debugf("MSG")
}

func TestDefaultLogger_Errorf(t *testing.T) {
	DefaultLogger.Errorf("MSG")
}

func TestDefaultLogger_Infof(t *testing.T) {
	DefaultLogger.Infof("MSG")
}

func TestDefaultLogger_Warningf(t *testing.T) {
	DefaultLogger.Warningf("MSG")
}
