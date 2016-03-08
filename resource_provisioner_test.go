package main

import (
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

func TestResourceProvisioner_impl(t *testing.T) {
	var _ terraform.ResourceProvisioner = new(ResourceProvisioner)
}

// builds and returns a terraform.ResourceConfig object pointer from a map of generic types
func testConfig(t *testing.T, c map[string]interface{}) *terraform.ResourceConfig {
	r, err := config.NewRawConfig(c)
	if err != nil {
		t.Fatalf("bad: %s", err)
	}

	return terraform.NewResourceConfig(r)
}

func TestResourceProvider_Validate_good(t *testing.T) {
	c := testConfig(t, map[string]interface{}{
		"username": "user",
		"password": "password",
		"account":  "account",
		"package":  "5f4e1b93dbf31747cd2ba2a9289d46d0bb9e6a9b",
	})

	r := new(ResourceProvisioner)
	warn, errs := r.Validate(c)

	if len(warn) > 0 {
		t.Fatalf("Warnings were not expected")
	}

	if len(errs) > 0 {
		t.Fatalf("Errors were not expected")
	}
}

func TestResourceProvider_Validate_missing(t *testing.T) {
	c := testConfig(t, map[string]interface{}{})
	p := new(ResourceProvisioner)
	warn, errs := p.Validate(c)
	if len(warn) > 0 {
		t.Fatalf("Warnings: %v", warn)
	}
	if len(errs) == 0 {
		t.Fatalf("Should have errors")
	}
}
