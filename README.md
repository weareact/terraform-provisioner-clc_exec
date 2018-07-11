# terraform-provisioner-clc_exec

A Terraform provisioner for CenturyLinkCloud (CLC) packages. 

## Installation

1. Download the plugin from the releases tab - https://github.com/fatmcgav/terraform-provisioner-clc_exec/releases/download/0.0.1/terraform-provisioner-clc_exec-0.0.1.tar.gz
2. Extract it somewhere on your PATH or note the full path to the binary.
3. Create or modify your `~/.terraformrc` file. You will need to add the `clc_exec` path within the `provisioner` section. E.g:
   ```
    provisioners {
      clc_exec = "~/path/to/terraform-provisioner-clc_exec"
    }
   ```
4. If using the make file to download dependencies rather than a dependency manager (see godeps file) switch the checkout branch of the hashicorp/terraform repository to the v0.9.9 tag prior to running make.

## Quick start

In order to execute a CLC package, you need to know the Package UUID, and any parameters that the package requires. 

```
  provisioner "clc_exec" {
    username = "[clc username]"
    password = "[clc password]"
    account = "[clc account alias]"
    package = "[clc pacakge uuid]"
    parameters = {
      "parameter1" = "value"
      "parameter2" = "value2"
    }
  }
```

## Usage

### Provisioner Configuration

`clc_exec`

Executes a CLC package as part of the resource provisioning process. 

TODO: Add table of params

## Known Issues

Currently it's necessary to provider the CLC authentication details to the provisioner in addition to the provider. 

## Contribution

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
