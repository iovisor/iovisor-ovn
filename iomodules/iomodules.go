package iomodules

type IoModule interface {
	Deploy() (err error)
	Destroy() (err error)
	AttachExternalInterface(name string) (err error)
	DetachExternalInterface(name string) (err error)
	AttachToIoModule(IfaceId int, name string) (err error)
	DetachFromIoModule(name string) (err error)
}
