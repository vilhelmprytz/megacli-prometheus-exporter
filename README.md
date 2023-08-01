# megacli-prometheus-exporter

Prometheus exporter for MegaRAID controllers using MegaCLI. Exporter depends on MegaCli64 and `megaclisas-status` script.

## Features

- Provides metrics for the MegaRAID card to be scraped by Prometheus.
- Supports the following metrics:
  - `megacli_exporter_up`: '0' if a scrape of the MegaRAID CLI was successful, '1' otherwise.
  - `megacli_controller_information`: Constant metric with value 1 labeled with info about MegaRAID controller.
  - `megacli_array`: MegaRAID Array status, 0 for 'Optimal', 1 for anything else.
  - `megacli_disk`: MegaRAID disk status, 0 for 'Online, Spun Up' 1 for anything else.

## Prerequisites

Before using the MegaCLI Prometheus Exporter, ensure you have the following prerequisites installed:

- `MegaCli64` in `$PATH`
- [this script](https://github.com/eLvErDe/hwraid/blob/master/wrapper-scripts/megaclisas-status) in `$PATH`

## Contributors ✨

Copyright (C) 2023, Vilhelm Prytz, <vilhelm@prytznet.se>

Licensed under the [MIT license](LICENSE).

Created and maintained by [Vilhelm Prytz](https://github.com/vilhelmprytz).
