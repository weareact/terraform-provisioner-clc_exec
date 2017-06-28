package main

import (
	"context"
	"fmt"
	"io"
	"log"

	clc "github.com/CenturyLinkCloud/clc-sdk"
	"github.com/CenturyLinkCloud/clc-sdk/api"
	"github.com/CenturyLinkCloud/clc-sdk/server"
	"github.com/CenturyLinkCloud/clc-sdk/status"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/go-linereader"
)

func Provisioner() terraform.ResourceProvisioner {
	return &schema.Provisioner{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CLC_USERNAME", nil),
			},

			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CLC_PASSWORD", nil),
			},

			"account": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CLC_ACCOUNT", nil),
			},

			"package": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"parameters": &schema.Schema{
				Type:     schema.TypeMap,
				Required: true,
			},
		},

		ApplyFunc: applyFn,
	}
}

// Apply executes the CLC Provisioer
func applyFn(ctx context.Context) error {
	data := ctx.Value(schema.ProvConfigDataKey).(*schema.ResourceData)
	instanceState := ctx.Value(schema.ProvRawStateKey).(*terraform.InstanceState)
	o := ctx.Value(schema.ProvOutputKey).(terraform.UIOutput)

	o.Output("Executing applyFn")

	// Get required config items
	username := data.Get("username").(string)
	password := data.Get("password").(string)
	account := data.Get("account").(string)
	packageUUID := data.Get("package").(string)
	o.Output(fmt.Sprintf("Username = %s, Password = %s, PackageUUID = %s", username, password, packageUUID))

	// Need to know serverID to use later
	serverID := instanceState.ID
	o.Output(fmt.Sprintf("serverID = %s", serverID))

	// Get any package parameters
	parametersRaw := data.Get("parameters").(map[string]interface{})

	// Process parametersRaw
	parameters := make(map[string]string)
	for k, v := range parametersRaw {
		parameters[k] = v.(string)
	}
	o.Output(fmt.Sprintf("ParametersRaw = %+v, Processed = %+v", parametersRaw, parameters))

	// Create a CLC config
	config, err := api.NewConfig(username, password, account, "")
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
		ID:     packageUUID,
		Params: parameters,
	}

	// Execute the package
	// TODO: Is this a bit hacky just picking the first array entry?
	resp, err := client.Server.ExecutePackage(package_exec_spec, serverID)
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

	// Got here, so it must have worked! :)
	o.Output(fmt.Sprintf("Package %s successfully executed on %s", packageUUID, serverID))

	return nil
}

func copyOutput(o terraform.UIOutput, r io.Reader, doneCh chan<- struct{}) {
	defer close(doneCh)
	lr := linereader.New(r)
	for line := range lr.Ch {
		o.Output(line)
	}
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
