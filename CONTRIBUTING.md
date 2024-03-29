# How to contribute

CassKop is Apache 2.0 licensed and accepts contributions via GitHub pull requests. This document outlines some of the
conventions on commit message formatting, contact points for developers, and other resources to help get contributions
into CassKop.

## Email and Chat

You can reach developpers directly on slack : https://casskop.slack.com or by email
at prj.casskop.support@list.orangeportails.net

## Getting started

- Fork the repository on GitHub
- See the [developer guide](documentation/development.md) for build instructions.

## Reporting bugs and creating issues

Reporting bugs is one of the best ways to contribute. However, a good bug report has some very specific qualities, so
please read over our short document on [reporting bugs](documentation/reporting_bugs.md) before submitting a bug report.


## Contribution flow

This is a rough outline of what a contributor's workflow looks like:
- Create a topic branch from where to base the contribution. This is usually master.
- Make commits of logical units.
- Make sure commit messages are in the proper format (see below).
- Push changes in a topic branch to a personal fork of the repository.
- Submit a pull request to cscetbon/casskop.

The PR must receive a LGTM from two maintainers found in the [MAINTAINERS](MAINTAINERS) file.
Thanks for contributing!

## Code style

The coding style suggested by the Golang community is used in CassKop. See the [style
doc](https://github.com/golang/go/wiki/CodeReviewComments) for details.

Please follow this style to make CassKop easy to review, maintain and develop.

## Format of the commit message

We follow a rough convention for commit messages that is designed to answer two questions: what changed and why. The
subject line should feature the what and the body of the commit should describe the why.

```
scripts: add the test-cluster command

this uses tmux to setup a test cluster that can easily be killed and started for debugging.

Fixes #38
```

The format can be described more formally as follows:

```
<subsystem>: <what changed>
<BLANK LINE>
<why this change was made>
<BLANK LINE>
<footer>
```

The first line is the subject and should be no longer than 70 characters, the second line is always blank, and other
lines should be wrapped at 80 characters. This allows the message to be easier to read on GitHub as well as in various
git tools.


## Documentation

If the contribution changes the existing APIs or user interface it must include sufficient documentation to explain the
use of the new or updated feature. Likewise the CHANGELOG should be updated with a summary of the change and link to the
pull request.

