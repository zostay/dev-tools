# ZX Dev Tools

These are a set of build and server running scripts.

## Why does this exist?

This is sort of an experiment, but sort of a serious attempt at a tool that
provides a single point of entry to anything else. I am not creating a dev tool
stack so much as a tool to unify access to multiple dev stacks.

We'll see how it goes.

## What does it provide?

The intent is to provide a set of simple tools, including the following:

### zxconfig

The toolkit for managing ZX's own configuration. Eventually, I would like this
to provide a peek into and possible manage the application's own configuration
as well (if I can figure out a sensible way to make that happen). 

### zxbuild

This is a tool for building and testing the project. This means checking the
application for compile time errors, running the test suite, and performing any
static analysis available to the project.

### zxinstall

This is a tool for installing the project onto the local machine, usually into
the local home directory.

### zxstart

This is a tool for running a development application server for the application.

### pingdb

And in an exercise of "which of these things does not belong," this is a little
tool I use to see if a MySQL server is ready to accept connections.
