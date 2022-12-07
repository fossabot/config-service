package main

import (
	"encoding/json"
	"fmt"
	"kubescape-config-service/types"
	"kubescape-config-service/utils/consts"
	"net/http"
	"sort"

	_ "embed"

	"github.com/google/go-cmp/cmp"
)

//go:embed test_data/clusters.json
var clustersJson []byte

func (suite *MainTestSuite) TestCluster() {
	var clusters []*types.Cluster
	if err := json.Unmarshal(clustersJson, &clusters); err != nil {
		panic(err)
	}

	modifyFunc := func(cluster *types.Cluster) *types.Cluster {
		if cluster.Attributes == nil {
			cluster.Attributes = make(map[string]interface{})
		}
		if _, ok := cluster.Attributes["test"]; ok {
			cluster.Attributes["test"] = cluster.Attributes["test"].(string) + "-modified"
		} else {
			cluster.Attributes["test"] = "test"
		}
		return cluster
	}

	newClusterCompareFilter := cmp.FilterPath(func(p cmp.Path) bool {
		switch p.String() {
		case "PortalBase.GUID", "SubscriptionDate", "LastLoginDate":
			return true
		case "PortalBase.Attributes":
			if p.Last().String() == `["alias"]` {
				return true
			}
		}
		return false
	}, cmp.Ignore())

	commonTest(suite, consts.ClusterPath, clusters, modifyFunc, newClusterCompareFilter)

	//cluster specific tests

	//put doc without alias - expect the alias not to be deleted
	cluster := testPostDoc(suite, consts.ClusterPath, clusters[0], newClusterCompareFilter)
	alias := cluster.Attributes["alias"].(string)
	suite.NotEmpty(alias)
	delete(cluster.Attributes, "alias")
	w := suite.doRequest(http.MethodPut, consts.ClusterPath, cluster)
	suite.Equal(http.StatusOK, w.Code)
	response, err := decodeArray[*types.Cluster](w)
	if err != nil || len(response) != 2 {
		panic(err)
	}
	suite.Equal(alias, response[1].Attributes["alias"].(string))

	//put doc without alias and wrong doc GUID
	cluster.GUID = "wrongGUID"
	delete(cluster.Attributes, "alias")
	testBadRequest(suite, http.MethodPut, consts.ClusterPath, errorDocumentNotFound, cluster, http.StatusNotFound)

	//put doc without attributes
	errorNoAttributes := `{"error":"cluster attributes are required"}`
	cluster.Attributes = nil
	testBadRequest(suite, http.MethodPut, consts.ClusterPath, errorNoAttributes, cluster, http.StatusBadRequest)
}

//go:embed test_data/posturePolicies.json
var posturePoliciesJson []byte

var newPolicyCompareFilter = cmp.FilterPath(func(p cmp.Path) bool {
	switch p.String() {
	case "PortalBase.GUID", "CreationTime":
		return true
	}
	return false
}, cmp.Ignore())

func (suite *MainTestSuite) TestPostureException() {
	var posturePolicies []*types.PostureExceptionPolicy
	if err := json.Unmarshal(posturePoliciesJson, &posturePolicies); err != nil {
		panic(err)
	}
	sort.Slice(posturePolicies, func(i, j int) bool {
		return posturePolicies[i].GetName() < posturePolicies[j].GetName()
	})

	modifyFunc := func(policy *types.PostureExceptionPolicy) *types.PostureExceptionPolicy {
		if policy.Attributes == nil {
			policy.Attributes = make(map[string]interface{})
		}
		if _, ok := policy.Attributes["test"]; ok {
			policy.Attributes["test"] = policy.Attributes["test"].(string) + "-modified"
		} else {
			policy.Attributes["test"] = "test"
		}
		return policy
	}

	commonTest(suite, consts.PostureExceptionPolicyPath, posturePolicies, modifyFunc, newPolicyCompareFilter)

	getQueries := []queryTest[*types.PostureExceptionPolicy]{
		{
			query:           "posturePolicies.controlName=Allowed hostPath&posturePolicies.controlName=Applications credentials in configuration files",
			expectedIndexes: []int{0, 1},
		},
		{
			query:           "resources.attributes.cluster=cluster1&scope.cluster=cluster3",
			expectedIndexes: []int{0, 2},
		},
		{
			query:           "scope.namespace=armo-system&scope.namespace=test-system&scope.cluster=cluster1&scope.cluster=cluster3",
			expectedIndexes: []int{0, 2},
		},
		{
			query:           "scope.namespace=armo-system&posturePolicies.frameworkName=MITRE",
			expectedIndexes: []int{1, 2},
		},
		{
			query:           "namespaceOnly=true",
			expectedIndexes: []int{1, 2},
		},
		{
			query:           "resources.attributes.cluster=cluster1",
			expectedIndexes: []int{2},
		},
		{
			query:           "posturePolicies.frameworkName=MITRE&posturePolicies.frameworkName=NSA",
			expectedIndexes: []int{0, 1, 2},
		},
		{
			query:           "posturePolicies.frameworkName=MITRE",
			expectedIndexes: []int{1, 2},
		},
		{
			query:           "posturePolicies.frameworkName=NSA",
			expectedIndexes: []int{0},
		},
	}
	testGetDeleteByNameAndQuery(suite, consts.PostureExceptionPolicyPath, consts.PolicyNameParam, posturePolicies, getQueries)
}

//go:embed test_data/vulnerabilityPolicies.json
var vulnerabilityPoliciesJson []byte

func (suite *MainTestSuite) TestVulnerabilityPolicies() {
	var vulnerabilities []*types.VulnerabilityExceptionPolicy
	if err := json.Unmarshal(vulnerabilityPoliciesJson, &vulnerabilities); err != nil {
		panic(err)
	}
	sort.Slice(vulnerabilities, func(i, j int) bool {
		return vulnerabilities[i].GetName() < vulnerabilities[j].GetName()
	})

	modifyFunc := func(policy *types.VulnerabilityExceptionPolicy) *types.VulnerabilityExceptionPolicy {
		if policy.Attributes == nil {
			policy.Attributes = make(map[string]interface{})
		}
		if _, ok := policy.Attributes["test"]; ok {
			policy.Attributes["test"] = policy.Attributes["test"].(string) + "-modified"
		} else {
			policy.Attributes["test"] = "test"
		}
		return policy
	}

	commonTest(suite, consts.VulnerabilityExceptionPolicyPath, vulnerabilities, modifyFunc, newPolicyCompareFilter)

	getQueries := []queryTest[*types.VulnerabilityExceptionPolicy]{
		{
			query:           "vulnerabilities.name=CVE-2005-2541&scope.cluster=dwertent",
			expectedIndexes: []int{2},
		},
		{
			query:           "scope.containerName=nginx&vulnerabilities.name=CVE-2009-5155",
			expectedIndexes: []int{0, 1},
		},
		{
			query:           "scope.containerName=nginx&vulnerabilities.name=CVE-2005-2541",
			expectedIndexes: []int{0, 2},
		},
		{
			query:           "scope.containerName=nginx&vulnerabilities.name=CVE-2005-2541&vulnerabilities.name=CVE-2005-2555",
			expectedIndexes: []int{0, 1, 2},
		},
		{
			query:           "scope.namespace=systest-ns-xpyz&designators.attributes.namespace=systest-ns-zao6",
			expectedIndexes: []int{1, 2},
		},
		{
			query:           "scope.namespace=systest-ns-xpyz&designators.attributes.namespace=systest-ns-9uqv&scope.containerName=nginx&vulnerabilities.name=CVE-2010-4756",
			expectedIndexes: []int{0},
		},
	}
	testGetDeleteByNameAndQuery(suite, consts.VulnerabilityExceptionPolicyPath, consts.PolicyNameParam, vulnerabilities, getQueries)
}

//go:embed test_data/customer_config/customerConfig.json
var customerConfigJson []byte

//go:embed test_data/customer_config/customerConfigMerged.json
var customerConfigMergedJson []byte

//go:embed test_data/customer_config/cluster1Config.json
var cluster1ConfigJson []byte

//go:embed test_data/customer_config/cluster1ConfigMerged.json
var cluster1ConfigMergedJson []byte

//go:embed test_data/customer_config/cluster1ConfigMergedWithDefault.json
var cluster1ConfigMergedWithDefaultJson []byte

//go:embed test_data/customer_config/cluster2Config.json
var cluster2ConfigJson []byte

//go:embed test_data/customer_config/cluster2ConfigMerged.json
var cluster2ConfigMergedJson []byte

func (suite *MainTestSuite) TestCustomerConfiguration() {
	//load test data
	var defaultCustomerConfig *types.CustomerConfig
	if err := json.Unmarshal(defaultCustomerConfigJson, &defaultCustomerConfig); err != nil {
		suite.FailNow("failed to unmarshal defaultCustomerConfigJson", err.Error())
	}
	var customerConfig *types.CustomerConfig
	if err := json.Unmarshal(customerConfigJson, &customerConfig); err != nil {
		suite.FailNow("failed to unmarshal defaultCustomerConfigJson", err.Error())
	}
	var customerConfigMerged *types.CustomerConfig
	if err := json.Unmarshal(customerConfigMergedJson, &customerConfigMerged); err != nil {
		suite.FailNow("failed to unmarshal defaultCustomerConfigJson", err.Error())
	}
	var cluster1Config *types.CustomerConfig
	if err := json.Unmarshal(cluster1ConfigJson, &cluster1Config); err != nil {
		suite.FailNow("failed to unmarshal clustersCustomerConfigJson", err.Error())
	}
	var cluster1MergedConfig *types.CustomerConfig
	if err := json.Unmarshal(cluster1ConfigMergedJson, &cluster1MergedConfig); err != nil {
		suite.FailNow("failed to unmarshal clustersCustomerConfigJson", err.Error())
	}
	var cluster1MergedWithDefaultConfig *types.CustomerConfig
	if err := json.Unmarshal(cluster1ConfigMergedWithDefaultJson, &cluster1MergedWithDefaultConfig); err != nil {
		suite.FailNow("failed to unmarshal clustersCustomerConfigJson", err.Error())
	}
	var cluster2Config *types.CustomerConfig
	if err := json.Unmarshal(cluster2ConfigJson, &cluster2Config); err != nil {
		suite.FailNow("failed to unmarshal clustersCustomerConfigJson", err.Error())
	}
	var cluster2MergedConfig *types.CustomerConfig
	if err := json.Unmarshal(cluster2ConfigMergedJson, &cluster2MergedConfig); err != nil {
		suite.FailNow("failed to unmarshal clustersCustomerConfigJson", err.Error())
	}
	//create compare options
	compareFilter := cmp.FilterPath(func(p cmp.Path) bool {
		switch p.String() {
		case "GUID", "CreationTime":
			return true
		}
		return false
	}, cmp.Ignore())

	//TESTS

	//get all customer configs - expect only the default one
	defaultCustomerConfig = testGetDocs(suite, consts.CustomerConfigPath, []*types.CustomerConfig{defaultCustomerConfig}, compareFilter)[0]
	//post new customer config
	customerConfig = testPostDoc(suite, consts.CustomerConfigPath, customerConfig, compareFilter)
	//post cluster configs
	clusterConfigs := testBulkPostDocs(suite, consts.CustomerConfigPath, []*types.CustomerConfig{cluster1Config, cluster2Config}, compareFilter)
	cluster1Config = clusterConfigs[0]
	cluster2Config = clusterConfigs[1]

	//test get names list
	configNames := []string{defaultCustomerConfig.Name, customerConfig.Name, cluster1Config.Name, cluster2Config.Name}
	testGetNameList(suite, consts.CustomerConfigPath, configNames)

	//test get default config
	//by name
	path := fmt.Sprintf("%s?%s=%s", consts.CustomerConfigPath, consts.ConfigNameParam, consts.GlobalConfigName)
	testGetDoc(suite, path, defaultCustomerConfig, compareFilter)
	//by scope
	path = fmt.Sprintf("%s?%s=%s", consts.CustomerConfigPath, consts.ScopeParam, consts.DefaultScope)
	testGetDoc(suite, path, defaultCustomerConfig, compareFilter)

	//test get merged customer config
	//by name
	path = fmt.Sprintf("%s?%s=%s", consts.CustomerConfigPath, consts.ConfigNameParam, consts.CustomerConfigName)
	testGetDoc(suite, path, customerConfigMerged, compareFilter)
	//by scope
	path = fmt.Sprintf("%s?%s=%s", consts.CustomerConfigPath, consts.ScopeParam, consts.CustomerScope)
	testGetDoc(suite, path, customerConfigMerged, compareFilter)
	//test get unmerged customer config
	//by name
	path = fmt.Sprintf("%s?%s=%s&unmerged=true", consts.CustomerConfigPath, consts.ConfigNameParam, consts.CustomerConfigName)
	testGetDoc(suite, path, customerConfig, compareFilter, compareFilter)
	//by scope
	path = fmt.Sprintf("%s?%s=%s&unmerged=true", consts.CustomerConfigPath, consts.ScopeParam, consts.CustomerScope)
	testGetDoc(suite, path, customerConfig, compareFilter)

	//test get merged cluster config by name
	//cluster1
	path = fmt.Sprintf("%s?%s=%s", consts.CustomerConfigPath, consts.ClusterNameParam, cluster1Config.GetName())
	testGetDoc(suite, path, cluster1MergedConfig, compareFilter)
	//cluster2
	path = fmt.Sprintf("%s?%s=%s", consts.CustomerConfigPath, consts.ClusterNameParam, cluster2Config.GetName())
	testGetDoc(suite, path, cluster2MergedConfig, compareFilter)
	//test get unmerged cluster config by name
	//cluster1
	path = fmt.Sprintf("%s?%s=%s&unmerged=true", consts.CustomerConfigPath, consts.ClusterNameParam, cluster1Config.GetName())
	testGetDoc(suite, path, cluster1Config, compareFilter)
	//cluster2
	path = fmt.Sprintf("%s?%s=%s&unmerged=true", consts.CustomerConfigPath, consts.ClusterNameParam, cluster2Config.GetName())
	testGetDoc(suite, path, cluster2Config, compareFilter)

	//delete customer config
	testDeleteDocByName(suite, consts.CustomerConfigPath, consts.ConfigNameParam, customerConfig)
	//get unmerged customer config - expect error 404
	path = fmt.Sprintf("%s?%s=%s&unmerged=true", consts.CustomerConfigPath, consts.ConfigNameParam, consts.CustomerConfigName)
	testBadRequest(suite, http.MethodGet, path, errorDocumentNotFound, nil, http.StatusNotFound)
	//get merged customer config - expect default config
	path = fmt.Sprintf("%s?%s=%s", consts.CustomerConfigPath, consts.ConfigNameParam, consts.CustomerConfigName)
	testGetDoc(suite, path, defaultCustomerConfig, compareFilter)
	//get merged cluster1 - expect merge with default config
	path = fmt.Sprintf("%s?%s=%s", consts.CustomerConfigPath, consts.ClusterNameParam, cluster1Config.GetName())
	testGetDoc(suite, path, cluster1MergedWithDefaultConfig, compareFilter)
	//delete cluster1 config
	testDeleteDocByName(suite, consts.CustomerConfigPath, consts.ClusterNameParam, cluster1Config)
	//get merged cluster1 - expect default config
	testGetDoc(suite, path, defaultCustomerConfig, compareFilter)
	//tets delete without name - expect error 400
	testBadRequest(suite, http.MethodDelete, consts.CustomerConfigPath, errorMissingName, nil, http.StatusBadRequest)

	//test put cluster2 config by cluster name
	oldCluster2 := clone(cluster2Config)
	cluster2Config.Settings.PostureScanConfig.ScanFrequency = "100h"
	path = fmt.Sprintf("%s?%s=%s", consts.CustomerConfigPath, consts.ClusterNameParam, cluster2Config.GetName())
	testPutDoc(suite, path, oldCluster2, cluster2Config)
	// put cluster2 config by config name
	oldCluster2 = clone(cluster2Config)
	cluster2Config.Settings.PostureControlInputs["allowedContainerRepos"] = []string{"repo1", "repo2"}
	path = fmt.Sprintf("%s?%s=%s", consts.CustomerConfigPath, consts.ConfigNameParam, cluster2Config.GetName())
	testPutDoc(suite, path, oldCluster2, cluster2Config)

	//put config with wrong name - expect error 400
	path = fmt.Sprintf("%s?%s=%s", consts.CustomerConfigPath, consts.ConfigNameParam, "notExist")
	testBadRequest(suite, http.MethodPut, path, errorDocumentNotFound, cluster2Config, http.StatusBadRequest)
	//test put with wrong config name param - expect error 400
	path = fmt.Sprintf("%s?%s=%s", consts.CustomerConfigPath, "wrongParamName", "someName")
	c2Name := cluster2Config.Name
	cluster2Config.Name = ""
	testBadRequest(suite, http.MethodPut, path, errorMissingName, cluster2Config, http.StatusBadRequest)
	//test put with no name in path but with name in config
	cluster2Config.Name = c2Name
	testPutDoc(suite, path, cluster2Config, cluster2Config)

	//post costumer config again
	customerConfig = testPostDoc(suite, consts.CustomerConfigPath, customerConfig, compareFilter)
	//update it by scope param
	oldCustomerConfig := clone(customerConfig)
	customerConfig.Settings.PostureScanConfig.ScanFrequency = "11h"
	path = fmt.Sprintf("%s?%s=%s", consts.CustomerConfigPath, consts.ScopeParam, consts.CustomerScope)
	testPutDoc(suite, path, oldCustomerConfig, customerConfig)

}
