package main

import (
    "fmt"
    "github.com/martini-contrib/render"
    "gopkg.in/mgo.v2/bson"
    "log"
    "net/http"
    "strconv"
)

type (
    // LatencyMetrics holds computed request latency metrics.
    LatencyMetrics struct {
        // Mean is the average request latency.
        Average float64 `json:"average"`
        // P50 is the 50th percentile request latency.
        P50 int64 `json:"50th"`
        // P90 is the 90th percentile request latency.
        P90 int64 `json:"90th"`
        // P95 is the 95th percentile request latency.
        P95 int64 `json:"95th"`
        // P99 is the 99th percentile request latency.
        P99 int64 `json:"99th"`
        // Min is the minimum observed request latency.
        Min int64 `json:"min"`
        // Max is the maximum observed request latency.
        Max int64 `json:"max"`
    }

    Elapsed struct {
        TotalElapsed int64
        Num          int32
    }

    TimeMetrics struct {
        Elapsed     []*Elapsed
        Tps         map[string]int32
    }

    JmeterMetrics struct {
        Label       string `json:"label"`
        Total       int64 `json:"total"`
        Latencies   LatencyMetrics `json:"latencies"`
        ErrorRate   float64 `json:"error_rate"`
        Qps         float64 `json:"qps"`
        Kbs         int64
        StartTime   int64
        StopTime    int64
        TimeMetrics map[string]*TimeMetrics
    }

    AttackJmeterLog struct {
        Id          bson.ObjectId `json:"id"        bson:"_id,omitempty"`
        JobName     string
        State       string
        MetricsList []*JmeterMetrics
        StartTs     int64
        EndTs       int64
    }
)

func (log *AttackJmeterLog) IsRunning() bool {
    return log.State == "Running"
}

func (log *AttackJmeterLog) TransactionMetrics() map[string]interface{} {
    var i int64
    var endTime = log.EndTs - log.StartTs
    var timeList []string
    var elapsedTime []int64
    for i = 0; i < endTime; i++ {
        elapsedTime = append(elapsedTime, i)
        timeList = append(timeList, strconv.FormatInt(log.StartTs+i, 10))
    }

    var series []string
    var dataList []map[string][]int32
    for _, metrics := range log.MetricsList {
        timeMetrics := metrics.TimeMetrics
        if timeMetrics == nil {
            continue
        }
        series = append(series, metrics.Label)
        var data = make(map[string][]int32)
        for _, ts := range timeList {
            FinalMap := make(map[string]int32)
            if timeMetrics[ts] != nil {
                TpsMap := timeMetrics[ts].Tps
                for k, v := range TpsMap {
                    if k == "200" {
                        FinalMap["success"] += v
                    } else {
                        FinalMap["failure"] += v
                    }
                }
                if FinalMap["success"] == 0 {
                    FinalMap["success"] = -1
                }
                if FinalMap["failure"] == 0 {
                    FinalMap["failure"] = -1
                }
            } else {
                FinalMap["success"] = -1
                FinalMap["failure"] = -1
            }
            data["success"] = append(data["success"], FinalMap["success"])
            data["failure"] = append(data["failure"], FinalMap["failure"])
        }
        dataList = append(dataList, data)
    }

    result := map[string]interface{} {
        "elapsed": elapsedTime,
        "series" : series,
        "data_list": dataList,
    }

    return result
}

func GetJmeterLogs(req *http.Request, r render.Render) {
    var jobName = req.FormValue("job_name")
    var page = req.FormValue("p")
    var logs []AttackJmeterLog
    var condition = bson.M{}
    if jobName != "" {
        condition = bson.M{"jobname": jobName}
    } else {
        condition = nil
    }
    total, err := G_MongoDB.C("jmeter_logs").Find(condition).Count()
    if err != nil {
        log.Panic(err)
    }
    var pager = NewPager(20, total)
    pager.CurrentPage, err = strconv.Atoi(page)
    pager.UrlPattern = fmt.Sprintf("/jmeter/logs?&p=%%d&job_name=%s", jobName)
    err = G_MongoDB.C("jmeter_logs").Find(condition).Skip(pager.Offset()).Sort("-startts").Limit(pager.Limit()).All(&logs)
    if err != nil {
        log.Panic(err)
    }
    var context = make(map[string]interface{})
    context["logs"] = logs
    context["jobName"] = jobName
    context["pager"] = pager
    RenderTemplate(r, "jmeter_logs", context)
}

func GetJmeterMetrics(req *http.Request, r render.Render) {
    var lg AttackJmeterLog
    var lgId = bson.ObjectIdHex(req.FormValue("log_id"))
    err := G_MongoDB.C("jmeter_logs").FindId(lgId).One(&lg)
    if err != nil {
        log.Panic(err)
    }
    var context = make(map[string]interface{})
    context["log"] = &lg
    RenderTemplate(r, "jmeter_metrics", context)
}

func DeleteJmeterLog(req *http.Request, r render.Render) {
    var logId = bson.ObjectIdHex(req.FormValue("log_id"))
    err := G_MongoDB.C("jmeter_logs").RemoveId(logId)
    if err != nil {
        log.Panic(err)
    }
    r.Redirect(req.Referer())
}
