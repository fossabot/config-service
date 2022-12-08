package main

import (
	"encoding/json"
	"fmt"
	"kubescape-config-service/types"

	"net/http"
	"net/http/httptest"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
)

////////////////////////////////////////// Test scenarios //////////////////////////////////////////

// common test for (almost) all documents
func commonTest[T types.DocContent](suite *MainTestSuite, path string, testDocs []T, modifyFunc func(T) T, compareNewOpts ...cmp.Option) {
	if len(testDocs) < 3 {
		suite.FailNow("commonTest: need at least 3 documents")
	}
	doc1 := testDocs[0]
	documents := testDocs[1:]
	//POST
	//create doc
	doc1 = testPostDoc(suite, path, doc1, compareNewOpts...)
	//post doc with same name should fail
	sameNameDoc := clone(doc1)
	testBadRequest(suite, http.MethodPost, path, errorNameExist(sameNameDoc.GetName()), sameNameDoc, http.StatusBadRequest)
	//post doc with no name should fail
	noNameDoc := clone(doc1)
	noNameDoc.SetName("")
	testBadRequest(suite, http.MethodPost, path, errorMissingName, &noNameDoc, http.StatusBadRequest)
	//bulk post documents
	documents = testBulkPostDocs(suite, path, documents, compareNewOpts...)
	//bulk post documents with same name should fail
	names := []string{documents[0].GetName(), documents[1].GetName()}
	sort.Strings(names)
	testBadRequest(suite, http.MethodPost, path, errorNameExist(names...), documents, http.StatusBadRequest)

	//PUT
	oldDoc1 := clone(doc1)
	doc1 = modifyFunc(doc1)
	testPutDoc(suite, path, oldDoc1, doc1)
	//test put with guid in path
	oldDoc1 = clone(doc1)
	doc1 = modifyFunc(doc1)
	testPutDocWGuid(suite, path, oldDoc1, doc1)
	//test put with no guid should fail
	noGuidDoc := clone(doc1)
	noGuidDoc.SetGUID("")
	testBadRequest(suite, http.MethodPut, path, errorMissingGUID, &noGuidDoc, http.StatusBadRequest)
	//not existing doc should fail
	noneExistingDoc := clone(doc1)
	noneExistingDoc.SetGUID("no_exist")
	testBadRequest(suite, http.MethodPut, path, errorDocumentNotFound, &noneExistingDoc, http.StatusNotFound)

	//GET
	//test get by guid
	pathWGuid := fmt.Sprintf("%s/%s", path, doc1.GetGUID())
	testGetDoc(suite, pathWGuid, doc1)
	//test get all
	docs := []T{doc1}
	docs = append(docs, documents...)
	testGetDocs(suite, path, docs, compareNewOpts...)
	//test get with wrong guid should fail
	testBadRequest(suite, http.MethodGet, fmt.Sprintf("%s/%s", path, "no_exist"), errorDocumentNotFound, nil, http.StatusNotFound)

	//test delete by guid
	testDeleteDocByGUID(suite, path, doc1)
	//test get all after delete
	testGetDocs(suite, path, documents, compareNewOpts...)
	//delete the rest of the docs
	for _, doc := range documents {
		testDeleteDocByGUID(suite, path, doc)
	}
	//test get all after delete all
	testGetDocs(suite, path, []T{}, compareNewOpts...)

	//test delete doc with wrong guid should fail
	testBadRequest(suite, http.MethodDelete, fmt.Sprintf("%s/%s", path, "no_exist"), errorDocumentNotFound, nil, http.StatusNotFound)
}

type queryTest[T types.DocContent] struct {
	query           string
	expectedIndexes []int
}

func testGetDeleteByNameAndQuery[T types.DocContent](suite *MainTestSuite, basePath, nameParam string, testDocs []T, getQueries []queryTest[T]) {
	newDocs := testBulkPostDocs(suite, basePath, testDocs, newPolicyCompareFilter)
	suite.Equal(len(testDocs), len(newDocs))
	docNames := []string{}
	for i := range newDocs {
		docNames = append(docNames, testDocs[i].GetName())
	}

	//test get name list
	testGetNameList(suite, basePath, docNames)
	//test get by name
	path := fmt.Sprintf("%s?%s=%s", basePath, nameParam, newDocs[0].GetName())
	testGetDoc(suite, path, newDocs[0])
	//test get by not existing name
	path = fmt.Sprintf("%s?%s=%s", basePath, nameParam, "notExistingName")
	testBadRequest(suite, http.MethodGet, path, errorDocumentNotFound, nil, http.StatusNotFound)

	//get Docs by query params
	for _, query := range getQueries {
		path = fmt.Sprintf("%s?%s", basePath, query.query)
		var expectedDocs []T
		for _, index := range query.expectedIndexes {
			expectedDocs = append(expectedDocs, newDocs[index])
		}
		testGetDocs(suite, path, expectedDocs)
	}

	//test delete by name
	testDeleteDocByName(suite, basePath, nameParam, newDocs[0])
	//test bulk delete by name
	docNames = docNames[1:]
	testBulkDeleteByName(suite, basePath, nameParam, docNames)
	//test delete by name with not existing name
	path = fmt.Sprintf("%s?%s=%s", basePath, nameParam, "notExistingName")
	testBadRequest(suite, http.MethodDelete, path, errorDocumentNotFound, nil, http.StatusNotFound)
	//deleteDoc by name with empty name
	path = fmt.Sprintf("%s?%s", basePath, nameParam)
	testBadRequest(suite, http.MethodDelete, path, errorMissingName, nil, http.StatusBadRequest)

}

////////////////////////////////////////// Test helpers //////////////////////////////////////////

const (
	//error messages
	errorMissingName      = `{"error":"name is required"}`
	errorMissingGUID      = `{"error":"document guid is required"}`
	errorDocumentNotFound = `{"error":"document not found"}`
)

func errorNameExist(name ...string) string {
	var msg string
	if len(name) == 0 {
		msg = "name already exists"
	} else if len(name) == 1 {
		msg = fmt.Sprintf("name %s already exists", name[0])
	} else {
		msg = fmt.Sprintf("names %s already exist", strings.Join(name, ","))
	}
	return `{"error":"` + msg + `"}`
}

func testBadRequest(suite *MainTestSuite, method, path, expectedResponse string, body interface{}, expectedCode int) {
	w := suite.doRequest(method, path, body)
	suite.Equal(expectedCode, w.Code)
	suite.Equal(expectedResponse, w.Body.String())
}

// //////////////////////////////////////// GET //////////////////////////////////////////
func testGetDoc[T types.DocContent](suite *MainTestSuite, path string, expectedDoc T, compareOpts ...cmp.Option) T {
	w := suite.doRequest(http.MethodGet, path, nil)
	suite.Equal(http.StatusOK, w.Code)
	doc, err := decode[T](w)
	if err != nil {
		suite.FailNow(err.Error())
	}
	diff := cmp.Diff(doc, expectedDoc, compareOpts...)
	suite.Equal("", diff)
	return doc
}

func testGetDocs[T types.DocContent](suite *MainTestSuite, path string, expectedDocs []T, compareOpts ...cmp.Option) (actualDocs []T) {
	w := suite.doRequest(http.MethodGet, path, nil)
	suite.Equal(http.StatusOK, w.Code)
	docs, err := decodeArray[T](w)
	if err != nil {
		suite.FailNow(err.Error())
	}
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].GetName() < docs[j].GetName()
	})
	sort.Slice(expectedDocs, func(i, j int) bool {
		return expectedDocs[i].GetName() < expectedDocs[j].GetName()
	})
	diff := cmp.Diff(docs, expectedDocs, compareOpts...)
	suite.Equal("", diff)
	return docs
}

func testGetNameList(suite *MainTestSuite, path string, expectedNames []string) {
	path = fmt.Sprintf("%s?list", path)
	w := suite.doRequest(http.MethodGet, path, nil)
	suite.Equal(http.StatusOK, w.Code)
	var names []string
	if err := json.Unmarshal(w.Body.Bytes(), &names); err != nil {
		suite.FailNow(err.Error())
	}
	sort.Strings(expectedNames)
	sort.Strings(names)
	diff := cmp.Diff(names, expectedNames)
	suite.Equal("", diff)
}

// //////////////////////////////////////// POST //////////////////////////////////////////
func testPostDoc[T types.DocContent](suite *MainTestSuite, path string, doc T, compareOpts ...cmp.Option) (newDoc T) {
	w := suite.doRequest(http.MethodPost, path, doc)
	suite.Equal(http.StatusCreated, w.Code)
	newDoc, err := decode[T](w)
	if err != nil {
		suite.FailNow(err.Error())
	}
	diff := cmp.Diff(doc, newDoc, compareOpts...)
	suite.Equal("", diff)
	return newDoc
}

func testBulkPostDocs[T types.DocContent](suite *MainTestSuite, path string, docs []T, compareOpts ...cmp.Option) (newDocs []T) {
	w := suite.doRequest(http.MethodPost, path, docs)
	suite.Equal(http.StatusOK, w.Code)
	newDocs, err := decodeArray[T](w)
	if err != nil {
		suite.FailNow(err.Error())
	}
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].GetName() < docs[j].GetName()
	})
	sort.Slice(newDocs, func(i, j int) bool {
		return newDocs[i].GetName() < newDocs[j].GetName()
	})
	diff := cmp.Diff(docs, newDocs, compareOpts...)
	suite.Equal("", diff)
	return newDocs
}

// //////////////////////////////////////// PUT //////////////////////////////////////////
func testPutDoc[T types.DocContent](suite *MainTestSuite, path string, oldDoc, newDoc T) {
	w := suite.doRequest(http.MethodPut, path, newDoc)
	suite.Equal(http.StatusOK, w.Code)
	response, err := decodeArray[T](w)
	expectedResponse := []T{oldDoc, newDoc}
	if err != nil {
		suite.FailNow(err.Error())
	}
	sort.Slice(response, func(i, j int) bool {
		return response[i].GetName() < response[j].GetName()
	})
	sort.Slice(expectedResponse, func(i, j int) bool {
		return expectedResponse[i].GetName() < expectedResponse[j].GetName()
	})
	diff := cmp.Diff(response, expectedResponse)
	suite.Equal("", diff)
}

func testPutDocWGuid[T types.DocContent](suite *MainTestSuite, path string, oldDoc, newDoc T) {
	guid := newDoc.GetGUID()
	path = fmt.Sprintf("%s/%s", path, guid)
	newDoc.SetGUID("")
	w := suite.doRequest(http.MethodPut, path, newDoc)
	suite.Equal(http.StatusOK, w.Code)
	response, err := decodeArray[T](w)
	if err != nil {
		suite.FailNow(err.Error())
	}
	newDoc.SetGUID(guid)
	expectedResponse := []T{oldDoc, newDoc}
	sort.Slice(response, func(i, j int) bool {
		return response[i].GetName() < response[j].GetName()
	})
	sort.Slice(expectedResponse, func(i, j int) bool {
		return expectedResponse[i].GetName() < expectedResponse[j].GetName()
	})
	diff := cmp.Diff(response, expectedResponse)
	suite.Equal("", diff)
}

// //////////////////////////////////////// DELETE //////////////////////////////////////////
func testDeleteDocByGUID[T types.DocContent](suite *MainTestSuite, path string, doc2Delete T, compareOpts ...cmp.Option) {
	path = fmt.Sprintf("%s/%s", path, doc2Delete.GetGUID())
	w := suite.doRequest(http.MethodDelete, path, nil)
	suite.Equal(http.StatusOK, w.Code)
	deleteDoc, err := decode[T](w)
	if err != nil {
		suite.FailNow(err.Error())
	}
	diff := cmp.Diff(deleteDoc, doc2Delete, compareOpts...)
	suite.Equal("", diff)
}

func testDeleteDocByName[T types.DocContent](suite *MainTestSuite, path string, nameParam string, doc2Delete T) {
	path = fmt.Sprintf("%s?%s=%s", path, nameParam, doc2Delete.GetName())
	w := suite.doRequest(http.MethodDelete, path, nil)
	suite.Equal(http.StatusOK, w.Code)
	deleteDoc, err := decode[T](w)
	if err != nil {
		suite.FailNow(err.Error())
	}
	diff := cmp.Diff(deleteDoc, doc2Delete)
	suite.Equal("", diff)
}

func testBulkDeleteByName(suite *MainTestSuite, path string, nameParam string, names []string) {
	if len(names) == 0 {
		return
	}
	path = fmt.Sprintf("%s?%s=%s", path, nameParam, names[0])
	for _, name := range names[1:] {
		path = fmt.Sprintf("%s&%s=%s", path, nameParam, name)
	}
	w := suite.doRequest(http.MethodDelete, path, nil)
	suite.Equal(http.StatusOK, w.Code)
	diff := cmp.Diff(fmt.Sprintf(`{"DeletedCount":%d}`, len(names)), w.Body.String())
	suite.Equal("", diff)
}

//helpers

func decode[T types.DocContent](w *httptest.ResponseRecorder) (T, error) {
	var content T
	err := json.Unmarshal(w.Body.Bytes(), &content)
	return content, err
}

func decodeArray[T types.DocContent](w *httptest.ResponseRecorder) ([]T, error) {
	var content []T
	err := json.Unmarshal(w.Body.Bytes(), &content)
	return content, err
}

func clone[T types.DocContent](orig T) T {
	origJSON, err := json.Marshal(orig)
	if err != nil {
		panic(err)
	}
	var clone T
	if err = json.Unmarshal(origJSON, &clone); err != nil {
		panic(err)
	}
	return clone
}
