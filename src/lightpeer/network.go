package main

import (
	pb "github.com/stefanprisca/lightchain/src/api/lightpeer"
)

type LightNetwork struct {
	Peers []pb.PeerInfo `json:peers`
}
