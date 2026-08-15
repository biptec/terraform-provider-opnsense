# Changelog

## Unreleased

### Added

- OPNsense Core HA synchronization resource/data source for pfsync and optional XMLRPC configuration synchronization.
- HAProxy settings, service status/configtest, and L4/SNI frontend, backend, server, healthcheck, ACL, and action resources/data sources.
- Shared CARP VHID support for IP Alias virtual IPs so additional service addresses can fail over with a parent CARP VIP.
- Guarded DNS service cutover resource that coordinates BIND, Unbound, and the dnsmasq DNS port with runtime verification and rollback.
- BIND primary-domain transfer attachment resource for downstream states to own TSIG/NOTIFY integration without owning the primary zone.

### Changed

- BIND listener sets reject explicit empty values because os-bind requires at least one IPv4 and one IPv6 configuration value; use `::1` with IPv6 execution disabled.
- Provider acceptance CI pins the current BIND and API extensions plugin commit and runs DNS cutover tests in the BIND shard.

### Fixed

- Interface assignment creation now waits for newly registered devices to appear in OPNsense assignment options, preventing fresh loopback deployments from requiring a second apply.

## v0.25.0 - 2026-07-28

First release maintained by the current project team after the upstream v0.24.0 baseline.

### Added

- Complete Core Interfaces resources and data sources, including assignments, bridges, GIF, GRE, LAGG, loopback, neighbours, settings, VIP, VLAN, and VXLAN.
- Routing gateway and gateway-group resources and data sources.
- Read-only routing gateway status data source.
- Firewall interface-group and NPTv6 resources and data sources.
- Source-build installation workflow and getting-started example for Terraform and OpenTofu.
- Disposable OPNsense acceptance topology with deterministic routing prerequisites.

### Fixed

- Interface API serialization and state reconciliation across OPNsense 26.x.
- Optional gateway-tier state stability.
- Kea DHCPv6 and static-route acceptance prerequisites.
- CI credential masking and generated documentation checks.
