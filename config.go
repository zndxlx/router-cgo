package main

import (
    "flag"
    "runtime"
    "time"

    "fmt"
    "github.com/Terry-Mao/goconf"
    "github.com/shima-park/agollo"
    "os"
)

var (
    gconf    *goconf.Config
    Conf     *Config
    confFile string
)

const (
    CommonCfgFile = "./common.conf"
)

func init() {
    flag.StringVar(&confFile, "c", "./router-example.conf", " set router config file path")
}

type Config struct {
    //apollo 配置
    ApolloAddr      string `goconf:"apollo:addr"`
    ApolloAppid     string `goconf:"apollo:appid"`
    ApolloCluster   string `goconf:"apollo:cluster"`
    ApolloNameSpace string `goconf:"apollo:namespace"`
    ApolloRefresh   int    `goconf:"apollo:refresh"`

    // base section
    PidFile    string   `goconf:"base:pidfile"`
    Dir        string   `goconf:"base:dir"`
    Log        string   `goconf:"base:log"`
    MaxProc    int      `goconf:"base:maxproc"`
    PprofAddrs []string `goconf:"base:pprof.addrs:,"`
    // rpc
    RPCAddrs []string `goconf:"rpc:addrs:,"`
    // bucket
    Bucket int `goconf:"bucket:bucket"`
    Server int `goconf:"bucket:server"`
    Cap    int `goconf:"bucket:cap"`
    //Cleaner           int           `goconf:"bucket:cleaner"`
    //BucketCleanPeriod time.Duration `goconf:"bucket:clean.period:time"`
    // session
    Session       int           `goconf:"session:session"`
    SessionExpire time.Duration `goconf:"session:expire:time"`
    // monitor
    MonitorOpen  bool     `goconf:"monitor:open"`
    MonitorAddrs []string `goconf:"monitor:addrs:,"`
}

func NewConfig() *Config {
    return &Config{
        // base section
        PidFile:    "/tmp/goim-router.pid",
        Dir:        "./",
        Log:        "./router-log.xml",
        MaxProc:    runtime.NumCPU(),
        PprofAddrs: []string{"localhost:6971"},
        // rpc
        RPCAddrs: []string{"localhost:9090"},
        // bucket
        Bucket: runtime.NumCPU(),
        Server: 5,
        // Cleaner:           1000,
        // BucketCleanPeriod: time.Hour * 1,
        // session
        Session:       1000,
        Cap:           1024,
        SessionExpire: time.Hour * 1,
    }
}

// InitConfig init the global config.
func InitConfig() (err error) {
    Conf = NewConfig()
    gconf = goconf.New()
    if err = gconf.Parse(confFile); err != nil {
        return err
    }
    if err := gconf.Unmarshal(Conf); err != nil {
        return err
    }

    UpdateCfg()
    commonConf := goconf.New()
    if err = commonConf.Parse(CommonCfgFile); err != nil {
        return err
    }
    if err := commonConf.Unmarshal(Conf); err != nil {
        return err
    }

    return nil
}

func ReloadConfig() (*Config, error) {
    conf := NewConfig()
    ngconf, err := gconf.Reload()
    if err != nil {
        return nil, err
    }
    if err := ngconf.Unmarshal(conf); err != nil {
        return nil, err
    }
    gconf = ngconf
    return conf, nil
}

func UpdateCfg() {
    a, err := agollo.New(Conf.ApolloAddr, Conf.ApolloAppid,
        agollo.Cluster(Conf.ApolloCluster),
        agollo.DefaultNamespace(Conf.ApolloNameSpace),
        agollo.PreloadNamespaces(Conf.ApolloNameSpace),
        agollo.AutoFetchOnCacheMiss(),
        agollo.FailTolerantOnBackupExists(),
        agollo.ConfigServerRefreshIntervalInSecond(time.Second*(time.Duration(Conf.ApolloRefresh))))
    if err != nil {
        fmt.Printf("StartApollo failed err = %+v", err)
    }

    // fmt.Printf("content=%+v\n", a.Get("content"))
    content := a.Get("content")
    if len(content) > 10 {
        f, err := os.OpenFile(CommonCfgFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
        if err != nil {
            fmt.Printf("open cfg file failed, err=%v\n", err)
        }
        defer f.Close()
        num, err := f.WriteString(content)
        if err != nil {
            fmt.Printf("write cfg file failed, num=%d, err=%v\n", num, err)
        }
    }

    return
}
