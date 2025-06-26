package protocol

type DataPacker interface {
	Pack() []byte
	UnPack()
}

type DefaultDataPacker struct {
	Adu AppDataUnit
	DataPacker
}

func NewDefaultDataPacker(Adu AppDataUnit) DefaultDataPacker {
	return DefaultDataPacker{
		Adu: Adu,
	}
}
