package main

import (
    // "errors"
    "goim/libs/define"
    "goim/libs/proto"
    "sync"
    // "time"

    log "github.com/thinkboy/log4go"
)

type Bucket struct {
    bLock             sync.RWMutex
    server            int // session server map init num
    session           int // bucket session init num
    sessionMgr        *SessionMgr
    roomCounter       map[int32]int32           // roomid->count
    serverCounter     map[int32]int32           // server->count
    userServerCounter map[int32]map[int64]int32 // serverid->userid count
}

// NewBucket new a bucket struct. store the subkey with im channel.
func NewBucket(session, server, cap int) *Bucket {
    b := new(Bucket)
    b.sessionMgr = NewSessionMgr(cap)
    b.roomCounter = make(map[int32]int32)
    b.serverCounter = make(map[int32]int32)
    b.userServerCounter = make(map[int32]map[int64]int32)
    b.server = server
    b.session = session
    return b
}

// counter incr or decr counter.
func (b *Bucket) counter(userId int64, server int32, roomId int32, incr bool) {
    var (
        sm  map[int64]int32
        v   int32
        ok  bool
    )
    if sm, ok = b.userServerCounter[server]; !ok {
        sm = make(map[int64]int32, b.session)
        b.userServerCounter[server] = sm
    }
    if incr {
        sm[userId]++
        b.roomCounter[roomId]++
        b.serverCounter[server]++
    } else {
        // WARN:
        // if decr a userid but key not exists just ignore
        // this may not happen
        if v, _ = sm[userId]; v-1 == 0 {
            delete(sm, userId)
        } else {
            sm[userId] = v - 1
        }
        b.roomCounter[roomId]--
        b.serverCounter[server]--
    }
}

// Put put a channel according with user id.
func (b *Bucket) Put(userId int64, server int32, roomId int32, ip string, data string) (seq int32, err error) {
    b.bLock.Lock()
    //	log.Info("put new client[uid: %d,ip: %s,data: %s ]", userId, ip, data)
    seq, err = b.sessionMgr.Put(userId, server, roomId, ip, data)
    if err != nil {
        b.bLock.Unlock()
        return
    }

   // b.counter(userId, server, roomId, true)  //统计信息没有什么用
    b.bLock.Unlock()
    //log.Info("-----session %v", s)
    return
}

func (b *Bucket) Update(userId int64, server int32, roomId int32, ip string, data string) (err error) {
    b.bLock.Lock()
    log.Info("update uid[%d] client[ip: %s,data: %s ]", userId, ip, data)
    b.bLock.Unlock()
    return
}

func (b *Bucket) Get(userId int64) (seqs []int32, servers []int32) {
    b.bLock.RLock()
    seqs, servers = b.sessionMgr.Get(userId)
    b.bLock.RUnlock()
    return
}

func (b *Bucket) GetWithAppid(userId int64, appid int32) (seqs []int32, servers []int32) {
    b.bLock.RLock()
    seqs, servers = b.sessionMgr.GetWithAppid(userId, appid)
    b.bLock.RUnlock()
    return
}

func (b *Bucket) GetInfo(userId int64) (info *proto.ClientInfo) {
    b.bLock.RLock()
    info, ok := b.sessionMgr.GetInfo(userId)
    if ok {
        log.Info("Get ClientInfo(uid:%d) [ip: %s, data: %s,status: %v", userId, info.Ip, info.ClientData, info.Status)
    }
    b.bLock.RUnlock()
    return
}

func (b *Bucket) GetAll() (userIds []int64, seqs [][]int32, servers [][]int32) {
    // b.bLock.RLock()
    // i := len(b.sessions)
    // userIds = make([]int64, i)
    // seqs = make([][]int32, i)
    // servers = make([][]int32, i)
    // for userId, s := range b.sessions {
    //     i--
    //     userIds[i] = userId
    //     seqs[i], servers[i] = s.Servers()
    // }
    // b.bLock.RUnlock()
    return
}

// Del delete the channel by sub key.
func (b *Bucket) Del(userId int64, seq int32, roomId int32) (ok bool) {
    b.bLock.Lock()
    ok = b.sessionMgr.DelSeq(userId, seq)
    b.bLock.Unlock()
    return
}

func (b *Bucket) DelByUid(userId int64) (ok bool) {
    b.bLock.Lock()
    ok = b.sessionMgr.DelUser(userId)
    b.bLock.Unlock()

    return
}

func (b *Bucket) DelServer(server int32) {
    return
}

func (b *Bucket) count(roomId int32) (count int32) {
    b.bLock.RLock()
    count = b.roomCounter[roomId]
    b.bLock.RUnlock()
    return
}

func (b *Bucket) Count() (count int32) {
    count = b.count(define.NoRoom)
    return
}

func (b *Bucket) RoomCount(roomId int32) (count int32) {
    count = b.count(roomId)
    return
}

func (b *Bucket) AllRoomCount() (roomCounter map[int32]int32) {
    var roomId, count int32
    b.bLock.RLock()
    roomCounter = make(map[int32]int32, len(b.roomCounter))
    for roomId, count = range b.roomCounter {
        if count > 0 {
            roomCounter[roomId] = count
        }
    }
    b.bLock.RUnlock()
    return
}

func (b *Bucket) AllServerCount() (serverCounter map[int32]int32) {
    var server, count int32
    b.bLock.RLock()
    serverCounter = make(map[int32]int32, len(b.serverCounter))
    for server, count = range b.serverCounter {
        serverCounter[server] = count
    }
    b.bLock.RUnlock()
    return
}

func (b *Bucket) UserCount(userId int64) (count int32) {
    b.bLock.RLock()
    seqs, _ := b.sessionMgr.Get(userId)
    count = int32(len(seqs))
    b.bLock.RUnlock()
    return
}
