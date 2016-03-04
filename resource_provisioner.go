package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/config"
	"github.com/hashicorp/terraform/terraform"

	clc "github.com/CenturyLinkCloud/clc-sdk"
	"github.com/CenturyLinkCloud/clc-sdk/api"
	"github.com/CenturyLinkCloud/clc-sdk/server"
	"github.com/CenturyLinkCloud/clc-sdk/status"

	"github.com/mitchellh/mapstructure"

	"github.com/kr/pretty"
)

type Provisioner struct {
	Username          string            `mapstructure:"username"`
	Password          string            `mapstructure:"password"`
	Account           string            `mapstructure:"account"`
	Package           string            `mapstructure:"package"`
	PackageParameters map[string]string `mapstructure:"parameters"`
}

type ResourceProvisioner struct{}

func (r *ResourceProvisioner) Apply(
	o terraform.UIOutput,
	s *terraform.InstanceState,
	c *terraform.ResourceConfig) error {

	log.Printf("Got to Apply for clc_exec provisioner")

	log.Printf("InstanceState looks like:")
	// fmt.Printf("%+v\n", s)
	// fmt.Printf("InstanceState: %# v", pretty.Formatter(s))
	pretty.Log(s)

	var err error

	// Decode the raw config for this provisioner
	p, err := r.decodeConfig(c)
	if err != nil {
		return err
	}

	log.Print("Provisioner looks like:")
	pretty.Log(p)

	log.Printf("Username = %s, Password = %s, Alias = %s", p.Username, p.Password, p.Account)
	log.Printf("Attempting to execute package %s", p.Package)
	log.Print("Package Parameters looks like:")
	pretty.Log(p.PackageParameters)

	server_id := s.ID
	log.Printf("Server name = %s", server_id)
	o.Output(fmt.Sprintf("Executing package '%s' on server '%s'", p.Package, server_id))

	// Create a CLC config
	config, err := api.NewConfig(p.Username, p.Password, p.Account, "")
	if err != nil {
		return fmt.Errorf("Failed to create CLC config with provided details: %v", err)
	}

	// Create a new CLC Client
	client := clc.New(config)

	// Make sure we can authentication
	err = client.Authenticate()
	if err != nil {
		return fmt.Errorf("Failed authenticated with provided credentials: %v", err)
	}

	// Create the pkg structure
	package_exec_spec := server.Package{
		ID:     p.Package,
		Params: p.PackageParameters,
	}

	// Execute the package
	// TODO: Is this a bit hacky just picking the first array entry?
	resp, err := client.Server.ExecutePackage(package_exec_spec, server_id)
	if err != nil || !resp[0].IsQueued {
		return fmt.Errorf("Failed executing package: %v", err)
	}

	// Check status
	// TODO: Is this a bit hacky just picking the first array entry?
	ok, st := resp[0].GetStatusID()
	if !ok {
		return fmt.Errorf("Failed extracting status to poll on %v: %v", resp, err)
	}
	err = waitStatus(client, st)
	if err != nil {
		return err
	}

	o.Output(fmt.Sprintf("Package %s successfully executed on %s", p.Package, server_id))

	return nil
}

func (r *ResourceProvisioner) Validate(c *terraform.ResourceConfig) (ws []string, es []error) {
	v := &config.Validator{
		Required: []string{
			"username", "password", "account", "package",
		},
		Optional: []string{
			"parameters.*",
		},
	}
	return v.Validate(c)
}

func (r *ResourceProvisioner) decodeConfig(c *terraform.ResourceConfig) (*Provisioner, error) {
	log.Printf("Got to decodeConfig for clc_exec")

	log.Print("ResourceConfig looks like:")
	pretty.Log(c)

	p := new(Provisioner)

	decConf := &mapstructure.DecoderConfig{
		ErrorUnused:      true,
		WeaklyTypedInput: true,
		Result:           p,
	}

	dec, err := mapstructure.NewDecoder(decConf)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	for k, v := range c.Raw {
		m[k] = v
	}

	for k, v := range c.Config {
		m[k] = v
	}

	if err := dec.Decode(m); err != nil {
		return nil, err
	}

	return p, nil

}

// package utility functions

func waitStatus(client *clc.Client, id string) error {
	// block until queue is processed and server is up
	poll := make(chan *status.Response, 1)
	err := client.Status.Poll(id, poll)
	if err != nil {
		return nil
	}
	status := <-poll
	log.Printf("[DEBUG] status %v", status)
	if status.Failed() {
		return fmt.Errorf("unsuccessful job %v failed with status: %v", id, status.Status)
	}
	return nil
}
