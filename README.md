### hpa-exporter

### 功能

* 通过 client-go 定时抓取集群内部 hpa 信息转换成 metric
 

### 核心 metrics
| Metric name                       | Metric type | Labels/tags                                                   | Status |
| --------------------------------  | ----------- | ------------------------------------------------------------- | ------ |
| hpa_status_last_scale_second             | Gauge       | `hpa_name`=&lt;hpa-name&gt; <br> `hpa_namespace`=&lt;hpa-namespace&gt; <br> `ref_kind`=&lt;ref_kind&gt; <br> `ref_name`=&lt;ref_name&gt; <br> `ref_apiversion`=&lt;ref_apiversion&gt;| STABLE |
| hpa_status_current_metrics_value         | Gauge       | `hpa_name`=&lt;hpa-name&gt; <br> `hpa_namespace`=&lt;hpa-namespace&gt; <br> `ref_kind`=&lt;ref_kind&gt; <br> `ref_name`=&lt;ref_name&gt; <br> `ref_apiversion`=&lt;ref_apiversion&gt;| STABLE |
| hpa_status_current_replicas         | Gauge       | `hpa_name`=&lt;hpa-name&gt; <br> `hpa_namespace`=&lt;hpa-namespace&gt; <br> `ref_kind`=&lt;ref_kind&gt; <br> `ref_name`=&lt;ref_name&gt; <br> `ref_apiversion`=&lt;ref_apiversion&gt;| STABLE |
| hpa_status_desired_replicas         | Gauge       | `hpa_name`=&lt;hpa-name&gt; <br> `hpa_namespace`=&lt;hpa-namespace&gt; <br> `ref_kind`=&lt;ref_kind&gt; <br> `ref_name`=&lt;ref_name&gt; <br> `ref_apiversion`=&lt;ref_apiversion&gt;| STABLE |
| hpa_status_able_to_scale                 | Gauge       | `hpa_name`=&lt;hpa-name&gt; <br> `hpa_namespace`=&lt;hpa-namespace&gt; <br> `ref_kind`=&lt;ref_kind&gt; <br> `ref_name`=&lt;ref_name&gt; <br> `ref_apiversion`=&lt;ref_apiversion&gt;| STABLE |
| hpa_status_scaling_limited               | Gauge       | `hpa_name`=&lt;hpa-name&gt; <br> `hpa_namespace`=&lt;hpa-namespace&gt; <br> `ref_kind`=&lt;ref_kind&gt; <br> `ref_name`=&lt;ref_name&gt; <br> `ref_apiversion`=&lt;ref_apiversion&gt;| STABLE |
| hpa_spec_max_replicas               | Gauge       | `hpa_name`=&lt;hpa-name&gt; <br> `hpa_namespace`=&lt;hpa-namespace&gt; <br> `ref_kind`=&lt;ref_kind&gt; <br> `ref_name`=&lt;ref_name&gt; <br> `ref_apiversion`=&lt;ref_apiversion&gt;| STABLE |
| hpa_spec_min_replicas               | Gauge       | `hpa_name`=&lt;hpa-name&gt; <br> `hpa_namespace`=&lt;hpa-namespace&gt; <br> `ref_kind`=&lt;ref_kind&gt; <br> `ref_name`=&lt;ref_name&gt; <br> `ref_apiversion`=&lt;ref_apiversion&gt;| STABLE |
| hpa_spec_target_metrics_value          | Gauge       | `hpa_name`=&lt;hpa-name&gt; <br> `hpa_namespace`=&lt;hpa-namespace&gt; <br> `ref_kind`=&lt;ref_kind&gt; <br> `ref_name`=&lt;ref_name&gt; <br> `ref_apiversion`=&lt;ref_apiversion&gt;| STABLE |




* `hpa_status_last_scale_second` hpa 上一次扩容时间
* `hpa_status_current_metrics_value` hpa 当前值,可以通过这个指标获取一组服务的平均 cpu 使用率
* `hpa_spec_target_metrics_value` hpa 阈值
* `hpa_status_able_to_scale` hpa 是否能扩容
* `hpa_status_scaling_limited` hpa 扩容是否受限

注: hpa 扩容受限条件 `(CurrentCPUUtilizationPercentage>=TargetCPUUtilizationPercentage && a.Status.DesiredReplicas >= a.Spec.MaxReplicas)`


### 部署
* 程序启动参数还可以通过 `--additional-label` flag 参数指定额外的 label

```bash
$ kubectl create ns monitoring # 如果有了 ns 可以跳过
$ kubectl apply -f deploy/deploy.yaml
```
