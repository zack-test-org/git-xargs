package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	awsgo "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/collections"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NumBenchmarks = 1

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
	//"us-east-1",
	"us-east-2",
	//"eu-west-1",
	//"ap-northeast-1",
}

// For each Fargate region:
// - Launch a Fargate cluster
// - Run a benchmark job
// - Collect results into a graph
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
	//os.Setenv("SKIP_setup", "true")
	//os.Setenv("SKIP_deploy_cluster", "true")
	//os.Setenv("SKIP_run_benchmark", "true")
	//os.Setenv("SKIP_collect_benchmark_results", "true")
	//os.Setenv("SKIP_destroy_benchmark", "true")
	//os.Setenv("SKIP_destroy_cluster", "true")

	workingDir := filepath.Join("stages", awsRegion)
	testFolder := test_structure.CopyTerraformFolderToTemp(t, ".", ".")

	test_structure.RunTestStage(t, "setup", func() {
		createEksClusterOptions(t, awsRegion, workingDir, filepath.Join(testFolder, EksClusterModulePath))
	})

	defer test_structure.RunTestStage(t, "destroy_cluster", func() {
		terraform.Destroy(t, test_structure.LoadTerraformOptions(t, workingDir))
	})

	test_structure.RunTestStage(t, "deploy_cluster", func() {
		terraform.InitAndApply(t, test_structure.LoadTerraformOptions(t, workingDir))
	})

	// This doesn't need to be stored, because all the dynamic information is captured in the eks cluster module
	// options. Given that, the benchmark options can be deterministically calculated across test stage runs.
	benchmarkOptions := createBenchmarkOptions(t, awsRegion, workingDir, filepath.Join(testFolder, BenchmarkModulePath))

	defer test_structure.RunTestStage(t, "destroy_benchmark", func() {
		terraform.Destroy(t, benchmarkOptions)
	})

	test_structure.RunTestStage(t, "run_benchmark", func() {
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
		summarizeResults(t, results)
	})
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

func runBenchmarkTrial(t *testing.T, options *terraform.Options, kubectlOptions *k8s.KubectlOptions) {
	// Destroy just the kubernetes job at the end of each benchmark trial
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

func summarizeResults(t *testing.T, results []map[string]*dynamodb.AttributeValue) {
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

	logger.Logf(t, "Histogram of processors:")
	for procInfo, count := range counter {
		logger.Logf(t, "\t%s : %d", procInfo, count)
	}

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
