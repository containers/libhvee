
# Release Notes

## 0.6.0
* Add method GetMachineExists
* Dependency updates
* Add ability to resize HyperV disks

## 0.5.0
* Improved error handling around detection of HyperV itself
* Fixed bug related to local sensitive properties
* Hushed some benign errors being emitted with logrus
* Removal of unused code and dead-end code
* Various go-lint fixes
* Added getmemory method
* Changed wql selector to find virtual machines over WMI
* Added ability for force stop of a VM
* Increased stop wait time and attempts

## 0.4.0
* Changed add-ign command to no longer require specifying the constant ignition config key prefix 
* Fixed get argument validation

## 0.3.0
* Small change in return type for error code

## 0.2
* Tweaks and export functions required for coreos ignition project

## 0.1
* Pruned unnecessary content
* Added sub-command `add-ign` to kvpctl

## v0.0.5
* Exported const and func for Ignition project
* Wait on VM to actually stop before returning

## v0.0.4
* Add bool for network creation when creating a vm

## v0.0.3
* Fix bug in processing key value pairs in Linux

## v0.0.2
* Development release
* Allow update to memory and processor counts
* Write kvp pool files same as daemon
* Code cleanup

## v0.0.1
* Development release

