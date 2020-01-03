package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"
	"time"

	awsgo "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gruntwork-io/gruntwork-cli/files"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/collections"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NumBenchmarks   = 1
	ChartOutputPath = "./charts"

	EksClusterModulePath  = "./eks-cluster"
	BenchmarkModulePath   = "./benchmark"
	BenchmarkJobName      = "cpu-benchmark"
	BenchmarkJobNamespace = "default"

	BenchmarkTableName        = "EKSFargateCPUBenchmark"
	BenchmarkTableProcInfoKey = "ProcInfo"
	BenchmarkTableRunTimeKey  = "RunTime"

	// Wait for up to 1 hour (60 tries, 1 minute between tries)
	MaxRetries        = 60
	SleepBetweenTries = 1 * time.Minute
)

var FargateRegions = []string{
	"us-east-1",
	"us-east-2",
	"eu-west-1",
	"ap-northeast-1",
}

// For each Fargate region:
// - Launch a Fargate cluster using eks-cluster module
// - Run a benchmark job using benchmark module
// - Collect results into a graph and store it in the `charts/REGION` folder.
func TestFargateCpuInfoAllRegions(t *testing.T) {
	for _, awsRegion := range FargateRegions {
		t.Run(awsRegion, func(t *testing.T) {
			t.Parallel()
			runFargateCpuInfoTest(t, awsRegion)
		})
	}
}

func runFargateCpuInfoTest(t *testing.T, awsRegion string) {
	// Uncomment any of the following to skip that stage in the test
	//os.Setenv("SKIP_foo", "true") // This is not a real stage, but helps with skipping the terraform folder copying
	//os.Setenv("SKIP_setup", "true")
	//os.Setenv("SKIP_deploy_cluster", "true")
	//os.Setenv("SKIP_run_benchmark", "true")
	//os.Setenv("SKIP_collect_benchmark_results", "true")
	//os.Setenv("SKIP_destroy_benchmark", "true")
	//os.Setenv("SKIP_destroy_cluster", "true")

	// Namespace the working dir and modules by AWS region
	workingDir := filepath.Join("stages", awsRegion)
	testFolder := test_structure.CopyTerraformFolderToTemp(t, ".", ".")

	// Start Section: EKS Cluster

	test_structure.RunTestStage(t, "setup", func() {
		createEksClusterOptions(t, awsRegion, workingDir, filepath.Join(testFolder, EksClusterModulePath))
	})

	defer test_structure.RunTestStage(t, "destroy_cluster", func() {
		terraform.Destroy(t, test_structure.LoadTerraformOptions(t, workingDir))
	})

	test_structure.RunTestStage(t, "deploy_cluster", func() {
		terraform.InitAndApply(t, test_structure.LoadTerraformOptions(t, workingDir))
	})

	// End Section: EKS Cluster

	// Start Section: Benchmark

	// This doesn't need to be stored, because all the dynamic information is captured in the eks cluster module
	// options. Given that, the benchmark options can be deterministically calculated across test stage runs.
	benchmarkOptions := createBenchmarkOptions(t, awsRegion, workingDir, filepath.Join(testFolder, BenchmarkModulePath))

	defer test_structure.RunTestStage(t, "destroy_benchmark", func() {
		terraform.Destroy(t, benchmarkOptions)
	})

	test_structure.RunTestStage(t, "run_benchmark", func() {
		// We repeatedly call apply on the benchmark module to repeat the trial multiple times, if we want to collect
		// results from more than 90 runs. This is due to a limitation of Fargate and Kubernetes Jobs, where completed
		// Pods from Jobs still take up a Fargate node even though nothing is running. This means that you can easily
		// blow through the Fargate limit without running anything!
		// To work around this, we destroy the Kubernetes Job for each trial so that all the Pods get culled at the end.
		terraform.Init(t, benchmarkOptions)
		for i := 0; i < NumBenchmarks; i++ {
			logger.Logf(t, "Running trial %d/%d", i+1, NumBenchmarks)
			start := time.Now()
			runBenchmarkTrial(t, benchmarkOptions, test_structure.LoadKubectlOptions(t, workingDir))
			elapsed := time.Now().Sub(start)
			logger.Logf(t, "Trial %d took %v to run", i+1, elapsed)
		}
	})

	test_structure.RunTestStage(t, "collect_benchmark_results", func() {
		results := scanTable(t, awsRegion)
		summarizeResults(t, awsRegion, results)
	})

	// End Section: Benchmark
}

func createEksClusterOptions(t *testing.T, awsRegion string, workingDir string, terraformDir string) *terraform.Options {
	uniqueID := random.UniqueId()
	tmpKubeConfigPath := k8s.CopyHomeKubeConfigToTemp(t)
	kubectlOptions := k8s.NewKubectlOptions("", tmpKubeConfigPath, BenchmarkJobNamespace)
	name := "eks-fargate-procinfo-" + uniqueID
	usableZones := getUsableAvailabilityZones(t, awsRegion)

	eksClusterOptions := &terraform.Options{
		TerraformDir: terraformDir,
		Vars: map[string]interface{}{
			"aws_region":                  awsRegion,
			"vpc_name":                    name,
			"availability_zone_whitelist": usableZones,
			"eks_cluster_name":            name,
			"configure_kubectl":           "1",
			"kubectl_config_path":         tmpKubeConfigPath,
		},
	}

	test_structure.SaveTerraformOptions(t, workingDir, eksClusterOptions)
	test_structure.SaveString(t, workingDir, "uniqueID", uniqueID)
	test_structure.SaveKubectlOptions(t, workingDir, kubectlOptions)

	return eksClusterOptions
}

func createBenchmarkOptions(t *testing.T, awsRegion string, workingDir string, terraformDir string) *terraform.Options {
	eksClusterOptions := test_structure.LoadTerraformOptions(t, workingDir)
	eksClusterName := terraform.Output(t, eksClusterOptions, "eks_cluster_name")
	eksOpenIdArn := terraform.Output(t, eksClusterOptions, "eks_openid_connect_provider_arn")
	eksOpenIdUrl := terraform.Output(t, eksClusterOptions, "eks_openid_connect_provider_url")
	benchmarkOptions := &terraform.Options{
		TerraformDir: terraformDir,
		Vars: map[string]interface{}{
			"aws_region":                      awsRegion,
			"eks_cluster_name":                eksClusterName,
			"eks_openid_connect_provider_arn": eksOpenIdArn,
			"eks_openid_connect_provider_url": eksOpenIdUrl,
		},
	}
	return benchmarkOptions
}

// runBenchmarkTrial applies the benchmark module, which will provision the DynamoDB table and Kubernetes Job. Then,
// repeatedly polls the Kubernetes API until the Job finishes. Once the Job finishes, destroy just the Job to prepare
// for the next run.
func runBenchmarkTrial(t *testing.T, options *terraform.Options, kubectlOptions *k8s.KubectlOptions) {
	// Destroy just the kubernetes job at the end of each benchmark trial so that all the Pods and Fargate instances are
	// culled.
	defer func() {
		destroyOptions := &terraform.Options{
			TerraformDir: options.TerraformDir,
			Vars:         options.Vars,
			Targets:      []string{"kubernetes_job.cpu_benchmark"},
		}
		terraform.Destroy(t, destroyOptions)
	}()

	terraform.Apply(t, options)
	retry.DoWithRetry(
		t,
		"waiting for benchmark job to finish",
		MaxRetries,
		SleepBetweenTries,
		func() (string, error) {
			job := getJob(t, kubectlOptions, BenchmarkJobName)
			conditions := job.Status.Conditions
			if len(conditions) == 0 {
				k8s.RunKubectl(t, kubectlOptions, "get", "jobs", "-n", "default")
				return "", fmt.Errorf("Job has not finished yet")
			}
			if len(conditions) != 1 {
				t.Fatalf("Unexpected job condition: %v", conditions)
			}
			status := conditions[0].Type
			if status != batchv1.JobComplete {
				k8s.RunKubectl(t, kubectlOptions, "get", "jobs", "-n", "default")
				return "", fmt.Errorf("Job did not successfully complete: %s", status)
			}
			return "", nil
		},
	)
}

func getJob(t *testing.T, options *k8s.KubectlOptions, jobName string) *batchv1.Job {
	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, options)
	require.NoError(t, err)
	job, err := clientset.BatchV1().Jobs(options.Namespace).Get(jobName, metav1.GetOptions{})
	require.NoError(t, err)
	return job
}

func scanTable(t *testing.T, awsRegion string) []map[string]*dynamodb.AttributeValue {
	items := []map[string]*dynamodb.AttributeValue{}
	clt := aws.NewDynamoDBClient(t, awsRegion)
	input := &dynamodb.ScanInput{TableName: awsgo.String(BenchmarkTableName)}
	output, err := clt.Scan(input)
	require.NoError(t, err)
	items = append(items, output.Items...)
	for output.LastEvaluatedKey != nil {
		input = input.SetExclusiveStartKey(output.LastEvaluatedKey)
		output, err := clt.Scan(input)
		require.NoError(t, err)
		items = append(items, output.Items...)
	}
	return items
}

// summarizeResults summarizes the results from the Benchmark trials by providing the following information for each CPU
// model found:
// - The number of instances seen in the benchmark
// - The minimum runtime seen from the benchmark
func summarizeResults(t *testing.T, awsRegion string, results []map[string]*dynamodb.AttributeValue) {
	counter := map[string]int{}
	runTimes := map[string][]float32{}
	for _, item := range results {
		procInfo := awsgo.StringValue(item[BenchmarkTableProcInfoKey].S)
		runTimeStr := awsgo.StringValue(item[BenchmarkTableRunTimeKey].N)
		runTimeFlt64, err := strconv.ParseFloat(runTimeStr, 32)
		require.NoError(t, err)
		runTimeFlt := float32(runTimeFlt64)
		oldVal, hasKey := counter[procInfo]
		if hasKey {
			counter[procInfo] = oldVal + 1
			runTimes[procInfo] = append(runTimes[procInfo], runTimeFlt)
		} else {
			counter[procInfo] = 1
			runTimes[procInfo] = []float32{runTimeFlt}
		}
	}

	logger.Logf(t, "Processors:")
	for procInfo, count := range counter {
		logger.Logf(t, "\t%s : %d", procInfo, count)
	}
	plotProcessors(t, counter, "Processors", filepath.Join(ChartOutputPath, awsRegion, "processors.png"))

	logger.Logf(t, "Min runtime of each processor:")
	for procInfo, data := range runTimes {
		minRunTime := data[0]
		for _, v := range data {
			if v < minRunTime {
				minRunTime = v
			}
		}
		logger.Logf(t, "\t%s : %f", procInfo, minRunTime)
	}
}

// getUsableAvailabilityZones returns the list of availability zones that work with EKS given a region.
func getUsableAvailabilityZones(t *testing.T, region string) []string {
	// us-east-1e currently does not have capacity to support EKS
	AVAILABILITY_ZONE_BLACKLIST := []string{
		"us-east-1e",
	}

	usableZones := []string{}
	zones := aws.GetAvailabilityZones(t, region)
	for _, zone := range zones {
		// If zone is not in blacklist, include it
		if !collections.ListContains(AVAILABILITY_ZONE_BLACKLIST, zone) {
			usableZones = append(usableZones, zone)
		}
	}
	return usableZones
}

// plotProcessors plots a bar chart of the processor models seen from the benchmark to the provided chartPath.
func plotProcessors(t *testing.T, counter map[string]int, label string, chartPath string) {
	data := plotter.Values{}
	labels := []string{}
	for label, _ := range counter {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	for _, label := range labels {
		data = append(data, float64(counter[label]))
	}

	bars, err := plotter.NewBarChart(data, vg.Points(20))
	require.NoError(t, err)
	bars.LineStyle.Width = vg.Length(0)
	bars.Color = plotutil.Color(0)

	p, err := plot.New()
	require.NoError(t, err)
	p.Title.Text = label
	p.Y.Label.Text = "Number of Instances"
	p.Add(bars)
	p.NominalX(labels...)

	dir := filepath.Dir(chartPath)
	if !files.IsDir(dir) {
		require.NoError(t, os.MkdirAll(dir, 0755))
	}
	require.NoError(t, p.Save(11*vg.Inch, 6*vg.Inch, chartPath))
}
