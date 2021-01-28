#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include "spool.h"

static int32_t resetCSessionItem(CSessionItem *item)
{
    if (item == NULL)
    {
        return -1;
    }
    item->server = 0;
    item->roomid = 0;
    item->status = 0;
    item->time = 0;
    memset(item->ip, 0, MAX_IP_STRING_LEN);
    if (item->data != NULL)
    {
        free(item->data);
        item->data = NULL;
    }
    return 0;
}

static int32_t resetCSession(CSession *session)
{
    if (session == NULL)
    {
        return -1;
    }
    int i = 0;
    for (i = 0; i < MAX_ITEM_IN_SESSION; i++)
    {
        resetCSessionItem(&(session->items[i]));
    }

    return 0;
}

static int32_t getCSessionItemCount(CSession *session)
{
    if (session == NULL)
    {
        return -1;
    }
    int r = 0;
    int i = 0;
    for (i = 0; i < MAX_ITEM_IN_SESSION; i++)
    {
        if (session->items[i].status != 0)
        {
            r++;
        }
    }
    return r;
}

CSessionPool *NewCSessionPool(int cap)
{
    CSessionPool *pool = (CSessionPool *)malloc(sizeof(CSessionPool));
    if (pool == NULL)
    {
        return NULL;
    }
    size_t msize = sizeof(CSession) * cap;

    CSession *s = (CSession *)malloc(msize);
    if (s == NULL)
    {
        free(pool);
        return NULL;
    }
    memset(s, 0, msize);
    pool->sessionArray = s;
    pool->cap = cap;

    return pool;
}

// <0  失败  >0 返回seq
int32_t putSession(CSessionPool *pool, int32_t index, int32_t server, int32_t roomId, char *ip, char *data, int64_t time)
{
    if (pool == NULL)
    {
        return -1;
    }
    if (index < 0 || index > pool->cap)
    {
        return -2;
    }

    CSession *session = &(pool->sessionArray[index]);

    int i = 0;
    for (i = 0; i < MAX_ITEM_IN_SESSION; i++)
    {
        if (session->items[i].status == 0)
        {
            CSessionItem *item = &(session->items[i]);
            item->status = 1;
            item->server = server;
            item->roomid = roomId;
            item->time = time;
            strncpy(item->ip, ip, MAX_IP_STRING_LEN - 1);
            item->data = (char *)malloc(strlen(data) + 1);
            strcpy(item->data, data);
            return i + 1;
        }
    }
    return -3;
}

// < 0 错误,>=0 剩余的item数量
int32_t deleteSessionSeq (CSessionPool *pool, int32_t index, int32_t seq)
{
    if (pool == NULL)
    {
        return -1;
    }
    if (index < 0 || index > pool->cap)
    {
        return -2;
    }
    if (seq < 1 || seq > MAX_ITEM_IN_SESSION)
    {
        return -3;
    }

    CSession *session = &(pool->sessionArray[index]);
    int64_t time = session->items[seq - 1].time;
    resetCSessionItem(&(session->items[seq - 1]));
    int i = 0;
    for (i = 0; i < MAX_ITEM_IN_SESSION; i++)
    {
        if (session->items[i].status != 0 && session->items[i].time < time)
        {
            resetCSessionItem(&(session->items[i]));
        }
    }

    return getCSessionItemCount(session);
}

int32_t deleteSessionAllSeq(CSessionPool *pool, int32_t index)
{
    if (pool == NULL)
    {
        return -1;
    }
    if (index < 0 || index > pool->cap)
    {
        return -2;
    }

    resetCSession(&(pool->sessionArray[index]));
    return 0;
}

// items 未返回值，用户传入，返回值为数量
CSession *getSession(CSessionPool *pool, int32_t index)
{
    if (pool == NULL)
    {
        return NULL;
    }
    if (index < 0 || index > pool->cap)
    {
        return NULL;
    }
    return &(pool->sessionArray[index]);
}

