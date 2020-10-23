# Paul The Alien
[![Go Report Card](https://goreportcard.com/badge/github.com/Spazzy757/paul)](https://goreportcard.com/report/github.com/Spazzy757/paul)
[![codecov](https://codecov.io/gh/Spazzy757/paul/branch/main/graph/badge.svg)](https://codecov.io/gh/Spazzy757/paul)

A Github Bot that has a love for furry things and will help with day to day tasks

A huge shout out to inspiration [Derek](https://github.com/alexellis/derek) created by [Alex Ellis](https://github.com/alexellis). If you need something for production work flows, I suggest having a look and sponsoring

## Setup

If you would like to install Paul, you can find him in the [Github Apps](https://github.com/apps/paulthealien). 

**Please Note** Paul is currently in Alpha. Backwards incompatible changes can occur. There also might be times where you will need to update permissions based on newly released features.

## Configuration

Paul is configured using the `PAUL.yaml` in the root of your default branch:

```yaml
maintainers:
- Spazzy757
# Allows for the /label and /remove-label commands
# usage: /label enhancement
# usage: /remove-label enhancement
# Will only add existing labels
# Can be used on PR's or Issues
labels: true
# Settings for branch destroyer
# branch destroyer will not delete your default branch
branch_destroyer:
  enabled: true
  # set other "protected" branches here
  protected_branches:
  - main
pull_requests:
  # This will limit the amount of PR's a single contributer can have
  # Limits work in progress
  limit_pull_requests:
    max_number: 3
     # This is the message that will displayed when a user opens a pull request
  open_message: |
      Greetings! Thanks for opening a PR
  # Enables the /cat command
  cats_enabled: true
  # enables the /dog command
  dogs_enabled: true
  # Allows any maintainer in the list to run /approve
  # Paul will approve the PR (Does not merge it)
  allow_approval: true
```

## Contributing

If you would like to contribute, have a look at the [CONTRIBUTING.md](https://github.com/Spazzy757/paul/blob/main/CONTRIBUTING.md)
