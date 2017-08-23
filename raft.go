package main

import (
    "github.com/coreos/etcd/raft"
    "github.com/coreos/etcd/raft/raftpb"
    "math"
    "time"
)
    
const hb = 1

type node struct {
	id uint64

	raft raft.Node

    cfg *raft.Config

    pstore map[string]string
}

func newNode(id int, peers []raft.Peer)*node(
    n := &node{
        id:id,
        cfg: &raft.Config{
            ID: uint64(id),
            ElectionTick: 10*hb,
            HeartbeatTick: hb,
            Storage: raft.NewMemoryStorage(),
            MaxSizePerMsg: math.MaxUint16,
            MaxInflightMsgs: 256,
        },
        pstore: make(map[string]string),
    }
    
    n.raft = raft.StartNode(n.cfg,peers)
    
    return n
)

func (n *node) run(){
    n.ticker = time.Tick(time.Second)
    for{
        select{
            case <- n.ticker:
                n.raft.Tick()
            case rd := <- n.raft.Ready():
                n.SaveToStorage(rd.HardState,rd.Entries,rd.Snapshot)
                n.Send(rd.Messages)
                if !raft.IsEmptySnap(rd.Snapshot){
                    n.processSnapshot(rd.Snapshot)
                }
                for _, entry := range rd.CommittedEntries{
                    n.process(entry)
                    if entry.Type == raftpb.EntryConfChange{
                        var cc raftpb.ConfChange
                        cc.Unmarshal(entry.Data)
                        n.node.ApplyConfChange(cc)
                    }
                }
                n.raft.Advance()
            case <- n.done:
                return
        }
    }
}