package pchome

import "testing"

const (
	Email = ""
	Password = ""
)

func TestDoGetKey(t *testing.T) {
	if _, err := NewConfigService().DoGetKey(Email, Password); err != nil {
		t.Error(err.Error())
	} else {
		t.Log("Authentication testing passed.")
	}
}

func TestUpdateZones(t *testing.T) {
	cs := NewConfigService()
	key, err := cs.DoGetKey(Email, Password)
	if err != nil {
		t.Fatal(err.Error())
	}

	if _, err = cs.UpdateZones(NewService(key)); err != nil {
		t.Fatal(err.Error())
	} else {
		t.Log("Updating zones passed.")
	}
}
