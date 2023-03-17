# libhvee

Libhvee is a special purposed library for interacting with Microsoft's HyperV hypervisor.  It was designed with the 
[Ignition](https://coreos.github.io/ignition/) and [Podman](https://github.com/containers/podman) projects in mind
largely because Podman Machine would like to support HyperV in addition to WSL2.

The libhvee library can do the following with HyperV virtual machines:

* Create
* Start
* Stop
* Remove
* Obtain various statuses
* Add and read key-value pairs used for passing information from the host to guest virtual machines.

For an example on how to use this library, consider consulting the examples
in the [cmd dir](https://github.com/containers/libhvee/tree/main/cmd).