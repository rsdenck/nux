# WireGuard (wg)

WireGuard VPN management com suporte a wg-quick, wgcf (Cloudflare Warp) e wgctrl.

## Usage

```
nux wg <subcommand> [args]
```

## Subcommands

### status
Show WireGuard interface status from kernel via wgctrl.
```
nux wg status
```

### list
List all WireGuard interfaces.
```
nux wg list
```

### connect
Connect WireGuard interface via wg-quick.
```
nux wg connect [config-file]
```

### disconnect
Disconnect WireGuard interface via wg-quick.
```
nux wg disconnect [interface]
```

### show
Show raw WireGuard interface configuration.
```
nux wg show [interface]
```

### genkey
Generate WireGuard keypair (private + public).
```
nux wg genkey
```

### genpsk
Generate WireGuard pre-shared key.
```
nux wg genpsk
```

### install
Install WireGuard tools (wg, wg-quick, wgcf).
```
nux wg install
```

### quick-status
Show wg-quick managed interface status via systemd.
```
nux wg quick-status
```

### warp
Cloudflare Warp management via wgcf.

Subcommands: generate, register, connect, disconnect, status.
```
nux wg warp generate
nux wg warp register
nux wg warp connect
nux wg warp disconnect
nux wg warp status
```

## Examples

List all WireGuard interfaces:
```
nux wg list
```

Connect a WireGuard interface:
```
nux wg connect wg0
```

Disconnect:
```
nux wg disconnect wg0
```

Generate a keypair:
```
nux wg genkey
```

Register and connect Cloudflare Warp:
```
nux wg warp register
nux wg warp connect
```

Check Warp status:
```
nux wg warp status
```
