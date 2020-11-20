package cache
import pb "distributed_cache/geecachepb"
type PeerPicker interface {
	PickPeer(key string)(PeerGetter, bool)
}

type PeerGetter interface {
	//Get(group string, key string)([]byte, error)
	GetRPC(in *pb.Request, out *pb.Response) error
}



