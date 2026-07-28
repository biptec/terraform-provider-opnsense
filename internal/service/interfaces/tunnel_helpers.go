package interfaces

import (
	"fmt"
	"net/netip"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func validateTunnelAddresses(remote, tunnelLocal, tunnelRemote types.String, prefix types.Int64) error {
	parseOptional := func(name string, value types.String) (netip.Addr, bool, error) {
		if value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
			return netip.Addr{}, false, nil
		}
		addr, err := netip.ParseAddr(value.ValueString())
		if err != nil {
			return netip.Addr{}, false, fmt.Errorf("invalid %s: %w", name, err)
		}
		return addr, true, nil
	}
	if _, _, err := parseOptional("remote_address", remote); err != nil {
		return err
	}
	localAddr, localSet, err := parseOptional("tunnel_local_address", tunnelLocal)
	if err != nil {
		return err
	}
	remoteAddr, remoteSet, err := parseOptional("tunnel_remote_address", tunnelRemote)
	if err != nil {
		return err
	}
	if localSet != remoteSet {
		return fmt.Errorf("tunnel_local_address and tunnel_remote_address must be configured together")
	}
	if localSet && localAddr.BitLen() != remoteAddr.BitLen() {
		return fmt.Errorf("tunnel_local_address and tunnel_remote_address must use the same address family")
	}
	if localSet && !prefix.IsNull() && !prefix.IsUnknown() && localAddr.Is4() && prefix.ValueInt64() > 32 {
		return fmt.Errorf("tunnel_remote_prefix must not exceed 32 for IPv4 tunnel addresses")
	}
	return nil
}
