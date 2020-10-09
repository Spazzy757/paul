# Paul The Alien
[![Go Report Card](https://goreportcard.com/badge/github.com/Spazzy757/paul)](https://goreportcard.com/report/github.com/Spazzy757/paul)
[![codecov](https://codecov.io/gh/Spazzy757/paul/branch/main/graph/badge.svg)](https://codecov.io/gh/Spazzy757/paul)

A Github Bot that has a love for furry things and will help with day to day tasks

A huge shout out to inspiration [Derek](https://github.com/alexellis/derek) created by [Alex Ellis](https://github.com/alexellis). If you need something for production work flows, I suggest having a look and sponsoring

## Setup

If you would like to install Paul, you can find him in the [Github Apps](https://github.com/apps/paulthealien). 

**Please Note** Paul is currently in Alpha. Backwards incompatible changes can occur. There also might be times where you will need to update permissions based on newly released features.

## Configuration

Paul is configured using the `PAUL.yaml` in the root of your `main` branch (Currently master branch is not supported). You can have the following configurations:

```yaml
# List of maintainers of the repo
maintainers:
- Spazzy757
pull_requests:
  # This is the message that will displayed when a user opens a pull request
  open_message: |
    Greetings! Thanks for opening a PR
  # Enables the /cat command
  cats_enabled: true
```

## Contributing

If you would like to contribute, have a look at the [CONTRIBUTING.md](https://github.com/Spazzy757/paul/blob/main/CONTRIBUTING.md)
