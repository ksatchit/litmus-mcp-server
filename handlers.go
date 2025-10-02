package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Utility functions for type conversion
func getStringFromArgs(args map[string]interface{}, key string, defaultValue string) string {
	if val, ok := args[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getIntFromArgs(args map[string]interface{}, key string, defaultValue int) int {
	if val, ok := args[key]; ok {
		if num, ok := val.(float64); ok {
			return int(num)
		}
	}
	return defaultValue
}

func getBoolFromArgs(args map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := args[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func getMapFromArgs(args map[string]interface{}, key string) map[string]interface{} {
	if val, ok := args[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

func getSliceFromArgs(args map[string]interface{}, key string) []interface{} {
	if val, ok := args[key]; ok {
		if s, ok := val.([]interface{}); ok {
			return s
		}
	}
	return nil
}

// Utility function to get nested string values
func getNestedString(obj map[string]interface{}, keys ...string) string {
	current := obj
	for i, key := range keys {
		if i == len(keys)-1 {
			if val, ok := current[key]; ok {
				if str, ok := val.(string); ok {
					return str
				}
			}
			return ""
		}
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return ""
		}
	}
	return ""
}

// Tool handler implementations

func (s *LitmusChaosServer) listChaosExperiments(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query := `
		query ListExperiment($projectID: ID!, $request: ListExperimentRequest!) {
			listExperiment(projectID: $projectID, request: $request) {
				totalNoOfExperiments
				experiments {
					projectID
					experimentID
					name
					description
					experimentType
					cronSyntax
					isCustomExperiment
					tags
					updatedAt
					createdAt
					infra {
						infraID
						name
						environmentID
						isActive
						isInfraConfirmed
						platformName
					}
					recentExperimentRunDetails {
						experimentRunID
						phase
						resiliencyScore
						updatedAt
						runSequence
					}
					createdBy {
						username
						email
					}
				}
			}
		}
	`

	request := map[string]interface{}{
		"pagination": map[string]interface{}{
			"page":  getIntFromArgs(getMapFromArgs(args, "pagination"), "page", 0),
			"limit": getIntFromArgs(getMapFromArgs(args, "pagination"), "limit", 10),
		},
	}

	if filter := getMapFromArgs(args, "filter"); filter != nil {
		request["filter"] = filter
	}

	variables := map[string]interface{}{
		"request": request,
	}

	data, err := s.graphqlRequest(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	listExperiment := result["listExperiment"].(map[string]interface{})
	experiments := listExperiment["experiments"].([]interface{})

	formattedExperiments := make([]map[string]interface{}, len(experiments))
	for i, exp := range experiments {
		experiment := exp.(map[string]interface{})
		
		var infrastructure map[string]interface{}
		if infra := experiment["infra"]; infra != nil {
			infraMap := infra.(map[string]interface{})
			infrastructure = map[string]interface{}{
				"id":          infraMap["infraID"],
				"name":        infraMap["name"],
				"environment": infraMap["environmentID"],
				"active":      infraMap["isActive"],
				"confirmed":   infraMap["isInfraConfirmed"],
				"platform":    infraMap["platformName"],
			}
		}

		var recentRun map[string]interface{}
		if runs := experiment["recentExperimentRunDetails"]; runs != nil {
			runsSlice := runs.([]interface{})
			if len(runsSlice) > 0 {
				run := runsSlice[0].(map[string]interface{})
				recentRun = map[string]interface{}{
					"id":             run["experimentRunID"],
					"status":         run["phase"],
					"resiliencyScore": run["resiliencyScore"],
					"lastRun":        run["updatedAt"],
					"sequence":       run["runSequence"],
				}
			}
		}

		formattedExperiments[i] = map[string]interface{}{
			"id":           experiment["experimentID"],
			"name":         experiment["name"],
			"description":  experiment["description"],
			"type":         experiment["experimentType"],
			"isCustom":     experiment["isCustomExperiment"],
			"schedule":     experiment["cronSyntax"],
			"tags":         experiment["tags"],
			"infrastructure": infrastructure,
			"recentRun":    recentRun,
			"createdBy":    getNestedString(experiment, "createdBy", "username"),
			"createdAt":    experiment["createdAt"],
			"updatedAt":    experiment["updatedAt"],
		}
	}

	response := map[string]interface{}{
		"summary":          fmt.Sprintf("Found %v chaos experiments", listExperiment["totalNoOfExperiments"]),
		"totalExperiments": listExperiment["totalNoOfExperiments"],
		"experiments":      formattedExperiments,
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) getChaosExperiment(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query := `
		query GetExperiment($projectID: ID!, $experimentID: String!) {
			getExperiment(projectID: $projectID, experimentID: $experimentID) {
				experimentDetails {
					projectID
					experimentID
					name
					description
					experimentManifest
					experimentType
					cronSyntax
					isCustomExperiment
					weightages {
						faultName
						weightage
					}
					tags
					infra {
						infraID
						name
						description
						environmentID
						platformName
						isActive
						infraScope
						version
						noOfExperiments
						noOfExperimentRuns
					}
					createdBy {
						username
						email
					}
					updatedBy {
						username
						email
					}
					createdAt
					updatedAt
				}
				averageResiliencyScore
			}
		}
	`

	experimentID := getStringFromArgs(args, "experimentId", "")
	if experimentID == "" {
		return nil, fmt.Errorf("experimentId is required")
	}

	variables := map[string]interface{}{
		"experimentID": experimentID,
	}

	data, err := s.graphqlRequest(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	getExperiment := result["getExperiment"].(map[string]interface{})
	exp := getExperiment["experimentDetails"].(map[string]interface{})

	var faults []map[string]interface{}
	if weightages := exp["weightages"]; weightages != nil {
		weightagesSlice := weightages.([]interface{})
		faults = make([]map[string]interface{}, len(weightagesSlice))
		for i, w := range weightagesSlice {
			weight := w.(map[string]interface{})
			faults[i] = map[string]interface{}{
				"name":   weight["faultName"],
				"weight": weight["weightage"],
			}
		}
	}

	infrastructure := map[string]interface{}{}
	if infra := exp["infra"]; infra != nil {
		infraMap := infra.(map[string]interface{})
		infrastructure = map[string]interface{}{
			"id":               infraMap["infraID"],
			"name":             infraMap["name"],
			"description":      infraMap["description"],
			"environment":      infraMap["environmentID"],
			"platform":         infraMap["platformName"],
			"active":           infraMap["isActive"],
			"scope":            infraMap["infraScope"],
			"version":          infraMap["version"],
			"totalExperiments": infraMap["noOfExperiments"],
			"totalRuns":        infraMap["noOfExperimentRuns"],
		}
	}

	response := map[string]interface{}{
		"experiment": map[string]interface{}{
			"id":                     exp["experimentID"],
			"name":                   exp["name"],
			"description":            exp["description"],
			"type":                   exp["experimentType"],
			"isCustom":               exp["isCustomExperiment"],
			"schedule":               exp["cronSyntax"],
			"manifest":               exp["experimentManifest"],
			"averageResiliencyScore": getExperiment["averageResiliencyScore"],
			"faults":                 faults,
			"tags":                   exp["tags"],
			"infrastructure":         infrastructure,
			"createdBy":              getNestedString(exp, "createdBy", "username"),
			"updatedBy":              getNestedString(exp, "updatedBy", "username"),
			"createdAt":              exp["createdAt"],
			"updatedAt":              exp["updatedAt"],
		},
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

// TODO: not being used in the main.go as a valid tool until further refinement for manifest accuracy
func (s *LitmusChaosServer) createChaosExperiment(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name := getStringFromArgs(args, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	faults := getSliceFromArgs(args, "faults")
	if len(faults) == 0 {
		return nil, fmt.Errorf("at least one fault is required")
	}

	infraId := getStringFromArgs(args, "infraId", s.config.DefaultInfraID)
	if infraId == "" {
		return nil, fmt.Errorf("infraId is required")
	}

	// Build fault manifests
	faultManifests := make([]map[string]interface{}, len(faults))
	weightages := make([]map[string]interface{}, len(faults))

	for i, fault := range faults {
		faultMap := fault.(map[string]interface{})
		faultName := getStringFromArgs(faultMap, "name", "")
		weight := getIntFromArgs(faultMap, "weight", 5)
		targetApp := getStringFromArgs(faultMap, "targetApp", "")
		duration := getStringFromArgs(faultMap, "duration", "60")
		parameters := getMapFromArgs(faultMap, "parameters")

		env := []map[string]interface{}{
			{"name": "TOTAL_CHAOS_DURATION", "value": duration},
			{"name": "TARGET_PODS", "value": targetApp},
		}

		if parameters != nil {
			for key, value := range parameters {
				env = append(env, map[string]interface{}{
					"name":  strings.ToUpper(key),
					"value": fmt.Sprintf("%v", value),
				})
			}
		}

		faultManifests[i] = map[string]interface{}{
			"name":   faultName,
			"weight": weight,
			"spec": map[string]interface{}{
				"components": map[string]interface{}{
					"env": env,
				},
			},
		}

		weightages[i] = map[string]interface{}{
			"faultName":  faultName,
			"weightage":  weight,
		}
	}

	// Create workflow manifest
	workflowSteps := make([]map[string]interface{}, len(faultManifests))
	templates := make([]map[string]interface{}, len(faultManifests)+1)

	for i, fault := range faultManifests {
		faultName := fault["name"].(string)
		workflowSteps[i] = map[string]interface{}{
			"name":     faultName,
			"template": faultName,
		}

		templates[i+1] = map[string]interface{}{
			"name": faultName,
			"container": map[string]interface{}{
				"image":   "litmuschaos/go-runner:latest",
				"command": []string{"chaos-runner"},
				"args":    []string{"-name", faultName},
				"env":     fault["spec"].(map[string]interface{})["components"].(map[string]interface{})["env"],
			},
		}
	}

	templates[0] = map[string]interface{}{
		"name": "chaos-experiment",
		"steps": [][]map[string]interface{}{workflowSteps},
	}

	manifest := map[string]interface{}{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Workflow",
		"metadata": map[string]interface{}{
			"name":      strings.ToLower(strings.ReplaceAll(name, " ", "-")),
			"namespace": "litmus",
		},
		"spec": map[string]interface{}{
			"entrypoint": "chaos-experiment",
			"templates":  templates,
		},
	}

	manifestJSON, _ := json.Marshal(manifest)

	mutation := `
		mutation CreateChaosExperiment($request: ChaosExperimentRequest!, $projectID: ID!) {
			createChaosExperiment(request: $request, projectID: $projectID) {
				experimentID
				experimentName
				experimentDescription
				cronSyntax
				isCustomExperiment
				tags
			}
		}
	`

	schedule := "none"
	if scheduleMap := getMapFromArgs(args, "schedule"); scheduleMap != nil {
		if cronExpr := getStringFromArgs(scheduleMap, "cronExpression", ""); cronExpr != "" {
			schedule = cronExpr
		}
	}

	tags := []string{}
	if tagsSlice := getSliceFromArgs(args, "tags"); tagsSlice != nil {
		tags = make([]string, len(tagsSlice))
		for i, tag := range tagsSlice {
			tags[i] = fmt.Sprintf("%v", tag)
		}
	}

	request := map[string]interface{}{
		"experimentName":        name,
		"experimentDescription": getStringFromArgs(args, "description", "Created via MCP Server"),
		"infraID":               infraId,
		"experimentManifest":    string(manifestJSON),
		"cronSyntax":            schedule,
		"isCustomExperiment":    true,
		"weightages":            weightages,
		"tags":                  tags,
	}

	variables := map[string]interface{}{
		"request": request,
	}

	data, err := s.graphqlRequest(ctx, mutation, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	createResult := result["createChaosExperiment"].(map[string]interface{})

	faultNames := make([]string, len(faultManifests))
	for i, f := range faultManifests {
		faultNames[i] = f["name"].(string)
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Chaos experiment '%s' created successfully", name),
		"experiment": map[string]interface{}{
			"id":          createResult["experimentID"],
			"name":        createResult["experimentName"],
			"description": createResult["experimentDescription"],
			"schedule":    createResult["cronSyntax"],
			"isCustom":    createResult["isCustomExperiment"],
			"tags":        createResult["tags"],
			"faults":      faultNames,
		},
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) runChaosExperiment(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	experimentID := getStringFromArgs(args, "experimentId", "")
	if experimentID == "" {
		return nil, fmt.Errorf("experimentId is required")
	}

	mutation := `
		mutation RunChaosExperiment($experimentID: String!, $projectID: ID!) {
			runChaosExperiment(experimentID: $experimentID, projectID: $projectID) {
				notifyID
			}
		}
	`

	variables := map[string]interface{}{
		"experimentID": experimentID,
	}

	data, err := s.graphqlRequest(ctx, mutation, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	runResult := result["runChaosExperiment"].(map[string]interface{})

	response := map[string]interface{}{
		"success":      true,
		"message":      "Chaos experiment started successfully",
		"notifyId":     runResult["notifyID"],
		"experimentId": experimentID,
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) stopChaosExperiment(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	experimentID := getStringFromArgs(args, "experimentId", "")
	if experimentID == "" {
		return nil, fmt.Errorf("experimentId is required")
	}

	mutation := `
		mutation StopExperimentRuns(
			$projectID: ID!,
			$experimentID: String!,
			$experimentRunID: String,
			$notifyID: String
		) {
			stopExperimentRuns(
				projectID: $projectID,
				experimentID: $experimentID,
				experimentRunID: $experimentRunID,
				notifyID: $notifyID
			)
		}
	`

	variables := map[string]interface{}{
		"experimentID": experimentID,
	}

	if experimentRunID := getStringFromArgs(args, "experimentRunId", ""); experimentRunID != "" {
		variables["experimentRunID"] = experimentRunID
	}

	data, err := s.graphqlRequest(ctx, mutation, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	success := result["stopExperimentRuns"].(bool)
	message := "Failed to stop chaos experiment"
	if success {
		message = "Chaos experiment stopped successfully"
	}

	response := map[string]interface{}{
		"success":           success,
		"message":           message,
		"experimentId":      experimentID,
		"experimentRunId":   getStringFromArgs(args, "experimentRunId", ""),
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) listExperimentRuns(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query := `
		query ListExperimentRun($projectID: ID!, $request: ListExperimentRunRequest!) {
			listExperimentRun(projectID: $projectID, request: $request) {
				totalNoOfExperimentRuns
				experimentRuns {
					projectID
					experimentRunID
					experimentID
					experimentName
					phase
					resiliencyScore
					faultsPassed
					faultsFailed
					faultsAwaited
					faultsStopped
					totalFaults
					updatedAt
					createdAt
					runSequence
					infra {
						infraID
						name
						environmentID
						platformName
					}
					createdBy {
						username
					}
				}
			}
		}
	`

	request := map[string]interface{}{
		"pagination": map[string]interface{}{
			"page":  0,
			"limit": getIntFromArgs(args, "limit", 20),
		},
	}

	if experimentID := getStringFromArgs(args, "experimentId", ""); experimentID != "" {
		request["experimentIDs"] = []string{experimentID}
	}

	if status := getStringFromArgs(args, "status", ""); status != "" {
		request["filter"] = map[string]interface{}{
			"experimentStatus": status,
		}
	}

	variables := map[string]interface{}{
		"request": request,
	}

	data, err := s.graphqlRequest(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	listRuns := result["listExperimentRun"].(map[string]interface{})
	runs := listRuns["experimentRuns"].([]interface{})

	formattedRuns := make([]map[string]interface{}, len(runs))
	for i, run := range runs {
		runMap := run.(map[string]interface{})

		infrastructure := map[string]interface{}{}
		if infra := runMap["infra"]; infra != nil {
			infraMap := infra.(map[string]interface{})
			infrastructure = map[string]interface{}{
				"id":          infraMap["infraID"],
				"name":        infraMap["name"],
				"environment": infraMap["environmentID"],
				"platform":    infraMap["platformName"],
			}
		}

		formattedRuns[i] = map[string]interface{}{
			"id":             runMap["experimentRunID"],
			"experimentId":   runMap["experimentID"],
			"experimentName": runMap["experimentName"],
			"status":         runMap["phase"],
			"resiliencyScore": runMap["resiliencyScore"],
			"faultsSummary": map[string]interface{}{
				"passed":  runMap["faultsPassed"],
				"failed":  runMap["faultsFailed"],
				"awaited": runMap["faultsAwaited"],
				"stopped": runMap["faultsStopped"],
				"total":   runMap["totalFaults"],
			},
			"infrastructure": infrastructure,
			"sequence":       runMap["runSequence"],
			"createdBy":      getNestedString(runMap, "createdBy", "username"),
			"createdAt":      runMap["createdAt"],
			"updatedAt":      runMap["updatedAt"],
		}
	}

	response := map[string]interface{}{
		"summary":   fmt.Sprintf("Found %v experiment runs", listRuns["totalNoOfExperimentRuns"]),
		"totalRuns": listRuns["totalNoOfExperimentRuns"],
		"runs":      formattedRuns,
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) getExperimentRunDetails(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	experimentRunID := getStringFromArgs(args, "experimentRunId", "")
	if experimentRunID == "" {
		return nil, fmt.Errorf("experimentRunId is required")
	}

	query := `
		query GetExperimentRun(
			$projectID: ID!,
			$experimentRunID: ID,
			$notifyID: ID
		) {
			getExperimentRun(
				projectID: $projectID,
				experimentRunID: $experimentRunID,
				notifyID: $notifyID
			) {
				projectID
				experimentRunID
				experimentID
				experimentName
				experimentManifest
				phase
				resiliencyScore
				faultsPassed
				faultsFailed
				faultsAwaited
				faultsStopped
				faultsNa
				totalFaults
				executionData
				updatedAt
				createdAt
				runSequence
				infra {
					infraID
					name
					environmentID
					platformName
					version
				}
				createdBy {
					username
					email
				}
				updatedBy {
					username
					email
				}
			}
		}
	`

	variables := map[string]interface{}{
		"experimentRunID": experimentRunID,
	}

	data, err := s.graphqlRequest(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	run := result["getExperimentRun"].(map[string]interface{})

	infrastructure := map[string]interface{}{}
	if infra := run["infra"]; infra != nil {
		infraMap := infra.(map[string]interface{})
		infrastructure = map[string]interface{}{
			"id":          infraMap["infraID"],
			"name":        infraMap["name"],
			"environment": infraMap["environmentID"],
			"platform":    infraMap["platformName"],
			"version":     infraMap["version"],
		}
	}

	var executionData interface{}
	if getBoolFromArgs(args, "includeLogs", false) {
		if execData := run["executionData"]; execData != nil {
			if execStr, ok := execData.(string); ok && execStr != "" {
				json.Unmarshal([]byte(execStr), &executionData)
			}
		}
	}

	response := map[string]interface{}{
		"run": map[string]interface{}{
			"id":             run["experimentRunID"],
			"experimentId":   run["experimentID"],
			"experimentName": run["experimentName"],
			"status":         run["phase"],
			"resiliencyScore": run["resiliencyScore"],
			"faultsSummary": map[string]interface{}{
				"passed":        run["faultsPassed"],
				"failed":        run["faultsFailed"],
				"awaited":       run["faultsAwaited"],
				"stopped":       run["faultsStopped"],
				"notApplicable": run["faultsNa"],
				"total":         run["totalFaults"],
			},
			"infrastructure": infrastructure,
			"executionData":  executionData,
			"sequence":       run["runSequence"],
			"createdBy":      getNestedString(run, "createdBy", "username"),
			"updatedBy":      getNestedString(run, "updatedBy", "username"),
			"createdAt":      run["createdAt"],
			"updatedAt":      run["updatedAt"],
		},
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) listChaosInfrastructures(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query := `
		query ListInfras($projectID: ID!, $request: ListInfraRequest) {
			listInfras(projectID: $projectID, request: $request) {
				totalNoOfInfras
				infras {
					projectID
					infraID
					name
					description
					environmentID
					platformName
					isActive
					isInfraConfirmed
					infraScope
					infraNamespace
					version
					noOfExperiments
					noOfExperimentRuns
					tags
					createdAt
					updatedAt
					createdBy {
						username
					}
					updateStatus
				}
			}
		}
	`

	request := map[string]interface{}{}

	if environmentID := getStringFromArgs(args, "environmentId", ""); environmentID != "" {
		request["environmentIDs"] = []string{environmentID}
	}

	if status := getStringFromArgs(args, "status", ""); status != "" {
		request["filter"] = map[string]interface{}{
			"isActive": status == "Active",
		}
	}

	variables := map[string]interface{}{}
	if len(request) > 0 {
		variables["request"] = request
	}

	data, err := s.graphqlRequest(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	listInfras := result["listInfras"].(map[string]interface{})
	infras := listInfras["infras"].([]interface{})

	formattedInfras := make([]map[string]interface{}, len(infras))
	for i, infra := range infras {
		infraMap := infra.(map[string]interface{})

		formattedInfras[i] = map[string]interface{}{
			"id":          infraMap["infraID"],
			"name":        infraMap["name"],
			"description": infraMap["description"],
			"environment": infraMap["environmentID"],
			"platform":    infraMap["platformName"],
			"active":      infraMap["isActive"],
			"confirmed":   infraMap["isInfraConfirmed"],
			"scope":       infraMap["infraScope"],
			"namespace":   infraMap["infraNamespace"],
			"version":     infraMap["version"],
			"statistics": map[string]interface{}{
				"experiments": infraMap["noOfExperiments"],
				"runs":        infraMap["noOfExperimentRuns"],
			},
			"tags":         infraMap["tags"],
			"updateStatus": infraMap["updateStatus"],
			"createdBy":    getNestedString(infraMap, "createdBy", "username"),
			"createdAt":    infraMap["createdAt"],
			"updatedAt":    infraMap["updatedAt"],
		}
	}

	response := map[string]interface{}{
		"summary":              fmt.Sprintf("Found %v chaos infrastructures", listInfras["totalNoOfInfras"]),
		"totalInfrastructures": listInfras["totalNoOfInfras"],
		"infrastructures":      formattedInfras,
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) getInfrastructureDetails(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	infraID := getStringFromArgs(args, "infraId", "")
	if infraID == "" {
		return nil, fmt.Errorf("infraId is required")
	}

	query := `
		query GetInfra($projectID: ID!, $infraID: String!) {
			getInfra(projectID: $projectID, infraID: $infraID) {
				projectID
				infraID
				name
				description
				environmentID
				platformName
				isActive
				isInfraConfirmed
				infraScope
				infraNamespace
				serviceAccount
				infraNsExists
				infraSaExists
				version
				token
				noOfExperiments
				noOfExperimentRuns
				lastExperimentTimestamp
				startTime
				tags
				createdAt
				updatedAt
				createdBy {
					username
					email
				}
				updatedBy {
					username
					email
				}
				updateStatus
			}
		}
	`

	variables := map[string]interface{}{
		"infraID": infraID,
	}

	data, err := s.graphqlRequest(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	infra := result["getInfra"].(map[string]interface{})

	var manifest interface{} = nil
	if getBoolFromArgs(args, "includeManifest", false) {
		manifestQuery := `
			query GetInfraManifest(
				$infraID: ID!,
				$upgrade: Boolean!,
				$projectID: ID!
			) {
				getInfraManifest(
					infraID: $infraID,
					upgrade: $upgrade,
					projectID: $projectID
				)
			}
		`

		manifestVars := map[string]interface{}{
			"infraID": infraID,
			"upgrade": false,
		}

		manifestData, manifestErr := s.graphqlRequest(ctx, manifestQuery, manifestVars)
		if manifestErr == nil {
			var manifestResult map[string]interface{}
			if json.Unmarshal(manifestData, &manifestResult) == nil {
				manifest = manifestResult["getInfraManifest"]
			}
		}
		if manifest == nil {
			manifest = "Manifest not available"
		}
	}

	response := map[string]interface{}{
		"infrastructure": map[string]interface{}{
			"id":          infra["infraID"],
			"name":        infra["name"],
			"description": infra["description"],
			"environment": infra["environmentID"],
			"platform":    infra["platformName"],
			"active":      infra["isActive"],
			"confirmed":   infra["isInfraConfirmed"],
			"scope":       infra["infraScope"],
			"namespace":   infra["infraNamespace"],
			"serviceAccount": infra["serviceAccount"],
			"namespaceExists": infra["infraNsExists"],
			"serviceAccountExists": infra["infraSaExists"],
			"version":     infra["version"],
			"statistics": map[string]interface{}{
				"experiments":     infra["noOfExperiments"],
				"runs":           infra["noOfExperimentRuns"],
				"lastExperiment": infra["lastExperimentTimestamp"],
			},
			"startTime":     infra["startTime"],
			"tags":          infra["tags"],
			"updateStatus":  infra["updateStatus"],
			"createdBy":     getNestedString(infra, "createdBy", "username"),
			"updatedBy":     getNestedString(infra, "updatedBy", "username"),
			"createdAt":     infra["createdAt"],
			"updatedAt":     infra["updatedAt"],
			"manifest":      manifest,
		},
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) listEnvironments(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query := `
		query ListEnvironments($projectID: ID!, $request: ListEnvironmentRequest) {
			listEnvironments(projectID: $projectID, request: $request) {
				totalNoOfEnvironments
				environments {
					projectID
					environmentID
					name
					description
					type
					tags
					infraIDs
					createdAt
					updatedAt
					createdBy {
						username
					}
					updatedBy {
						username
					}
				}
			}
		}
	`

	request := map[string]interface{}{}

	if envType := getStringFromArgs(args, "type", ""); envType != "" {
		request["filter"] = map[string]interface{}{
			"type": envType,
		}
	}

	variables := map[string]interface{}{}
	if len(request) > 0 {
		variables["request"] = request
	}

	data, err := s.graphqlRequest(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	listEnvs := result["listEnvironments"].(map[string]interface{})
	environments := listEnvs["environments"].([]interface{})

	formattedEnvs := make([]map[string]interface{}, len(environments))
	for i, env := range environments {
		envMap := env.(map[string]interface{})

		infraCount := 0
		var infraIDs []interface{}
		if infraIDsRaw := envMap["infraIDs"]; infraIDsRaw != nil {
			infraIDs = infraIDsRaw.([]interface{})
			infraCount = len(infraIDs)
		}

		formattedEnvs[i] = map[string]interface{}{
			"id":                    envMap["environmentID"],
			"name":                  envMap["name"],
			"description":           envMap["description"],
			"type":                  envMap["type"],
			"tags":                  envMap["tags"],
			"infrastructureCount":   infraCount,
			"infrastructureIds":     infraIDs,
			"createdBy":             getNestedString(envMap, "createdBy", "username"),
			"updatedBy":             getNestedString(envMap, "updatedBy", "username"),
			"createdAt":             envMap["createdAt"],
			"updatedAt":             envMap["updatedAt"],
		}
	}

	response := map[string]interface{}{
		"summary":           fmt.Sprintf("Found %v environments", listEnvs["totalNoOfEnvironments"]),
		"totalEnvironments": listEnvs["totalNoOfEnvironments"],
		"environments":      formattedEnvs,
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) createEnvironment(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name := getStringFromArgs(args, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	envType := getStringFromArgs(args, "type", "")
	if envType == "" {
		return nil, fmt.Errorf("type is required")
	}

	mutation := `
		mutation CreateEnvironment($projectID: ID!, $request: CreateEnvironmentRequest) {
			createEnvironment(projectID: $projectID, request: $request) {
				projectID
				environmentID
				name
				description
				type
				tags
				createdAt
				createdBy {
					username
				}
			}
		}
	`

	tags := []string{}
	if tagsSlice := getSliceFromArgs(args, "tags"); tagsSlice != nil {
		tags = make([]string, len(tagsSlice))
		for i, tag := range tagsSlice {
			tags[i] = fmt.Sprintf("%v", tag)
		}
	}

	request := map[string]interface{}{
		"environmentID": strings.ToLower(strings.ReplaceAll(name, " ", "-")),
		"name":          name,
		"description":   getStringFromArgs(args, "description", ""),
		"type":          envType,
		"tags":          tags,
	}

	variables := map[string]interface{}{
		"request": request,
	}

	data, err := s.graphqlRequest(ctx, mutation, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	createResult := result["createEnvironment"].(map[string]interface{})

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Environment '%s' created successfully", name),
		"environment": map[string]interface{}{
			"id":          createResult["environmentID"],
			"name":        createResult["name"],
			"description": createResult["description"],
			"type":        createResult["type"],
			"tags":        createResult["tags"],
			"createdBy":   getNestedString(createResult, "createdBy", "username"),
			"createdAt":   createResult["createdAt"],
		},
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) listResilienceProbes(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query := `
		query ListProbes(
			$projectID: ID!,
			$infrastructureType: InfrastructureType,
			$probeNames: [ID!],
			$filter: ProbeFilterInput
		) {
			listProbes(
				projectID: $projectID,
				infrastructureType: $infrastructureType,
				probeNames: $probeNames,
				filter: $filter
			) {
				projectID
				name
				description
				type
				infrastructureType
				tags
				referencedBy
				updatedAt
				createdAt
				createdBy {
					username
				}
				updatedBy {
					username
				}
			}
		}
	`

	variables := map[string]interface{}{}

	if probeType := getStringFromArgs(args, "type", ""); probeType != "" {
		variables["filter"] = map[string]interface{}{
			"type": []string{probeType},
		}
	}

	data, err := s.graphqlRequest(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	probes := result["listProbes"].([]interface{})

	formattedProbes := make([]map[string]interface{}, len(probes))
	for i, probe := range probes {
		probeMap := probe.(map[string]interface{})

		formattedProbes[i] = map[string]interface{}{
			"name":               probeMap["name"],
			"description":        probeMap["description"],
			"type":               probeMap["type"],
			"infrastructureType": probeMap["infrastructureType"],
			"tags":               probeMap["tags"],
			"referencedBy":       probeMap["referencedBy"],
			"createdBy":          getNestedString(probeMap, "createdBy", "username"),
			"updatedBy":          getNestedString(probeMap, "updatedBy", "username"),
			"createdAt":          probeMap["createdAt"],
			"updatedAt":          probeMap["updatedAt"],
		}
	}

	response := map[string]interface{}{
		"summary":     fmt.Sprintf("Found %d resilience probes", len(probes)),
		"totalProbes": len(probes),
		"probes":      formattedProbes,
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) createResilienceProbe(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name := getStringFromArgs(args, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	probeType := getStringFromArgs(args, "type", "")
	if probeType == "" {
		return nil, fmt.Errorf("type is required")
	}

	properties := getMapFromArgs(args, "properties")
	if properties == nil {
		return nil, fmt.Errorf("properties are required")
	}

	mutation := `
		mutation AddProbe($request: ProbeRequest!, $projectID: ID!) {
			addProbe(request: $request, projectID: $projectID) {
				projectID
				name
				description
				type
				infrastructureType
				tags
				createdAt
				createdBy {
					username
				}
			}
		}
	`

	tags := []string{}
	if tagsSlice := getSliceFromArgs(args, "tags"); tagsSlice != nil {
		tags = make([]string, len(tagsSlice))
		for i, tag := range tagsSlice {
			tags[i] = fmt.Sprintf("%v", tag)
		}
	}

	request := map[string]interface{}{
		"name":               name,
		"description":        getStringFromArgs(args, "description", ""),
		"type":               probeType,
		"infrastructureType": "Kubernetes",
		"tags":               tags,
	}

	// Add type-specific properties
	switch probeType {
	case "httpProbe":
		method := strings.ToLower(getStringFromArgs(properties, "method", "get"))
		methodConfig := map[string]interface{}{
			"criteria":     "==",
			"responseCode": "200",
		}

		request["kubernetesHTTPProperties"] = map[string]interface{}{
			"url": getStringFromArgs(properties, "url", ""),
			"method": map[string]interface{}{
				method: methodConfig,
			},
			"probeTimeout":        getStringFromArgs(properties, "timeout", "5s"),
			"interval":            getStringFromArgs(properties, "interval", "2s"),
			"retry":               3,
			"attempt":             1,
			"insecureSkipVerify":  false,
		}

	case "cmdProbe":
		request["kubernetesCMDProperties"] = map[string]interface{}{
			"command":      getStringFromArgs(properties, "command", ""),
			"probeTimeout": getStringFromArgs(properties, "timeout", "5s"),
			"interval":     getStringFromArgs(properties, "interval", "2s"),
			"retry":        3,
			"attempt":      1,
			"comparator": map[string]interface{}{
				"type":     "string",
				"criteria": "==",
				"value":    "success",
			},
		}

	case "k8sProbe":
		request["k8sProperties"] = map[string]interface{}{
			"group":        "",
			"version":      "v1",
			"resource":     getStringFromArgs(properties, "resource", ""),
			"operation":    "present",
			"probeTimeout": getStringFromArgs(properties, "timeout", "5s"),
			"interval":     getStringFromArgs(properties, "interval", "2s"),
			"retry":        3,
			"attempt":      1,
		}

	case "promProbe":
		request["promProperties"] = map[string]interface{}{
			"endpoint":     getStringFromArgs(properties, "endpoint", ""),
			"query":        getStringFromArgs(properties, "query", ""),
			"probeTimeout": getStringFromArgs(properties, "timeout", "5s"),
			"interval":     getStringFromArgs(properties, "interval", "2s"),
			"retry":        3,
			"attempt":      1,
			"comparator": map[string]interface{}{
				"type":     "float",
				"criteria": ">=",
				"value":    "0",
			},
		}
	}

	variables := map[string]interface{}{
		"request": request,
	}

	data, err := s.graphqlRequest(ctx, mutation, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	createResult := result["addProbe"].(map[string]interface{})

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Resilience probe '%s' created successfully", name),
		"probe": map[string]interface{}{
			"name":               createResult["name"],
			"description":        createResult["description"],
			"type":               createResult["type"],
			"infrastructureType": createResult["infrastructureType"],
			"tags":               createResult["tags"],
			"createdBy":          getNestedString(createResult, "createdBy", "username"),
			"createdAt":          createResult["createdAt"],
		},
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) listChaosHubs(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query := `
		query ListChaosHub($projectID: ID!, $request: ListChaosHubRequest) {
			listChaosHub(projectID: $projectID, request: $request) {
				id
				name
				description
				repoURL
				repoBranch
				remoteHub
				hubType
				isPrivate
				isAvailable
				totalFaults
				totalExperiments
				tags
				lastSyncedAt
				createdAt
				updatedAt
				createdBy {
					username
				}
				updatedBy {
					username
				}
			}
		}
	`

	request := map[string]interface{}{}

	if hubType := getStringFromArgs(args, "hubType", ""); hubType != "" {
		request["filter"] = map[string]interface{}{
			"hubType": hubType,
		}
	}

	variables := map[string]interface{}{}
	if len(request) > 0 {
		variables["request"] = request
	}

	data, err := s.graphqlRequest(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	hubs := result["listChaosHub"].([]interface{})

	formattedHubs := make([]map[string]interface{}, len(hubs))
	for i, hub := range hubs {
		hubMap := hub.(map[string]interface{})

		formattedHubs[i] = map[string]interface{}{
			"id":          hubMap["id"],
			"name":        hubMap["name"],
			"description": hubMap["description"],
			"repoUrl":     hubMap["repoURL"],
			"branch":      hubMap["repoBranch"],
			"remoteHub":   hubMap["remoteHub"],
			"type":        hubMap["hubType"],
			"private":     hubMap["isPrivate"],
			"available":   hubMap["isAvailable"],
			"statistics": map[string]interface{}{
				"totalFaults":      hubMap["totalFaults"],
				"totalExperiments": hubMap["totalExperiments"],
			},
			"tags":        hubMap["tags"],
			"lastSynced":  hubMap["lastSyncedAt"],
			"createdBy":   getNestedString(hubMap, "createdBy", "username"),
			"updatedBy":   getNestedString(hubMap, "updatedBy", "username"),
			"createdAt":   hubMap["createdAt"],
			"updatedAt":   hubMap["updatedAt"],
		}
	}

	response := map[string]interface{}{
		"summary":   fmt.Sprintf("Found %d chaos hubs", len(hubs)),
		"totalHubs": len(hubs),
		"hubs":      formattedHubs,
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) getChaosFaults(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	hubID := getStringFromArgs(args, "hubId", "")
	if hubID == "" {
		return nil, fmt.Errorf("hubId is required")
	}

	query := `
		query ListChaosFaults($hubID: ID!, $projectID: ID!) {
			listChaosFaults(hubID: $hubID, projectID: $projectID) {
				apiVersion
				kind
				metadata {
					name
					version
					annotations {
						categories
						vendor
						repository
					}
				}
				spec {
					displayName
					categoryDescription
					keywords
					maturity
					platforms
					chaosType
					faults {
						name
						displayName
						description
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"hubID": hubID,
	}

	data, err := s.graphqlRequest(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	faultCategories := result["listChaosFaults"].([]interface{})

	category := getStringFromArgs(args, "category", "")
	filteredCategories := make([]map[string]interface{}, 0)

	for _, cat := range faultCategories {
		categoryMap := cat.(map[string]interface{})
		metadata := categoryMap["metadata"].(map[string]interface{})
		spec := categoryMap["spec"].(map[string]interface{})

		categoryName := metadata["name"].(string)
		if category == "" || strings.Contains(strings.ToLower(categoryName), strings.ToLower(category)) {
			var annotations map[string]interface{}
			if ann := metadata["annotations"]; ann != nil {
				annotations = ann.(map[string]interface{})
			}

			var faults []map[string]interface{}
			if faultsList := spec["faults"]; faultsList != nil {
				faultsSlice := faultsList.([]interface{})
				faults = make([]map[string]interface{}, len(faultsSlice))
				for i, fault := range faultsSlice {
					faultMap := fault.(map[string]interface{})
					faults[i] = map[string]interface{}{
						"name":        faultMap["name"],
						"displayName": faultMap["displayName"],
						"description": faultMap["description"],
					}
				}
			}

			formattedCategory := map[string]interface{}{
				"name":        categoryName,
				"displayName": spec["displayName"],
				"description": spec["categoryDescription"],
				"version":     metadata["version"],
				"keywords":    spec["keywords"],
				"maturity":    spec["maturity"],
				"platforms":   spec["platforms"],
				"chaosType":   spec["chaosType"],
				"faults":      faults,
			}

			if annotations != nil {
				formattedCategory["vendor"] = annotations["vendor"]
				formattedCategory["repository"] = annotations["repository"]
			}

			filteredCategories = append(filteredCategories, formattedCategory)
		}
	}

	response := map[string]interface{}{
		"hubId":                hubID,
		"totalFaultCategories": len(filteredCategories),
		"faultCategories":      filteredCategories,
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) getExperimentStatistics(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	statsQuery := `
		query GetExperimentStats($projectID: ID!) {
			getExperimentStats(projectID: $projectID) {
				totalExperiments
				totalExpCategorizedByResiliencyScore {
					id
					count
				}
			}
		}
	`

	runStatsQuery := `
		query GetExperimentRunStats($projectID: ID!) {
			getExperimentRunStats(projectID: $projectID) {
				totalExperimentRuns
				totalCompletedExperimentRuns
				totalTerminatedExperimentRuns
				totalRunningExperimentRuns
				totalStoppedExperimentRuns
				totalErroredExperimentRuns
			}
		}
	`

	infraStatsQuery := `
		query GetInfraStats($projectID: ID!) {
			getInfraStats(projectID: $projectID) {
				totalInfrastructures
				totalActiveInfrastructure
				totalInactiveInfrastructures
				totalConfirmedInfrastructure
				totalNonConfirmedInfrastructures
			}
		}
	`

	// Execute all queries concurrently
	expStatsCh := make(chan json.RawMessage, 1)
	runStatsCh := make(chan json.RawMessage, 1)
	infraStatsCh := make(chan json.RawMessage, 1)
	errCh := make(chan error, 3)

	go func() {
		data, err := s.graphqlRequest(ctx, statsQuery, nil)
		if err != nil {
			errCh <- err
			return
		}
		expStatsCh <- data
	}()

	go func() {
		data, err := s.graphqlRequest(ctx, runStatsQuery, nil)
		if err != nil {
			errCh <- err
			return
		}
		runStatsCh <- data
	}()

	go func() {
		data, err := s.graphqlRequest(ctx, infraStatsQuery, nil)
		if err != nil {
			errCh <- err
			return
		}
		infraStatsCh <- data
	}()

	// Collect results
	var expStatsData, runStatsData, infraStatsData json.RawMessage
	for i := 0; i < 3; i++ {
		select {
		case expStatsData = <-expStatsCh:
		case runStatsData = <-runStatsCh:
		case infraStatsData = <-infraStatsCh:
		case err := <-errCh:
			return nil, err
		}
	}

	// Parse results
	var expStats, runStats, infraStats map[string]interface{}

	if err := json.Unmarshal(expStatsData, &expStats); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(runStatsData, &runStats); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(infraStatsData, &infraStats); err != nil {
		return nil, err
	}

	expStatsResult := expStats["getExperimentStats"].(map[string]interface{})
	runStatsResult := runStats["getExperimentRunStats"].(map[string]interface{})
	infraStatsResult := infraStats["getInfraStats"].(map[string]interface{})

	response := map[string]interface{}{
		"overview": map[string]interface{}{
			"totalExperiments":      expStatsResult["totalExperiments"],
			"totalExperimentRuns":   runStatsResult["totalExperimentRuns"],
			"totalInfrastructures":  infraStatsResult["totalInfrastructures"],
		},
		"experimentStatistics": map[string]interface{}{
			"total": expStatsResult["totalExperiments"],
			"resiliencyScoreDistribution": expStatsResult["totalExpCategorizedByResiliencyScore"],
		},
		"experimentRunStatistics": map[string]interface{}{
			"total":      runStatsResult["totalExperimentRuns"],
			"completed":  runStatsResult["totalCompletedExperimentRuns"],
			"terminated": runStatsResult["totalTerminatedExperimentRuns"],
			"running":    runStatsResult["totalRunningExperimentRuns"],
			"stopped":    runStatsResult["totalStoppedExperimentRuns"],
			"errored":    runStatsResult["totalErroredExperimentRuns"],
		},
		"infrastructureStatistics": map[string]interface{}{
			"total":       infraStatsResult["totalInfrastructures"],
			"active":      infraStatsResult["totalActiveInfrastructure"],
			"inactive":    infraStatsResult["totalInactiveInfrastructures"],
			"confirmed":   infraStatsResult["totalConfirmedInfrastructure"],
			"unconfirmed": infraStatsResult["totalNonConfirmedInfrastructures"],
		},
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}

func (s *LitmusChaosServer) registerChaosInfrastructure(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name := getStringFromArgs(args, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	environmentID := getStringFromArgs(args, "environmentId", s.config.DefaultEnvironmentID)
	if environmentID == "" {
		return nil, fmt.Errorf("environmentId is required")
	}

	infraScope := getStringFromArgs(args, "infraScope", "")
	if infraScope == "" {
		return nil, fmt.Errorf("infraScope is required")
	}

	mutation := `
		mutation RegisterInfra($projectID: ID!, $request: RegisterInfraRequest!) {
			registerInfra(projectID: $projectID, request: $request) {
				token
				infraID
				name
				manifest
			}
		}
	`

	tags := []string{}
	if tagsSlice := getSliceFromArgs(args, "tags"); tagsSlice != nil {
		tags = make([]string, len(tagsSlice))
		for i, tag := range tagsSlice {
			tags[i] = fmt.Sprintf("%v", tag)
		}
	}

	request := map[string]interface{}{
		"name":                name,
		"description":         getStringFromArgs(args, "description", "Registered via MCP Server"),
		"environmentID":       environmentID,
		"infrastructureType":  "Kubernetes",
		"platformName":        getStringFromArgs(args, "platformName", "Generic Kubernetes"),
		"infraScope":          infraScope,
		"infraNamespace":      getStringFromArgs(args, "infraNamespace", "litmus"),
		"serviceAccount":      "litmus-admin",
		"infraNsExists":       false,
		"infraSaExists":       false,
		"skipSsl":             false,
		"tags":                tags,
	}

	variables := map[string]interface{}{
		"request": request,
	}

	data, err := s.graphqlRequest(ctx, mutation, variables)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	registerResult := result["registerInfra"].(map[string]interface{})

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Chaos infrastructure '%s' registered successfully", name),
		"infrastructure": map[string]interface{}{
			"id":    registerResult["infraID"],
			"name":  registerResult["name"],
			"token": registerResult["token"],
			"installationInstructions": map[string]interface{}{
				"step1":    "Apply the following manifest to your Kubernetes cluster:",
				"manifest": registerResult["manifest"],
				"step2":    "Wait for the infrastructure to be confirmed in the Chaos Center",
				"step3":    "Start creating and running chaos experiments",
			},
		},
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return &ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(responseJSON),
			},
		},
	}, nil
}
