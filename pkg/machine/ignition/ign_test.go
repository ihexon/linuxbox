package ignition

import "testing"

func TestNewDynamicIgnition(t *testing.T) {
	ign := &DynamicIgnition{}

	cfg, err := ign.generateSetTimeZoneIgnitionCfg()
	if err != nil {
		t.Errorf(err.Error())
	}
	cmds := cfg.GetDynamicIgnition()
	for _, cmd := range cmds {
		t.Log(cmd)

	}
}
