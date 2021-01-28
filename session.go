package main

import (
    // "bytes"
    // "errors"
    log "github.com/thinkboy/log4go"
    // "goim/libs/define"
    "goim/libs/proto"
    // "fmt"
    "errors"
    "time"
    "unsafe"
)

/*
#include <stdlib.h>
#include "spool.h"
*/
import "C"

type SessionMgr struct {
    cap          int              //容量
    idleIndexSet map[int]struct{} //idle集合
    uidPosMap    map[int64]int    //uid:pos
    pool *C.CSessionPool // session列表
}

func NewSessionMgr(cap int) *SessionMgr {
    mgr := new(SessionMgr)
    log.Info("cap=%v", cap)
    mgr.cap = cap
    mgr.idleIndexSet = make(map[int]struct{}, cap)
    for i := 0; i < cap; i++ {
        mgr.idleIndexSet[i] = struct{}{}
    }
    mgr.uidPosMap = make(map[int64]int, cap)
    mgr.pool = C.NewCSessionPool(C.int32_t(cap))
    if mgr.pool == nil {
        panic("session pool create failed")
    }
    return mgr
}

func (m *SessionMgr) Put(userId int64, server int32, roomId int32, ip string, data string) (seq int32, err error) {
    index, ok := m.uidPosMap[userId]
    if !ok {
        if len(m.idleIndexSet) == 0 {
            return 0, errors.New("session pool is full")
        }
        for idleIndex, _ := range m.idleIndexSet {
            m.uidPosMap[userId] = idleIndex
            delete(m.idleIndexSet, idleIndex)
            index = idleIndex
            break
        }
    }
    cip := C.CString(ip)
    cdata := C.CString(data)
    t := time.Now().UnixNano()
    seqt := C.putSession(m.pool, C.int32_t(index), C.int32_t(server), C.int32_t(roomId), cip, cdata, C.int64_t(t))
    seq = int32(seqt)
    C.free(unsafe.Pointer(cip))
    C.free(unsafe.Pointer(cdata))
    if seq < 0 {
        log.Error("put %v to pool err, seq=%v", userId, seq)
        err = errors.New("put to pool err")
        return 0, err
    }
    log.Info("put %v to pool, index=%v, seq=%v, roomId=%v, server=%v, ip=%v, data=%v",
        userId, index, seq, roomId, server, ip, data)
    return seq, nil
}

func (m *SessionMgr) Get(userId int64) (seqs []int32, servers []int32) {
    index, ok := m.uidPosMap[userId]
    if !ok {
        return
    }

    var s *C.CSession
    s = C.getSession(m.pool, C.int32_t(index))
    if s == nil {
        log.Error("not found session in pool, uid=%v", userId)
        return
    }
    var i int32 = 0
    for i = 0; i < C.MAX_ITEM_IN_SESSION; i++ {
        item := &(s.items[i])
        if item.status > 0 {
            seqs = append(seqs, i+1)
            servers = append(servers, int32(item.server))
        }
        log.Info("index=%v, i=%d, status=%v, server=%v, roomid=%v, ip=%v, data=%v, time=%v", index,
            i, item.status, item.server, item.roomid, C.GoString(&(item.ip[0])), C.GoString(item.data), item.time)
    }

    return
}

func (m *SessionMgr) GetWithAppid(userId int64, appid int32) (seqs []int32, servers []int32) {
    index, ok := m.uidPosMap[userId]
    if !ok {
        return
    }

    var s *C.CSession
    s = C.getSession(m.pool, C.int32_t(index))
    if s == nil {
        log.Error("not found session in pool, uid=%v", userId)
        return
    }
    var i int32 = 0
    for i = 0; i < C.MAX_ITEM_IN_SESSION; i++ {
        item := &(s.items[i])
        if item.status > 0 && int32(item.roomid) == appid {
            seqs = append(seqs, i+1)
            servers = append(servers, int32(item.server))
        }
        log.Info("index=%v, i=%d, status=%v, server=%v, roomid=%v, ip=%v, data=%v, time=%v", index,
            i, item.status, item.server, item.roomid, C.GoString(&(item.ip[0])), C.GoString(item.data), item.time)
    }

    return
}

func (m *SessionMgr) GetInfo(userId int64) (info *proto.ClientInfo, ok bool) {
    index, ok := m.uidPosMap[userId]
    if !ok {
        return nil, false
    }

    var s *C.CSession
    s = C.getSession(m.pool, C.int32_t(index))
    if s == nil {
        log.Error("not found session in pool, uid=%v", userId)
        return nil, false
    }
    info = &proto.ClientInfo{}
    for i := 0; i < C.MAX_ITEM_IN_SESSION; i++ {
        item := &(s.items[i])
        if item.status > 0 {
            info.AppId = int32(item.roomid)
            info.Ip = C.GoString(&(item.ip[0]))
            info.ClientData = C.GoString(item.data)
            info.Status = true
            return info, ok
        }
        // log.Info("i=%d, status=%v, server=%v, roomid=%v, ip=%v, data=%v, time=%v\n",
        //     i, item.status, item.server, item.roomid, C.GoString(&(item.ip[0])),  C.GoString(item.data), item.time)
    }

    return info, false
}

func (m *SessionMgr) DelSeq(userId int64, seq int32) (ok bool) {
    index, ok := m.uidPosMap[userId]
    if ok {
        r := C.deleteSessionSeq(m.pool, C.int32_t(index), C.int32_t(seq))
        if r == 0 {
            m.idleIndexSet[index] = struct{}{}
            delete(m.uidPosMap, userId)

            log.Info("2222222222 userId=%v index = %v, len(idleIndexSet)=%v, len(uidPosMap)=%v",
                userId, index, len(m.idleIndexSet), len(m.uidPosMap))
        }
        return
    }
    return
}

func (m *SessionMgr) DelUser(userId int64) (ok bool) {
    index, ok := m.uidPosMap[userId]
    if ok {
        m.idleIndexSet[index] = struct{}{}
        delete(m.uidPosMap, userId)
        log.Info("2222222222 userId=%v index = %v, len(idleIndexSet)=%v, len(uidPosMap)=%v",
            userId, index, len(m.idleIndexSet), len(m.uidPosMap))
        C.deleteSessionAllSeq(m.pool, C.int32_t(index))
        return
    }
    return
}
