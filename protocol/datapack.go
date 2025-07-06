package protocol

type DataPacker interface {
	Pack() []byte
	UnPack()
}
