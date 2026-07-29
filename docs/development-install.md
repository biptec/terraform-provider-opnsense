# Install from source

The provider is not yet distributed through Terraform Registry releases. For evaluation and development, build it from source and use a Terraform or OpenTofu development override.

## Prerequisites

- Go 1.25.8 or later
- GNU Make
- Terraform or OpenTofu
- A dedicated test OPNsense instance with API access

Do not test unreleased provider builds against a production firewall.

## Build

```sh
git clone https://github.com/biptec/terraform-provider-opnsense.git
cd terraform-provider-opnsense
gmake build-dev
```

The provider binary is written to the absolute `build` directory inside the repository. Override the output directory when needed:

```sh
gmake build-dev DEV_DIR=/absolute/path/to/provider-build
```

## Configure the development override

Create a CLI configuration file such as `/tmp/opnsense-dev.tfrc`:

```hcl
provider_installation {
  dev_overrides {
    "biptec/opnsense" = "/absolute/path/to/terraform-provider-opnsense/build"
  }

  direct {}
}
```

Export its path before running Terraform or OpenTofu:

```sh
export TF_CLI_CONFIG_FILE=/tmp/opnsense-dev.tfrc
```

Use an absolute path in `dev_overrides`. Run `plan` directly while the development override is active; `init` is unnecessary for the overridden provider.

## Configure credentials

Keep credentials outside Terraform configuration and state:

```sh
export OPNSENSE_URI="https://router.example.net"
export OPNSENSE_API_KEY="..."
export OPNSENSE_API_SECRET="..."
```

For a self-signed certificate in an isolated test environment only:

```sh
export OPNSENSE_ALLOW_INSECURE=true
```

Protect any file containing credentials with mode `0600`, and do not commit it.

## Verify the installation

The getting-started example performs a read-only interface query:

```sh
cd examples/getting-started
tofu plan
# or: terraform plan
```

A successful plan lists the interface devices returned by OPNsense and makes no changes. Run `gmake check` before testing a modified provider build.
