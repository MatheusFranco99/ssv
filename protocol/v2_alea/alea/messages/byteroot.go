package messages


type ByteRoot struct {
	Value []byte
}


func NewByteRoot(root []byte) *ByteRoot {
	return &ByteRoot{
		Value: root,
	}
}

func (b *ByteRoot) GetRoot() ([]byte,error) {
	return b.Value,nil
}