---
description: This section describes the configuration parameters and their types for your Hornet node.
image: /img/Banner/banner_hornet_configuration.png
keywords:
- IOTA Node 
- Hornet Node
- Configuration
- JSON
- Customize
- Config
- reference
---


# Core Configuration

![Hornet Node Configuration](/img/Banner/banner_hornet_configuration.png)

Hornet uses a JSON standard format as a config file. If you are unsure about JSON syntax, you can find more information in the [official JSON specs](https://www.json.org).

The default config file is `config.json`. You can change the path or name of the config file by using the `-c` or `--config` argument while executing `hornet` executable.

For example:
```bash
hornet -c config_example.json
```

You can always get the most up-to-date description of the config parameters by running:

```bash
hornet -h --full
```

