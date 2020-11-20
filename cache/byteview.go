package cache

type byteView struct{
	b []byte
}

func (v byteView) Len() int{
	return len(v.b)
}

func (v byteView) ByteSlice() []byte{
	return cloneBytes(v.b)
}

func cloneBytes(b []byte)[]byte{
	c := make([]byte, len(b))
	copy(c, b)
	return c
}



