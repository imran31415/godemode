package runner

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/imran31415/godemode/benchmark/agents"
	"github.com/imran31415/godemode/benchmark/scenarios"
	"github.com/imran31415/godemode/benchmark/systems/database"
	"github.com/imran31415/godemode/benchmark/systems/email"
	"github.com/imran31415/godemode/benchmark/systems/filesystem"
	"github.com/imran31415/godemode/benchmark/systems/graph"
	"github.com/imran31415/godemode/benchmark/systems/security"
)

// BenchmarkRunner orchestrates benchmark execution
type BenchmarkRunner struct {
	scenario         *scenarios.SupportScenario
	codeModeAgent    *agents.CodeModeAgent
	functionAgent    *agents.FunctionCallingAgent
	testEnvironment  *scenarios.TestEnvironment
}

// TaskResult contains results for a single task
type TaskResult struct {
	TaskName          string
	TaskComplexity    string
	ExpectedOps       int

	// Code Mode Results
	CodeModeMetrics   *agents.AgentMetrics
	CodeModeVerified  bool
	CodeModeErrors    []string

	// Function Calling Results
	FunctionMetrics   *agents.AgentMetrics
	FunctionVerified  bool
	FunctionErrors    []string
}

// BenchmarkReport contains the full comparison
type BenchmarkReport struct {
	ScenarioName  string
	StartTime     time.Time
	EndTime       time.Time
	TotalDuration time.Duration
	Results       []TaskResult
}

// NewBenchmarkRunner creates a new benchmark runner
func NewBenchmarkRunner(fixturesPath string) (*BenchmarkRunner, error) {
	// Initialize test environment
	env, err := setupTestEnvironment(fixturesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to setup test environment: %w", err)
	}

	// Create scenario
	scenario := scenarios.NewSupportScenario()

	// Create agents
	codeModeAgent := agents.NewCodeModeAgent(env)
	functionAgent := agents.NewFunctionCallingAgent(env)

	return &BenchmarkRunner{
		scenario:        scenario,
		codeModeAgent:   codeModeAgent,
		functionAgent:   functionAgent,
		testEnvironment: env,
	}, nil
}

// setupTestEnvironment initializes all systems
func setupTestEnvironment(fixturesPath string) (*scenarios.TestEnvironment, error) {
	// Email system
	emailSystem := email.NewEmailSystem(fixturesPath+"/emails", fixturesPath+"/emails/sent")

	// Database
	db, err := database.NewSQLiteDB(":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Knowledge graph
	graphDB, err := graph.NewKnowledgeGraph(fixturesPath + "/graph_data")
	if err != nil {
		return nil, fmt.Errorf("failed to create knowledge graph: %w", err)
	}

	// Log system
	logSystem := filesystem.NewLogSystem(fixturesPath + "/logs")

	// Config system
	configSystem := filesystem.NewConfigSystem(fixturesPath + "/configs")

	// Security monitor
	securityMonitor := security.NewSecurityMonitor()

	return &scenarios.TestEnvironment{
		EmailSystem:     emailSystem,
		Database:        db,
		Graph:           graphDB,
		LogSystem:       logSystem,
		ConfigSystem:    configSystem,
		SecurityMonitor: securityMonitor,
	}, nil
}

// RunBenchmark executes all tasks with both agents
func (br *BenchmarkRunner) RunBenchmark(ctx context.Context) (*BenchmarkReport, error) {
	report := &BenchmarkReport{
		ScenarioName: br.scenario.Name,
		StartTime:    time.Now(),
		Results:      make([]TaskResult, 0, len(br.scenario.Tasks)),
	}

	// Filter tasks based on environment variable
	taskFilter := os.Getenv("TASK_FILTER")
	var tasksToRun []scenarios.Task

	if taskFilter != "" && taskFilter != "all" {
		fmt.Printf("Running only %s complexity tasks\n", taskFilter)
		for _, task := range br.scenario.Tasks {
			if task.Complexity == taskFilter {
				tasksToRun = append(tasksToRun, task)
			}
		}
	} else {
		tasksToRun = br.scenario.Tasks
	}

	if len(tasksToRun) == 0 {
		return nil, fmt.Errorf("no tasks match filter: %s", taskFilter)
	}

	for _, task := range tasksToRun {
		fmt.Printf("\n=== Running Task: %s ===\n", task.Name)

		result, err := br.runTask(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("task %s failed: %w", task.Name, err)
		}

		report.Results = append(report.Results, *result)
	}

	report.EndTime = time.Now()
	report.TotalDuration = report.EndTime.Sub(report.StartTime)

	return report, nil
}

// runTask executes a single task with both agents
func (br *BenchmarkRunner) runTask(ctx context.Context, task scenarios.Task) (*TaskResult, error) {
	result := &TaskResult{
		TaskName:       task.Name,
		TaskComplexity: task.Complexity,
		ExpectedOps:    task.ExpectedOps,
	}

	// Reset environment before starting new task
	if err := br.resetEnvironment(); err != nil {
		return nil, fmt.Errorf("initial environment reset failed: %w", err)
	}

	// Run setup
	if err := task.SetupFunc(br.testEnvironment); err != nil {
		return nil, fmt.Errorf("setup failed: %w", err)
	}

	fmt.Println("\n--- Running CODE MODE Agent ---")

	// Run with Code Mode Agent
	codeModeMetrics, err := br.codeModeAgent.RunTask(ctx, task, br.testEnvironment)
	result.CodeModeMetrics = codeModeMetrics

	if err != nil {
		result.CodeModeErrors = append(result.CodeModeErrors, err.Error())
	}

	// Verify Code Mode results
	verified, verifyErrors := task.VerificationFunc(br.testEnvironment)
	result.CodeModeVerified = verified
	result.CodeModeErrors = append(result.CodeModeErrors, verifyErrors...)

	// Print Code Mode code visibility
	fmt.Println("\nGenerated Code:")
	codeLogs := br.codeModeAgent.GetCodeLogs()
	if len(codeLogs) > 0 {
		latestLog := codeLogs[len(codeLogs)-1]
		fmt.Printf("Task: %s\n", latestLog.Task)
		fmt.Printf("Success: %v\n", latestLog.Success)
		fmt.Printf("Compile Time: %v\n", latestLog.CompileTime)
		fmt.Printf("Execute Time: %v\n", latestLog.ExecuteTime)
		fmt.Println("\nCode:")
		fmt.Println(latestLog.Code)
		if latestLog.Output != "" {
			fmt.Println("\nOutput:")
			fmt.Println(latestLog.Output)
		}
	}

	// Reset environment for function calling
	if err := br.resetEnvironment(); err != nil {
		return nil, fmt.Errorf("environment reset failed: %w", err)
	}

	// Re-run setup for function calling
	if err := task.SetupFunc(br.testEnvironment); err != nil {
		return nil, fmt.Errorf("setup failed on second run: %w", err)
	}

	fmt.Println("\n--- Running FUNCTION CALLING Agent ---")

	// Run with Function Calling Agent
	functionMetrics, err := br.functionAgent.RunTask(ctx, task, br.testEnvironment)
	result.FunctionMetrics = functionMetrics

	if err != nil {
		result.FunctionErrors = append(result.FunctionErrors, err.Error())
	}

	// Verify Function Calling results
	verified, verifyErrors = task.VerificationFunc(br.testEnvironment)
	result.FunctionVerified = verified
	result.FunctionErrors = append(result.FunctionErrors, verifyErrors...)

	// Print Function Calling details
	fmt.Println("\nFunction Calls:")
	functionCalls := br.functionAgent.GetFunctionCalls()
	for i, call := range functionCalls {
		fmt.Printf("%d. %s (Duration: %v)\n", i+1, call.ToolName, call.Duration)
	}

	return result, nil
}

// resetEnvironment clears all data from the test environment
func (br *BenchmarkRunner) resetEnvironment() error {
	// Clear database (truncate tables)
	if err := br.testEnvironment.Database.Close(); err != nil {
		return err
	}

	// Recreate database
	db, err := database.NewSQLiteDB(":memory:")
	if err != nil {
		return err
	}
	br.testEnvironment.Database = db

	// Note: In a full implementation, we'd also reset the graph, logs, etc.
	// For now, just resetting the database is sufficient for the demo

	return nil
}

// PrintReport prints a formatted comparison report
func (r *BenchmarkReport) PrintReport() {
	fmt.Println("\n" + repeat("=", 100))
	fmt.Printf("BENCHMARK REPORT: %s\n", r.ScenarioName)
	fmt.Println(repeat("=", 100))
	fmt.Printf("Total Duration: %v\n", r.TotalDuration)
	fmt.Printf("Tasks Completed: %d\n\n", len(r.Results))

	for i, result := range r.Results {
		fmt.Printf("%d. %s (%s complexity, %d expected ops)\n",
			i+1, result.TaskName, result.TaskComplexity, result.ExpectedOps)
		fmt.Println("   " + repeat("-", 90))

		// Code Mode Results
		fmt.Println("   CODE MODE:")
		if result.CodeModeMetrics != nil {
			fmt.Printf("     Duration: %v\n", result.CodeModeMetrics.TotalDuration)
			fmt.Printf("     Tokens: %d\n", result.CodeModeMetrics.TokensUsed)
			fmt.Printf("     API Calls: %d\n", result.CodeModeMetrics.APICallCount)
			fmt.Printf("     Operations: %d\n", result.CodeModeMetrics.OperationsCount)
			fmt.Printf("     Success: %v\n", result.CodeModeMetrics.Success)
			fmt.Printf("     Verified: %v\n", result.CodeModeVerified)
		}
		if len(result.CodeModeErrors) > 0 {
			fmt.Printf("     Errors: %v\n", result.CodeModeErrors)
		}

		// Function Calling Results
		fmt.Println("   FUNCTION CALLING:")
		if result.FunctionMetrics != nil {
			fmt.Printf("     Duration: %v\n", result.FunctionMetrics.TotalDuration)
			fmt.Printf("     Tokens: %d\n", result.FunctionMetrics.TokensUsed)
			fmt.Printf("     API Calls: %d\n", result.FunctionMetrics.APICallCount)
			fmt.Printf("     Operations: %d\n", result.FunctionMetrics.OperationsCount)
			fmt.Printf("     Success: %v\n", result.FunctionMetrics.Success)
			fmt.Printf("     Verified: %v\n", result.FunctionVerified)
		}
		if len(result.FunctionErrors) > 0 {
			fmt.Printf("     Errors: %v\n", result.FunctionErrors)
		}

		// Comparison
		if result.CodeModeMetrics != nil && result.FunctionMetrics != nil {
			fmt.Println("   COMPARISON:")

			// Duration comparison
			if result.CodeModeMetrics.TotalDuration < result.FunctionMetrics.TotalDuration {
				speedup := float64(result.FunctionMetrics.TotalDuration) / float64(result.CodeModeMetrics.TotalDuration)
				fmt.Printf("     Code Mode was %.2fx faster\n", speedup)
			} else {
				speedup := float64(result.CodeModeMetrics.TotalDuration) / float64(result.FunctionMetrics.TotalDuration)
				fmt.Printf("     Function Calling was %.2fx faster\n", speedup)
			}

			// Token comparison
			tokenDiff := result.FunctionMetrics.TokensUsed - result.CodeModeMetrics.TokensUsed
			if tokenDiff > 0 {
				fmt.Printf("     Code Mode used %d fewer tokens\n", tokenDiff)
			} else {
				fmt.Printf("     Function Calling used %d fewer tokens\n", -tokenDiff)
			}

			// API call comparison
			apiDiff := result.FunctionMetrics.APICallCount - result.CodeModeMetrics.APICallCount
			if apiDiff > 0 {
				fmt.Printf("     Code Mode made %d fewer API calls\n", apiDiff)
			} else {
				fmt.Printf("     Function Calling made %d fewer API calls\n", -apiDiff)
			}
		}

		fmt.Println()
	}

	fmt.Println(repeat("=", 100))
}

// repeat repeats a string n times
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// RepeatString is an exported version of repeat
func RepeatString(s string, count int) string {
	return repeat(s, count)
}

// ResetEnvironment resets the test environment (exported for use by other benchmarks)
func ResetEnvironment(env *scenarios.TestEnvironment) error {
	// Close and recreate database
	if err := env.Database.Close(); err != nil {
		return err
	}

	db, err := database.NewSQLiteDB(":memory:")
	if err != nil {
		return err
	}
	env.Database = db

	// Reset security monitor
	if env.SecurityMonitor != nil {
		env.SecurityMonitor = security.NewSecurityMonitor()
	}

	return nil
}
