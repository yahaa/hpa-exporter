package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yahaa/hap-exporter/utils"
	asv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

var (
	hpaStatusLastScaleSecond     *prometheus.GaugeVec
	hpaStatusCurrentMetricsValue *prometheus.GaugeVec
	hpaStatusCurrentReplicas     *prometheus.GaugeVec
	hpaStatusDesiredReplicas     *prometheus.GaugeVec
	hpaStatusAbleToScale         *prometheus.GaugeVec
	hpaStatusScalingLimited      *prometheus.GaugeVec

	hpaSpecMinReplicas        *prometheus.GaugeVec
	hpaSpecMaxReplicas        *prometheus.GaugeVec
	hpaSpecTargetMetricsValue *prometheus.GaugeVec

	kubeClient *kubernetes.Clientset

	config = struct {
		Kubeconfig      string
		SrvAddr         string
		AdditionalLabel string
		CollectInterval int
	}{}

	baseLabels = []string{
		"hpa_name",
		"hpa_namespace",
		"ref_kind",
		"ref_name",
		"ref_apiversion",
	}
)

func init() {
	flagSet := flag.CommandLine
	klog.InitFlags(flagSet)

	flagSet.StringVar(&config.Kubeconfig, "kubeconfig", "", "kubeconfig path")
	flagSet.StringVar(&config.SrvAddr, "listen-addr", ":9099", "server listen address")
	flagSet.StringVar(&config.AdditionalLabel, "additional-label", "", "additional label to append on metrics")
	flagSet.IntVar(&config.CollectInterval, "collect-interval", 15, "collect hpa interval")

	flagSet.Parse(os.Args[1:])
}

func initCollectors() []prometheus.Collector {
	if config.AdditionalLabel != "" {
		baseLabels = append(baseLabels, config.AdditionalLabel)
	}

	hpaSpecMinReplicas = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpa_spec_min_replicas",
			Help: "hpa spec min replicas.",
		},
		baseLabels,
	)

	hpaSpecMaxReplicas = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpa_spec_max_replicas",
			Help: "hpa spec max replicas.",
		},
		baseLabels,
	)

	hpaSpecTargetMetricsValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpa_spce_target_metrics_value",
			Help: "Target Metrics Value.",
		},
		baseLabels,
	)

	hpaStatusCurrentReplicas = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpa_status_current_replicas",
			Help: "hpa current replicas.",
		},
		baseLabels,
	)

	hpaStatusLastScaleSecond = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpa_status_last_scale_second",
			Help: "Time the scale was last executed.",
		},
		baseLabels,
	)

	hpaStatusCurrentMetricsValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpa_status_current_metrics_value",
			Help: "Current Metrics Value.",
		},
		baseLabels,
	)

	hpaStatusAbleToScale = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpa_status_able_to_scale",
			Help: "status able to scale from annotation.",
		},
		baseLabels,
	)

	hpaStatusScalingLimited = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpa_status_scaling_limited",
			Help: "status scaling limited from annotation.",
		},
		baseLabels,
	)
	hpaStatusDesiredReplicas = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpa_status_desired_replicas",
			Help: "hpa status desired replicas.",
		},
		baseLabels,
	)

	return []prometheus.Collector{
		hpaSpecMaxReplicas,
		hpaSpecMinReplicas,
		hpaSpecTargetMetricsValue,
		hpaStatusLastScaleSecond,
		hpaStatusCurrentReplicas,
		hpaStatusCurrentMetricsValue,
		hpaStatusAbleToScale,
		hpaStatusDesiredReplicas,
		hpaStatusScalingLimited,
	}

}

func getHpaListV1() ([]asv1.HorizontalPodAutoscaler, error) {
	var err error
	if kubeClient == nil {
		kubeClient, err = utils.NewClientset(config.Kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	out, err := kubeClient.AutoscalingV1().HorizontalPodAutoscalers(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return out.Items, err
}

func collectorV1(hpa []asv1.HorizontalPodAutoscaler, additionalLabel string) {
	for _, a := range hpa {
		baseLabel := prometheus.Labels{
			"hpa_name":       a.ObjectMeta.Name,
			"hpa_namespace":  a.ObjectMeta.Namespace,
			"ref_kind":       a.Spec.ScaleTargetRef.Kind,
			"ref_name":       a.Spec.ScaleTargetRef.Name,
			"ref_apiversion": a.Spec.ScaleTargetRef.APIVersion,
		}

		if additionalLabel != "" {
			baseLabel[additionalLabel] = a.Labels[additionalLabel]
		}

		if a.Status.LastScaleTime != nil {
			hpaStatusLastScaleSecond.With(baseLabel).Set(float64(a.Status.LastScaleTime.Unix()))
		}

		if a.Spec.TargetCPUUtilizationPercentage != nil {
			hpaSpecTargetMetricsValue.With(baseLabel).Set(float64(*a.Spec.TargetCPUUtilizationPercentage))
		}

		hpaSpecMaxReplicas.With(baseLabel).Set(float64(a.Spec.MaxReplicas))

		if a.Spec.MinReplicas != nil {
			hpaSpecMinReplicas.With(baseLabel).Set(float64(*a.Spec.MinReplicas))
		}

		if a.Status.CurrentCPUUtilizationPercentage != nil {
			hpaStatusCurrentMetricsValue.With(baseLabel).Set(float64(*a.Status.CurrentCPUUtilizationPercentage))
		}

		if a.Status.CurrentCPUUtilizationPercentage != nil && a.Spec.TargetCPUUtilizationPercentage != nil {
			if *a.Status.CurrentCPUUtilizationPercentage >= *a.Spec.TargetCPUUtilizationPercentage && a.Status.DesiredReplicas >= a.Spec.MaxReplicas {
				hpaStatusScalingLimited.With(baseLabel).Set(float64(1))
				hpaStatusAbleToScale.With(baseLabel).Set(float64(0))
			} else {
				hpaStatusScalingLimited.With(baseLabel).Set(float64(0))
				hpaStatusAbleToScale.With(baseLabel).Set(float64(1))
			}
		}

		hpaStatusCurrentReplicas.With(baseLabel).Set(float64(a.Status.CurrentReplicas))
		hpaStatusDesiredReplicas.With(baseLabel).Set(float64(a.Status.DesiredReplicas))
	}
}

func main() {
	collectors := initCollectors()
	prometheus.MustRegister(collectors...)

	klog.Info("start hpa exporter...")

	go func() {
		for {

			hpaV1, err := getHpaListV1()
			if err != nil {
				klog.Error("list hpa v1 err:", err)
				continue
			}

			collectorV1(hpaV1, config.AdditionalLabel)

			time.Sleep(time.Duration(config.CollectInterval) * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	klog.Fatal(http.ListenAndServe(config.SrvAddr, nil))
}
