# fnOS (飞牛 fnOS) native packaging

This directory packages `fast-note-sync-service` as a native **飞牛 fnOS**
application (`.fpk`), installable via the fnOS App Center.

## Layout

```
packaging/fnos/
├── manifest              # fnOS app metadata (version/platform are stamped at build)
├── config/
│   ├── privilege         # run as a dedicated package user (fastnotesync)
│   └── resource          # no extra system resources
├── cmd/                  # lifecycle scripts: start/stop/status/restart + hooks
├── app/
│   └── ui/
│       ├── config        # desktop entry: port-service on :9000 (iframe)
│       └── images/       # icon_64.png / icon_256.png
├── ICON.PNG, ICON_256.PNG
├── wizard/               # (reserved)
├── scripts/gen_icons.py  # regenerate the icons from the project SVG
├── build.sh              # build binary from source + package .fpk
└── .gitignore
```

## Runtime model

- **Access**: port service on `service_port=9000`. The service ships its own web
  admin panel and auth, so it does not integrate with the NAS login gateway.
- **Persistence**: `config/config.yaml`, the SQLite database and uploads all
  live in `TRIM_PKGVAR` (`/vol*/@appdata/fastnotesync`), surviving restart and
  upgrade. On first start the service auto-generates `config.yaml` with a
  random `auth-token-key`.
- **Privilege**: dedicated package user `fastnotesync` (`run-as=package`).
- **Architecture**: the binary is arch-specific, so two packages are produced —
  `x86` (amd64) and `arm` (arm64).

## Build

From the repository root:

```bash
packaging/fnos/build.sh amd64    # -> fastnotesync-<ver>-x86.fpk
packaging/fnos/build.sh arm64    # -> fastnotesync-<ver>-arm.fpk
```

`build.sh` builds the Go binary from source (the frontend is embedded via
`//go:embed`), stamps `version`/`platform` in `manifest` from
`internal/app/version.go`, downloads the official `fnpack` tool (pinned by
SHA256), and runs `fnpack build`.

## CI

`.github/workflows/build-fnos-fpk.yml` builds both architectures when a release
is published and attaches the `.fpk` files (plus `SHA256SUMS`) to that release.

## First-run security note

Upstream defaults to `user.register-is-enable: true`. On fnOS the service
listens on a LAN/public port, so after install register the first admin, set
`user.admin-uid`, then turn `register-is-enable` off. See the project README.
