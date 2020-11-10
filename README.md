## hc-swap
-------------

A simple tool to help you manage your local HashiCorp binaries.

Currently functions for Terraform, Vault, and Packer.

## Usage
-------------

Download the latest release tarball and untar the binary within your $PATH.

Execute the binary using `hc-swap`.

Follow to prompts to select your application, then the version you'd like to Install/Uninstall/Activate.

## Considerations
-------------

This app will create and retain multiple versions of HashiCorp binaries on your local machine which could pile up over time. 

It's highly advised to make full usage of the Uninstall options to remove unused versions from your system.

This app has been designed and used primarily on *nix systems, as a result there may be unforseen bugs when using the tool from a machine running Windows or another OS.

## Troubleshooting
-------------

This app creates and houses all of it's content within a directory named `hc-swap` located directly within your home directory.

Symlinks to each of the applications can be found under `/usr/local/bin` and all downloaded application versions can be located within their respective `~/hcswap/*-versions` directories.

Should you run into issues, you can cleanup the symlinks, version directories, or the entire `hc-swap` directory altogether and the next execution should rebuild everything cleanly.