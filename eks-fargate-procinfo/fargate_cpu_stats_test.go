package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	awsgo "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NumBenchmarks = 5

	KubectlConfigPath = "/var/folders/n2/pljz6dq52bd1ksmw23qyr3sr0000gn/T/333325088"
	InputsTfVarsPath  = "./terraform.tfvars"
	Region            = "us-east-1"

	BenchmarkModulePath = "./benchmark"
	BenchmarkJobName    = "cpu-benchmark"

	BenchmarkTableName        = "EKSFargateCPUBenchmark"
	BenchmarkTableProcInfoKey = "ProcInfo"
	BenchmarkTableRunTimeKey  = "RunTime"

	// Wait for up to 1 hour (60 tries, 1 minute between tries)
	MaxRetries        = 60
	SleepBetweenTries = 1 * time.Minute
)

func TestFargateCPUInfo(t *testing.T) {
	os.Setenv("SKIP_run_benchmark", "true")
	os.Setenv("SKIP_collect_benchmark_results", "true")
	//os.Setenv("SKIP_destroy", "true")

	absInputsTfVarsPath, err := filepath.Abs(InputsTfVarsPath)
	require.NoError(t, err)

	options := &terraform.Options{
		TerraformDir: BenchmarkModulePath,
		VarFiles:     []string{absInputsTfVarsPath},
	}

	defer test_structure.RunTestStage(t, "destroy", func() {
		terraform.Destroy(t, options)
	})

	test_structure.RunTestStage(t, "run_benchmark", func() {
		terraform.Init(t, options)
		for i := 0; i < NumBenchmarks; i++ {
			logger.Logf(t, "Running trial %d/%d", i+1, NumBenchmarks)
			start := time.Now()
			runBenchmarkTrial(t, options)
			elapsed := time.Now().Sub(start)
			logger.Logf(t, "Trial %d took %v to run", i+1, elapsed)
		}
	})

	test_structure.RunTestStage(t, "collect_benchmark_results", func() {
		results := scanTable(t)
		summarizeResults(t, results)
	})
}

func runBenchmarkTrial(t *testing.T, options *terraform.Options) {
	// Destroy just the kubernetes job at the end of each benchmark trial
	defer func() {
		destroyOptions := &terraform.Options{
			TerraformDir: options.TerraformDir,
			VarFiles:     options.VarFiles,
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
			kubectlOptions := k8s.NewKubectlOptions("", KubectlConfigPath, "default")
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

func scanTable(t *testing.T) []map[string]*dynamodb.AttributeValue {
	items := []map[string]*dynamodb.AttributeValue{}
	clt := aws.NewDynamoDBClient(t, Region)
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
