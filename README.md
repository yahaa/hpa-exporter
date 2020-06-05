### hpa-exporter

### 功能

* 通过 client-go 定时抓取集群内部 hpa 信息转换成 metric
 

### 核心 metrics
* `hpa_last_scale_second` hpa 上一次扩容时间
* `hpa_current_metrics_value` hpa 当前值,可以通过这个指标获取一组服务的平均 cpu 使用率
* `hpa_target_metrics_value` hpa 阈值
* `hpa_able_to_scale` hpa 是否能扩容
* `hpa_scaling_limited` hpa 扩容是否受限

注: hpa 扩容受限条件 `(CurrentCPUUtilizationPercentage>=TargetCPUUtilizationPercentage && a.Status.DesiredReplicas >= a.Spec.MaxReplicas)`


### 部署

```bash
$ kubectl create ns monitoring # 如果有了 ns 可以跳过
$ kubectl apply -f deploy/deploy.yaml
```
