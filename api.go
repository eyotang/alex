package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/martini-contrib/render"
	vegeta "github.com/eyotang/vegeta/lib"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
)

func GetSystemStatus(req *http.Request, r render.Render) {
	// get system status of benchmark machine
	var result = map[string]string{}
	var l, _ = load.Avg()
	result["load:1"] = fmt.Sprintf("%v", l.Load1)
	result["load:5"] = fmt.Sprintf("%v", l.Load5)
	result["load:15"] = fmt.Sprintf("%v", l.Load15)
	var m, _ = mem.VirtualMemory()
	result["mem:total"] = fmt.Sprintf("total:%vM", m.Total>>20)
	result["mem:free"] = fmt.Sprintf("free:%vM", m.Free>>20)
	result["mem:buffers"] = fmt.Sprintf("buffers:%vM", m.Buffers>>20)
	result["mem:cached"] = fmt.Sprintf("cached:%vM", m.Cached>>20)
	result["mem:wired"] = fmt.Sprintf("wired:%vM", m.Wired>>20)
	result["mem:used"] = fmt.Sprintf("used:%.2f%%", m.UsedPercent)
	r.JSON(200, result)
}

func GetVegetaJobState(req *http.Request, r render.Render) {
	var jobId = req.FormValue("job_id")
	var job VegetaJob
	err := G_MongoDB.C("vegeta_jobs").FindId(bson.ObjectIdHex(jobId)).One(&job)
	var result = map[string]interface{}{}
	if err != nil {
		result["is_running"] = false
		result["current_rate"] = 0
	} else {
		result["is_running"] = job.IsRunning()
		result["current_rate"] = job.CurrentRate
	}
	r.JSON(200, result)
}

func GetBoomJobState(req *http.Request, r render.Render) {
	var jobId = req.FormValue("job_id")
	var job BoomJob
	err := G_MongoDB.C("boom_jobs").FindId(bson.ObjectIdHex(jobId)).One(&job)
	var result = map[string]interface{}{}
	if err != nil {
		result["is_running"] = false
		result["current_concurrency"] = 0
	} else {
		result["is_running"] = job.IsRunning()
		result["current_concurrency"] = job.CurrentConcurrency
	}
	r.JSON(200, result)
}

func TestParam(req *http.Request, r render.Render) {
	var host = req.FormValue("host")
	var url = req.FormValue("url")
	var header = req.FormValue("header")
	var params = req.FormValue("param")
	var data = req.FormValue("body")
	var expectation = req.FormValue("expectation")
	var method = req.FormValue("method")
	var headerMap map[string]interface{}
	var paramMap map[string]interface{}
	var dataMap map[string]interface{}
	var expectationMap map[string]interface{}
	json.Unmarshal([]byte(header), &headerMap)
	json.Unmarshal([]byte(params), &paramMap)
	json.Unmarshal([]byte(data), &dataMap)
	json.Unmarshal([]byte(expectation), &expectationMap)
	Host, relativeUrl := UrlSplit(url)
	rq, _ := http.NewRequest(method, Urlcat(host, relativeUrl, paramMap), bytes.NewReader(BodyBytes(dataMap)))
	for k, vs := range headerMap {
		rq.Header.Add(k, vs.(string))
	}
	rq.Host = Host
	if host := rq.Header.Get("Host"); host != "" {
		rq.Host = host
	}
	if method == "POST" {
		if contentType := rq.Header.Get("Content-Type"); contentType == "" {
			contentType = "application/x-www-form-urlencoded"
			rq.Header.Add("Content-Type", contentType)
		}
	}
	client := &http.Client{}
	var result = map[string]interface{}{}
	resp, err := client.Do(rq)
	if err == nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			result["err"] = err.Error()
		} else {
			var bodyMap map[string]interface{}
			err := json.Unmarshal(body, &bodyMap)
			if err != nil {
				result["err"] = err.Error()
			} else {
				result["result"] = bodyMap
				if vegeta.ExpectNear(expectationMap, bodyMap) {
					result["expectation"] = "OK"
				} else {
					result["expectation"] = "NOK"
				}
			}
		}
	} else {
		result["err"] = err
	}
	r.JSON(200, result)
}
