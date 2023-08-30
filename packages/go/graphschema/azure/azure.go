// Copyright 2023 Specter Ops, Inc.
// 
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// 
// SPDX-License-Identifier: Apache-2.0

// Code generated by Cuelang code gen. DO NOT EDIT!
// Cuelang source: github.com/specterops/bloodhound/-/tree/main/packages/cue/schemas/

package azure

import (
	"errors"
	graph "github.com/specterops/bloodhound/dawgs/graph"
)

var (
	Entity                               = graph.StringKind("AZBase")
	VMScaleSet                           = graph.StringKind("AZVMScaleSet")
	App                                  = graph.StringKind("AZApp")
	Role                                 = graph.StringKind("AZRole")
	Device                               = graph.StringKind("AZDevice")
	FunctionApp                          = graph.StringKind("AZFunctionApp")
	Group                                = graph.StringKind("AZGroup")
	KeyVault                             = graph.StringKind("AZKeyVault")
	ManagementGroup                      = graph.StringKind("AZManagementGroup")
	ResourceGroup                        = graph.StringKind("AZResourceGroup")
	ServicePrincipal                     = graph.StringKind("AZServicePrincipal")
	Subscription                         = graph.StringKind("AZSubscription")
	Tenant                               = graph.StringKind("AZTenant")
	User                                 = graph.StringKind("AZUser")
	VM                                   = graph.StringKind("AZVM")
	ManagedCluster                       = graph.StringKind("AZManagedCluster")
	ContainerRegistry                    = graph.StringKind("AZContainerRegistry")
	WebApp                               = graph.StringKind("AZWebApp")
	LogicApp                             = graph.StringKind("AZLogicApp")
	AutomationAccount                    = graph.StringKind("AZAutomationAccount")
	AvereContributor                     = graph.StringKind("AZAvereContributor")
	Contains                             = graph.StringKind("AZContains")
	Contributor                          = graph.StringKind("AZContributor")
	GetCertificates                      = graph.StringKind("AZGetCertificates")
	GetKeys                              = graph.StringKind("AZGetKeys")
	GetSecrets                           = graph.StringKind("AZGetSecrets")
	HasRole                              = graph.StringKind("AZHasRole")
	EligibleRole                         = graph.StringKind("AZEligibleRole")
	EligibleGroup                        = graph.StringKind("AZEligibleGroup")
	MemberOf                             = graph.StringKind("AZMemberOf")
	Owner                                = graph.StringKind("AZOwner")
	RunsAs                               = graph.StringKind("AZRunsAs")
	VMContributor                        = graph.StringKind("AZVMContributor")
	AutomationContributor                = graph.StringKind("AZAutomationContributor")
	KeyVaultContributor                  = graph.StringKind("AZKeyVaultContributor")
	VMAdminLogin                         = graph.StringKind("AZVMAdminLogin")
	AddMembers                           = graph.StringKind("AZAddMembers")
	AddSecret                            = graph.StringKind("AZAddSecret")
	ExecuteCommand                       = graph.StringKind("AZExecuteCommand")
	GlobalAdmin                          = graph.StringKind("AZGlobalAdmin")
	PrivilegedAuthAdmin                  = graph.StringKind("AZPrivilegedAuthAdmin")
	Grant                                = graph.StringKind("AZGrant")
	GrantSelf                            = graph.StringKind("AZGrantSelf")
	PrivilegedRoleAdmin                  = graph.StringKind("AZPrivilegedRoleAdmin")
	ResetPassword                        = graph.StringKind("AZResetPassword")
	UserAccessAdministrator              = graph.StringKind("AZUserAccessAdministrator")
	Owns                                 = graph.StringKind("AZOwns")
	ScopedTo                             = graph.StringKind("AZScopedTo")
	CloudAppAdmin                        = graph.StringKind("AZCloudAppAdmin")
	AppAdmin                             = graph.StringKind("AZAppAdmin")
	AddOwner                             = graph.StringKind("AZAddOwner")
	ManagedIdentity                      = graph.StringKind("AZManagedIdentity")
	ApplicationReadWriteAll              = graph.StringKind("AZMGApplication_ReadWrite_All")
	AppRoleAssignmentReadWriteAll        = graph.StringKind("AZMGAppRoleAssignment_ReadWrite_All")
	DirectoryReadWriteAll                = graph.StringKind("AZMGDirectory_ReadWrite_All")
	GroupReadWriteAll                    = graph.StringKind("AZMGGroup_ReadWrite_All")
	GroupMemberReadWriteAll              = graph.StringKind("AZMGGroupMember_ReadWrite_All")
	RoleManagementReadWriteDirectory     = graph.StringKind("AZMGRoleManagement_ReadWrite_Directory")
	ServicePrincipalEndpointReadWriteAll = graph.StringKind("AZMGServicePrincipalEndpoint_ReadWrite_All")
	AKSContributor                       = graph.StringKind("AZAKSContributor")
	NodeResourceGroup                    = graph.StringKind("AZNodeResourceGroup")
	WebsiteContributor                   = graph.StringKind("AZWebsiteContributor")
	LogicAppContributor                  = graph.StringKind("AZLogicAppContributor")
	AZMGAddMember                        = graph.StringKind("AZMGAddMember")
	AZMGAddOwner                         = graph.StringKind("AZMGAddOwner")
	AZMGAddSecret                        = graph.StringKind("AZMGAddSecret")
	AZMGGrantAppRoles                    = graph.StringKind("AZMGGrantAppRoles")
	AZMGGrantRole                        = graph.StringKind("AZMGGrantRole")
)

type Property string

const (
	AppOwnerOrganizationID  Property = "appownerorganizationid"
	AppDescription          Property = "appdescription"
	AppDisplayName          Property = "appdisplayname"
	ServicePrincipalType    Property = "serviceprincipaltype"
	UserType                Property = "usertype"
	TenantID                Property = "tenantid"
	ServicePrincipalID      Property = "service_principal_id"
	OperatingSystemVersion  Property = "operatingsystemversion"
	TrustType               Property = "trustype"
	IsBuiltIn               Property = "isbuiltin"
	AppID                   Property = "appid"
	AppRoleID               Property = "approleid"
	DeviceID                Property = "deviceid"
	NodeResourceGroupID     Property = "noderesourcegroupid"
	OnPremID                Property = "onpremid"
	OnPremSyncEnabled       Property = "onpremsyncenabled"
	SecurityEnabled         Property = "securityenabled"
	SecurityIdentifier      Property = "securityidentifier"
	EnableRBACAuthorization Property = "enablerbacauthorization"
	Scope                   Property = "scope"
	Offer                   Property = "offer"
	MFAEnabled              Property = "mfaenabled"
	License                 Property = "license"
	Licenses                Property = "licenses"
	MFAEnforced             Property = "mfaenforced"
	UserPrincipalName       Property = "userprincipalname"
	IsAssignableToRole      Property = "isassignabletorole"
	PublisherDomain         Property = "publisherdomain"
	SignInAudience          Property = "signinaudience"
	RoleTemplateID          Property = "templateid"
)

func AllProperties() []Property {
	return []Property{AppOwnerOrganizationID, AppDescription, AppDisplayName, ServicePrincipalType, UserType, TenantID, ServicePrincipalID, OperatingSystemVersion, TrustType, IsBuiltIn, AppID, AppRoleID, DeviceID, NodeResourceGroupID, OnPremID, OnPremSyncEnabled, SecurityEnabled, SecurityIdentifier, EnableRBACAuthorization, Scope, Offer, MFAEnabled, License, Licenses, MFAEnforced, UserPrincipalName, IsAssignableToRole, PublisherDomain, SignInAudience, RoleTemplateID}
}
func ParseProperty(source string) (Property, error) {
	switch source {
	case "appownerorganizationid":
		return AppOwnerOrganizationID, nil
	case "appdescription":
		return AppDescription, nil
	case "appdisplayname":
		return AppDisplayName, nil
	case "serviceprincipaltype":
		return ServicePrincipalType, nil
	case "usertype":
		return UserType, nil
	case "tenantid":
		return TenantID, nil
	case "service_principal_id":
		return ServicePrincipalID, nil
	case "operatingsystemversion":
		return OperatingSystemVersion, nil
	case "trustype":
		return TrustType, nil
	case "isbuiltin":
		return IsBuiltIn, nil
	case "appid":
		return AppID, nil
	case "approleid":
		return AppRoleID, nil
	case "deviceid":
		return DeviceID, nil
	case "noderesourcegroupid":
		return NodeResourceGroupID, nil
	case "onpremid":
		return OnPremID, nil
	case "onpremsyncenabled":
		return OnPremSyncEnabled, nil
	case "securityenabled":
		return SecurityEnabled, nil
	case "securityidentifier":
		return SecurityIdentifier, nil
	case "enablerbacauthorization":
		return EnableRBACAuthorization, nil
	case "scope":
		return Scope, nil
	case "offer":
		return Offer, nil
	case "mfaenabled":
		return MFAEnabled, nil
	case "license":
		return License, nil
	case "licenses":
		return Licenses, nil
	case "mfaenforced":
		return MFAEnforced, nil
	case "userprincipalname":
		return UserPrincipalName, nil
	case "isassignabletorole":
		return IsAssignableToRole, nil
	case "publisherdomain":
		return PublisherDomain, nil
	case "signinaudience":
		return SignInAudience, nil
	case "templateid":
		return RoleTemplateID, nil
	default:
		return "", errors.New("Invalid enumeration value: " + source)
	}
}
func (s Property) String() string {
	switch s {
	case AppOwnerOrganizationID:
		return string(AppOwnerOrganizationID)
	case AppDescription:
		return string(AppDescription)
	case AppDisplayName:
		return string(AppDisplayName)
	case ServicePrincipalType:
		return string(ServicePrincipalType)
	case UserType:
		return string(UserType)
	case TenantID:
		return string(TenantID)
	case ServicePrincipalID:
		return string(ServicePrincipalID)
	case OperatingSystemVersion:
		return string(OperatingSystemVersion)
	case TrustType:
		return string(TrustType)
	case IsBuiltIn:
		return string(IsBuiltIn)
	case AppID:
		return string(AppID)
	case AppRoleID:
		return string(AppRoleID)
	case DeviceID:
		return string(DeviceID)
	case NodeResourceGroupID:
		return string(NodeResourceGroupID)
	case OnPremID:
		return string(OnPremID)
	case OnPremSyncEnabled:
		return string(OnPremSyncEnabled)
	case SecurityEnabled:
		return string(SecurityEnabled)
	case SecurityIdentifier:
		return string(SecurityIdentifier)
	case EnableRBACAuthorization:
		return string(EnableRBACAuthorization)
	case Scope:
		return string(Scope)
	case Offer:
		return string(Offer)
	case MFAEnabled:
		return string(MFAEnabled)
	case License:
		return string(License)
	case Licenses:
		return string(Licenses)
	case MFAEnforced:
		return string(MFAEnforced)
	case UserPrincipalName:
		return string(UserPrincipalName)
	case IsAssignableToRole:
		return string(IsAssignableToRole)
	case PublisherDomain:
		return string(PublisherDomain)
	case SignInAudience:
		return string(SignInAudience)
	case RoleTemplateID:
		return string(RoleTemplateID)
	default:
		panic("Invalid enumeration case: " + string(s))
	}
}
func (s Property) Name() string {
	switch s {
	case AppOwnerOrganizationID:
		return "App Owner Organization ID"
	case AppDescription:
		return "App Description"
	case AppDisplayName:
		return "App Display Name"
	case ServicePrincipalType:
		return "Service Principal Type"
	case UserType:
		return "User Type"
	case TenantID:
		return "Tenant ID"
	case ServicePrincipalID:
		return "Service Principal ID"
	case OperatingSystemVersion:
		return "Operating System Version"
	case TrustType:
		return "Trust Type"
	case IsBuiltIn:
		return "Is Built In"
	case AppID:
		return "App ID"
	case AppRoleID:
		return "App Role ID"
	case DeviceID:
		return "Device ID"
	case NodeResourceGroupID:
		return "Node Resource Group ID"
	case OnPremID:
		return "On Prem ID"
	case OnPremSyncEnabled:
		return "On Prem Sync Enabled"
	case SecurityEnabled:
		return "Security Enabled"
	case SecurityIdentifier:
		return "Security Identifier"
	case EnableRBACAuthorization:
		return "RBAC Authorization Enabled"
	case Scope:
		return "Scope"
	case Offer:
		return "Offer"
	case MFAEnabled:
		return "MFA Enabled"
	case License:
		return "License"
	case Licenses:
		return "Licenses"
	case MFAEnforced:
		return "MFA Enforced"
	case UserPrincipalName:
		return "User Principal Name"
	case IsAssignableToRole:
		return "Is Role Assignable"
	case PublisherDomain:
		return "Publisher Domain"
	case SignInAudience:
		return "Sign In Audience"
	case RoleTemplateID:
		return "Role Template ID"
	default:
		panic("Invalid enumeration case: " + string(s))
	}
}
func (s Property) Is(others ...graph.Kind) bool {
	for _, other := range others {
		if value, err := ParseProperty(other.String()); err == nil && value == s {
			return true
		}
	}
	return false
}
func Relationships() []graph.Kind {
	return []graph.Kind{AvereContributor, Contains, Contributor, GetCertificates, GetKeys, GetSecrets, HasRole, EligibleRole, EligibleGroup, MemberOf, Owner, RunsAs, VMContributor, AutomationContributor, KeyVaultContributor, VMAdminLogin, AddMembers, AddSecret, ExecuteCommand, GlobalAdmin, PrivilegedAuthAdmin, Grant, GrantSelf, PrivilegedRoleAdmin, ResetPassword, UserAccessAdministrator, Owns, ScopedTo, CloudAppAdmin, AppAdmin, AddOwner, ManagedIdentity, ApplicationReadWriteAll, AppRoleAssignmentReadWriteAll, DirectoryReadWriteAll, GroupReadWriteAll, GroupMemberReadWriteAll, RoleManagementReadWriteDirectory, ServicePrincipalEndpointReadWriteAll, AKSContributor, NodeResourceGroup, WebsiteContributor, LogicAppContributor, AZMGAddMember, AZMGAddOwner, AZMGAddSecret, AZMGGrantAppRoles, AZMGGrantRole}
}
func AppRoleTransitRelationshipKinds() []graph.Kind {
	return []graph.Kind{AZMGAddMember, AZMGAddOwner, AZMGAddSecret, AZMGGrantAppRoles, AZMGGrantRole}
}
func AbusableAppRoleRelationshipKinds() []graph.Kind {
	return []graph.Kind{ApplicationReadWriteAll, AppRoleAssignmentReadWriteAll, DirectoryReadWriteAll, GroupReadWriteAll, GroupMemberReadWriteAll, RoleManagementReadWriteDirectory, ServicePrincipalEndpointReadWriteAll}
}
func ControlRelationships() []graph.Kind {
	return []graph.Kind{AvereContributor, Contributor, Owner, VMContributor, AutomationContributor, KeyVaultContributor, AddMembers, AddSecret, ExecuteCommand, GlobalAdmin, Grant, GrantSelf, PrivilegedRoleAdmin, ResetPassword, UserAccessAdministrator, Owns, CloudAppAdmin, AppAdmin, AddOwner, ManagedIdentity, AKSContributor, WebsiteContributor, LogicAppContributor, AZMGAddMember, AZMGAddOwner, AZMGAddSecret, AZMGGrantAppRoles, AZMGGrantRole}
}
func ExecutionPrivileges() []graph.Kind {
	return []graph.Kind{VMAdminLogin, VMContributor, AvereContributor, WebsiteContributor, Contributor, ExecuteCommand}
}
func PathfindingRelationships() []graph.Kind {
	return []graph.Kind{AvereContributor, Contains, Contributor, GetCertificates, GetKeys, GetSecrets, HasRole, EligibleRole, EligibleGroup, MemberOf, Owner, RunsAs, VMContributor, AutomationContributor, KeyVaultContributor, VMAdminLogin, AddMembers, AddSecret, ExecuteCommand, GlobalAdmin, PrivilegedAuthAdmin, Grant, GrantSelf, PrivilegedRoleAdmin, ResetPassword, UserAccessAdministrator, Owns, CloudAppAdmin, AppAdmin, AddOwner, ManagedIdentity, AKSContributor, NodeResourceGroup, WebsiteContributor, LogicAppContributor, AZMGAddMember, AZMGAddOwner, AZMGAddSecret, AZMGGrantAppRoles, AZMGGrantRole}
}
func NodeKinds() []graph.Kind {
	return []graph.Kind{Entity, VMScaleSet, App, Role, Device, FunctionApp, Group, KeyVault, ManagementGroup, ResourceGroup, ServicePrincipal, Subscription, Tenant, User, VM, ManagedCluster, ContainerRegistry, WebApp, LogicApp, AutomationAccount}
}
