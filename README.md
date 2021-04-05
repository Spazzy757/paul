# Paul The Alien

[![Go Report Card](https://goreportcard.com/badge/github.com/Spazzy757/paul)](https://goreportcard.com/report/github.com/Spazzy757/paul)
[![codecov](https://codecov.io/gh/Spazzy757/paul/branch/main/graph/badge.svg)](https://codecov.io/gh/Spazzy757/paul)

A Github Bot that has a love for furry things and will help with day to day tasks

A huge shout out to inspiration [Derek](https://github.com/alexellis/derek) created by [Alex Ellis](https://github.com/alexellis). If you need something for production work flows, I suggest having a look and sponsoring

## Setup

If you would like to install Paul, you can find him in the [Github Apps](https://github.com/apps/paul-the-alien).

**Please Note** Paul is currently in Alpha. Backwards incompatible changes can occur. There also might be times where you will need to update permissions based on newly released features.

## Usage

### PR's and Issues

Commands:

- `/approve`: Paul will approve a Pull Request (conditions: must be a maintainer in PAUL.yaml)
- `/merge`: Paul will merge the Pull Request (conditions: must be a maintainer in PAUL.yaml)
- `/label <some-label>`: Paul will label the issue/PR with that label (conditions: must be maintainer and label must exists)
- `/remove-label <some-label>`: Paul will remove a label from a issue/PR (conditions: must be maintainer in PAUL.yaml and label must exists)
- `/dog`: Paul will add and image of a dog
- `/cat`: Paul will add an Image of a cat
- `/giphy <some description>`: Paul will fetch a giphy that matches the description and add it to the PR/Issue (only single word descriptions are currently supported)

Other Functions:

- Branch Destroyer: Will delete a branch when it has been merged (conditions: won't delete default branch or any protected branch, see configuration)
- New PR Message: Paul will post a review message when a new PR is created (condition: wont post message if maintainer opens PR)
- Pull Request Limiter: Paul will close PR's for a user if they have more than x amount of pull requests already open (see configuration). This will limit the amount of **Work In Progress**
- Empty Pull Requests: Does not allow Empty Descriptions, two levels, enforced means Paul will close the Pull Request with a message, without enforced Paul will just send a review saying to add a description
- Label Stale Pull Requests: This setting will mark Pull Requests stale if they have not been updated within the specified days
- Automated Merging of Pull Requests: Any pull request labeled with `merge` will be automatically merged every hour if they are mergeable. This means that you can mark a Pull Requests as mergeable before all required checks have passed and once they have passed Paul will merge the Pull Request

## Configuration

Paul is configured using the `PAUL.yaml` in the `.github/` directory of your default branch:

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
  # The Setting to enable automaed merges
  automated_merge: true
  # The time in days after a PR should be labeled inactive
  stale_time: 15
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
  # enables the /giphy command
  giphy_enabled: true
```

## Contributing

If you would like to contribute, have a look at the [CONTRIBUTING.md](https://github.com/Spazzy757/paul/blob/main/CONTRIBUTING.md)
