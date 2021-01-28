#ifndef SPOOL_H
#define SPOOL_H

#define MAX_ITEM_IN_SESSION 16
#define MAX_IP_STRING_LEN 32
#include "stdint.h"

typedef struct _CSessionItem
{
    int32_t server; // session对于的服务器标识
    int32_t roomid; // session 对于的roomid
    int64_t time;
    uint8_t status; // 0 未使用  其它已经使用
    char *data;
    char ip[MAX_IP_STRING_LEN]; //客户端ip
} CSessionItem;

typedef struct _CSession
{
    CSessionItem items[MAX_ITEM_IN_SESSION];
} CSession;

typedef struct _CSessionPool
{
    CSession *sessionArray;
    int32_t cap;
} CSessionPool;

CSessionPool *NewCSessionPool(int cap);

// <0  失败  >0 返回seq
int32_t putSession(CSessionPool *pool, int32_t index, int32_t server, int32_t roomId, char *ip, char *data, int64_t time);

// < 0 错误,>=0 剩余的item数量
int32_t deleteSessionSeq(CSessionPool *pool, int32_t index, int32_t seq);

int32_t deleteSessionAllSeq(CSessionPool *pool, int32_t index);

CSession *getSession(CSessionPool *pool, int32_t index);


#endif //SPOOL_H
