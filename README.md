# fah-collector-exporter

Prometheus exporter for [fah-collector](https://github.com/ebiiim/fah-collector).

## Usage

### Version 1

Requirements: `fah-collector` versions in [2.0, 2.1]

#### Usage

`./fah-collector-exporter [-addr] [-insecure] COLLECTOR_VIEWER_URL`

- `COLLECTOR_VIEWER_URL`
  - `fah-collector` v2: `HTTP(S)://{COLLECTOR_ADDRESS}/all`
- `addr`: Listening address. Default is `:8080`.
- `insecure`: Skip TLS validation while accessing `fah-collector`.
