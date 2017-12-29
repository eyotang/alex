package main

import (
    "fmt"
    "github.com/martini-contrib/render"
    "gopkg.in/mgo.v2/bson"
    "log"
    "net/http"
    "strconv"
)

type Elapsed struct {
    TotalElapsed int64
    Num          int32
}

type TimeMetrics struct {
    Elapsed     []*Elapsed
    Tps         map[string]int32
}

type JmeterMetrics struct {
    Success     int64
    Elapsed     []int32
    StartTime   int64
    StopTime    int64
    Bytes       int32
    TimeMetrics map[string]TimeMetrics
}

type AttackJmeterLog struct {
    Id          bson.ObjectId `json:"id"        bson:"_id,omitempty"`
    JobName     string
    State       string
    Metrics     map[string]JmeterMetrics
    StartTs     int64
    EndTs       int64
}

func (log *AttackJmeterLog) IsRunning() bool {
    return log.State == "Running"
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

func DeleteJmeterLog(req *http.Request, r render.Render) {
    var logId = bson.ObjectIdHex(req.FormValue("log_id"))
    err := G_MongoDB.C("jmeter_logs").RemoveId(logId)
    if err != nil {
        log.Panic(err)
    }
    r.Redirect(req.Referer())
}
