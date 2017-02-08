# IOModules

This folder contains the different IOModules that are implemented.
Each IOModule is composed of two source files, the `xxx.go` file contains the datapath implementation while the `xxxAPI.go` contains the management functions.

## IOModules API

Inside each IOModule package there is defined the `Create()` function that returns an instance of such module. That instance implements the functions defined in the `IOModule` interface (defined in `iomodules.go`) that can be used to perform actions over the module:

- **GetModuleId()**
Returns the ID assgined to the module by Hover.

- **Deploy()**
Loads the IOModule into the system.

- **Destroy()**
Unloads the IOModule.
Before calling this function all the links to external interfaces and other modules should be removed.

- **AttachExternalInterface(name string)**
Attaches the IOModule to a given external interface.

- **DetachExternalInterface(name string)**
Dettaches the IOModule from a network interface.

- **AttachToIoModule(IfaceId int, name string)**
This function is used by a upper layer to create a link between two IOModules. The creation of the link in this case is responsability of the upper layer, the module should only configure its internal data structures.

- **DetachFromIoModule(name string)**
Removes the connection to another IOModule.

- **Configure(conf interface{})**
Performs the configuration of the parameters of the IOModule.
The data structure to be passed is defined by each IOModule.

Additional to those functions, the `iomodules` package defines:

- **AttachIoModules()**
Creates a link between the two modules and then calls the `AttachToIoModule()` function on each module to update its internal sctructures.

- **DetachIoModules()**
TBD

#How to use

IOModules can be deployed in two different ways:

  * Using **IOModules APIs**, and importing the correspondent package it's possible to write your own program that exploit the APIs to Deploy, Destroy, Attach, Configure an IOModule.

  * Using **iovisorovnd in standalone mode** it's possible to deploy a single or a chain of IOModules using a YAML configuration file.  
  Please read [README_STANDALONE](./../README_STANDALONE.md#how-to-use).
