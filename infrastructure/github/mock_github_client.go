// Code generated by MockGen. DO NOT EDIT.
// Source: infrastructure/github/client_interface.go
//
// Generated by this command:
//
//	mockgen -source=infrastructure/github/client_interface.go -destination=infrastructure/github/mock_github_client.go -package github
//

// Package github is a generated GoMock package.
package github

import (
	reflect "reflect"

	githubv4 "github.com/shurcooL/githubv4"
	gomock "go.uber.org/mock/gomock"
)

// MockGitHubClient is a mock of GitHubClient interface.
type MockGitHubClient struct {
	ctrl     *gomock.Controller
	recorder *MockGitHubClientMockRecorder
	isgomock struct{}
}

// MockGitHubClientMockRecorder is the mock recorder for MockGitHubClient.
type MockGitHubClientMockRecorder struct {
	mock *MockGitHubClient
}

// NewMockGitHubClient creates a new mock instance.
func NewMockGitHubClient(ctrl *gomock.Controller) *MockGitHubClient {
	mock := &MockGitHubClient{ctrl: ctrl}
	mock.recorder = &MockGitHubClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGitHubClient) EXPECT() *MockGitHubClientMockRecorder {
	return m.recorder
}

// GetIssueFields mocks base method.
func (m *MockGitHubClient) GetIssueFields(nodeID githubv4.ID) (*IssueFields, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIssueFields", nodeID)
	ret0, _ := ret[0].(*IssueFields)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetIssueFields indicates an expected call of GetIssueFields.
func (mr *MockGitHubClientMockRecorder) GetIssueFields(nodeID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIssueFields", reflect.TypeOf((*MockGitHubClient)(nil).GetIssueFields), nodeID)
}

// GetIssueNodeID mocks base method.
func (m *MockGitHubClient) GetIssueNodeID(q *QueryRequest) (githubv4.ID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIssueNodeID", q)
	ret0, _ := ret[0].(githubv4.ID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetIssueNodeID indicates an expected call of GetIssueNodeID.
func (mr *MockGitHubClientMockRecorder) GetIssueNodeID(q any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIssueNodeID", reflect.TypeOf((*MockGitHubClient)(nil).GetIssueNodeID), q)
}

// GetTrackedIssueNodeIDs mocks base method.
func (m *MockGitHubClient) GetTrackedIssueNodeIDs(q *QueryRequest) ([]githubv4.ID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTrackedIssueNodeIDs", q)
	ret0, _ := ret[0].([]githubv4.ID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTrackedIssueNodeIDs indicates an expected call of GetTrackedIssueNodeIDs.
func (mr *MockGitHubClientMockRecorder) GetTrackedIssueNodeIDs(q any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTrackedIssueNodeIDs", reflect.TypeOf((*MockGitHubClient)(nil).GetTrackedIssueNodeIDs), q)
}

// MutateIssue mocks base method.
func (m *MockGitHubClient) MutateIssue(input githubv4.UpdateIssueInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MutateIssue", input)
	ret0, _ := ret[0].(error)
	return ret0
}

// MutateIssue indicates an expected call of MutateIssue.
func (mr *MockGitHubClientMockRecorder) MutateIssue(input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MutateIssue", reflect.TypeOf((*MockGitHubClient)(nil).MutateIssue), input)
}

// MutateProject mocks base method.
func (m *MockGitHubClient) MutateProject(input githubv4.AddProjectV2ItemByIdInput) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MutateProject", input)
	ret0, _ := ret[0].(error)
	return ret0
}

// MutateProject indicates an expected call of MutateProject.
func (mr *MockGitHubClientMockRecorder) MutateProject(input any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MutateProject", reflect.TypeOf((*MockGitHubClient)(nil).MutateProject), input)
}
