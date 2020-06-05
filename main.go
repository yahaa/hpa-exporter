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
	hpaLastScaleSecond     *prometheus.GaugeVec
	hpaCurrentMetricsValue *prometheus.GaugeVec
	hpaTargetMetricsValue  *prometheus.GaugeVec
	hpaAbleToScale         *prometheus.GaugeVec
	hpaScalingLimited      *prometheus.GaugeVec

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

	hpaLastScaleSecond = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpa_last_scale_second",
			Help: "Time the scale was last executed.",
		},
		baseLabels,
	)

	hpaCurrentMetricsValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpa_current_metrics_value",
			Help: "Current Metrics Value.",
		},
		baseLabels,
	)

	hpaTargetMetricsValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpa_target_metrics_value",
			Help: "Target Metrics Value.",
		},
		baseLabels,
	)

	hpaAbleToScale = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpa_able_to_scale",
			Help: "status able to scale from annotation.",
		},
		baseLabels,
	)

	hpaScalingLimited = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hpa_scaling_limited",
			Help: "status scaling limited from annotation.",
		},
		baseLabels,
	)

	return []prometheus.Collector{
		hpaLastScaleSecond,
		hpaCurrentMetricsValue,
		hpaTargetMetricsValue,
		hpaAbleToScale,
		hpaScalingLimited,
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
			hpaLastScaleSecond.With(baseLabel).Set(float64(a.Status.LastScaleTime.Unix()))
		}

		if a.Spec.TargetCPUUtilizationPercentage != nil {
			hpaTargetMetricsValue.With(baseLabel).Set(float64(*a.Spec.TargetCPUUtilizationPercentage))
		}

		if a.Status.CurrentCPUUtilizationPercentage != nil {
			hpaCurrentMetricsValue.With(baseLabel).Set(float64(*a.Status.CurrentCPUUtilizationPercentage))
		}

		if a.Status.CurrentCPUUtilizationPercentage != nil && a.Spec.TargetCPUUtilizationPercentage != nil {
			if *a.Status.CurrentCPUUtilizationPercentage >= *a.Spec.TargetCPUUtilizationPercentage && a.Status.DesiredReplicas >= a.Spec.MaxReplicas {
				hpaScalingLimited.With(baseLabel).Set(float64(1))
				hpaAbleToScale.With(baseLabel).Set(float64(0))
			} else {
				hpaScalingLimited.With(baseLabel).Set(float64(0))
				hpaAbleToScale.With(baseLabel).Set(float64(1))
			}
		}
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
