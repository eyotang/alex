package main

const PERF_INTERVAL = 1
type (
    PerfMetrics struct {
        CpuUsage CpuUsageMetrics `json:"cpuusage"`
        MemUsage MemUsageMetrics `json:"memusage"`
        DiskUsage DiskUsageMetrics `json:"diskusage"`
        Network   NetworkMetrics `json:"network"`
    }

    CpuUsageMetrics struct {
        User   float64 `json:"user"`
        System float64 `json:"system"`
    }

    MemUsageMetrics struct {
        Mem     float64 `json:"mem"`
        MemSwap float64 `json:"memswap"`
    }

    DiskUsageMetrics struct {
        Rrqm float64 `json:"rrqm"`
        Wrqm float64 `json:"wrqm"`
    }

    NetworkMetrics struct {
        Rxpck float64 `json:"rxpck"`
        Txpck float64 `json:"txpck"`
        Rxkb  float64 `json:"rxkb"`
        Txkb  float64 `json:"txkb"`
    }
)
