// Package care holds the shared vocabulary of the care domain: the role a user
// holds in one senior's care circle, the permissions that role carries, and the
// lifecycle of the relationship between them.
//
// Authorization is relationship-based, never a global flag: the same user can
// be a senior for their own profile, a family member for a parent, and a
// professional caregiver for a client (docs/02-permissions-and-authorization.md).
//
// Keep this in sync with packages/contracts/src/care.ts.
package care

import "slices"

// Role is a user's relationship to one senior.
type Role string

const (
	// RoleSenior is the person the care circle exists for.
	RoleSenior Role = "senior"
	// RoleFamilyMember is a relative coordinating care.
	RoleFamilyMember Role = "family_member"
	// RoleProfessionalCaregiver is a paid caregiver, who may serve many seniors.
	RoleProfessionalCaregiver Role = "professional_caregiver"
)

// Roles lists every role the MVP supports.
var Roles = []Role{RoleSenior, RoleFamilyMember, RoleProfessionalCaregiver}

// Valid reports whether the role is one the MVP recognises.
func (r Role) Valid() bool { return slices.Contains(Roles, r) }

// RelationshipStatus is the lifecycle of a care relationship.
type RelationshipStatus string

const (
	// StatusPending is an invitation that has not been accepted yet.
	StatusPending RelationshipStatus = "pending"
	// StatusActive grants access.
	StatusActive RelationshipStatus = "active"
	// StatusRevoked is a membership that was removed. The row is kept so care
	// history retains its author (docs/07: preserve care-event history).
	StatusRevoked RelationshipStatus = "revoked"
)

// RelationshipStatuses lists every status.
var RelationshipStatuses = []RelationshipStatus{StatusPending, StatusActive, StatusRevoked}

// Valid reports whether the status is recognised.
func (s RelationshipStatus) Valid() bool { return slices.Contains(RelationshipStatuses, s) }

// Permission is a single capability within one senior's care circle.
//
// Permissions are stored per relationship rather than derived from the role at
// request time, so a care circle can narrow or widen an individual's access
// without inventing new roles.
type Permission string

const (
	PermissionSeniorView Permission = "senior.view"
	PermissionSeniorEdit Permission = "senior.edit"

	PermissionTasksView     Permission = "tasks.view"
	PermissionTasksManage   Permission = "tasks.manage"
	PermissionTasksComplete Permission = "tasks.complete"

	PermissionMedicationsView   Permission = "medications.view"
	PermissionMedicationsManage Permission = "medications.manage"
	PermissionMedicationsRecord Permission = "medications.record"

	PermissionAppointmentsView   Permission = "appointments.view"
	PermissionAppointmentsManage Permission = "appointments.manage"

	PermissionNotesView   Permission = "notes.view"
	PermissionNotesCreate Permission = "notes.create"

	PermissionActivityView Permission = "activity.view"

	PermissionMembersView   Permission = "members.view"
	PermissionMembersInvite Permission = "members.invite"
	PermissionMembersManage Permission = "members.manage"

	PermissionMessagesParticipate Permission = "messages.participate"
)

// Permissions lists every permission the MVP recognises. The database CHECK
// constraint on care_relationships.permissions mirrors this list.
var Permissions = []Permission{
	PermissionSeniorView,
	PermissionSeniorEdit,
	PermissionTasksView,
	PermissionTasksManage,
	PermissionTasksComplete,
	PermissionMedicationsView,
	PermissionMedicationsManage,
	PermissionMedicationsRecord,
	PermissionAppointmentsView,
	PermissionAppointmentsManage,
	PermissionNotesView,
	PermissionNotesCreate,
	PermissionActivityView,
	PermissionMembersView,
	PermissionMembersInvite,
	PermissionMembersManage,
	PermissionMessagesParticipate,
}

// Valid reports whether the permission is recognised.
func (p Permission) Valid() bool { return slices.Contains(Permissions, p) }

// defaultPermissions are the permissions granted when a relationship is created
// without an explicit set. They follow the role definitions in
// docs/02-permissions-and-authorization.md.
var defaultPermissions = map[Role][]Permission{
	// The senior owns their own care data and may run the circle themselves,
	// which is what makes Solo Mode work without anybody else involved.
	RoleSenior: {
		PermissionSeniorView,
		PermissionSeniorEdit,
		PermissionTasksView,
		PermissionTasksManage,
		PermissionTasksComplete,
		PermissionMedicationsView,
		PermissionMedicationsManage,
		PermissionMedicationsRecord,
		PermissionAppointmentsView,
		PermissionAppointmentsManage,
		PermissionNotesView,
		PermissionNotesCreate,
		PermissionActivityView,
		PermissionMembersView,
		PermissionMembersInvite,
		PermissionMembersManage,
		PermissionMessagesParticipate,
	},

	// Family members coordinate care: they schedule the work and can bring
	// other people in, but they do not administer the circle's membership.
	RoleFamilyMember: {
		PermissionSeniorView,
		PermissionSeniorEdit,
		PermissionTasksView,
		PermissionTasksManage,
		PermissionTasksComplete,
		PermissionMedicationsView,
		PermissionMedicationsManage,
		PermissionMedicationsRecord,
		PermissionAppointmentsView,
		PermissionAppointmentsManage,
		PermissionNotesView,
		PermissionNotesCreate,
		PermissionActivityView,
		PermissionMembersView,
		PermissionMembersInvite,
		PermissionMessagesParticipate,
	},

	// Professional caregivers carry out care. They deliberately cannot edit the
	// profile, restructure the schedule, or see or change who else is involved
	// beyond the member list — docs/02 requires that they not automatically
	// receive family or private information.
	RoleProfessionalCaregiver: {
		PermissionSeniorView,
		PermissionTasksView,
		PermissionTasksComplete,
		PermissionMedicationsView,
		PermissionMedicationsRecord,
		PermissionAppointmentsView,
		PermissionNotesView,
		PermissionNotesCreate,
		PermissionActivityView,
		PermissionMembersView,
		PermissionMessagesParticipate,
	},
}

// DefaultPermissions returns a fresh copy of the permissions granted to role.
// An unknown role yields no permissions, so a mistake fails closed.
func DefaultPermissions(role Role) []Permission {
	return slices.Clone(defaultPermissions[role])
}

// PermissionSet is the set of permissions held by one relationship.
type PermissionSet []Permission

// Has reports whether the set grants permission.
func (s PermissionSet) Has(permission Permission) bool {
	return slices.Contains(s, permission)
}

// HasAll reports whether the set grants every listed permission.
func (s PermissionSet) HasAll(permissions ...Permission) bool {
	for _, permission := range permissions {
		if !s.Has(permission) {
			return false
		}
	}
	return true
}

// Normalise removes unrecognised and duplicate permissions and sorts the
// result, so what is stored is always a clean, comparable set.
//
// Unknown values are dropped rather than rejected: a client cannot widen its
// own access by inventing a permission name.
func Normalise(permissions []Permission) PermissionSet {
	seen := make(map[Permission]struct{}, len(permissions))
	cleaned := make(PermissionSet, 0, len(permissions))

	for _, permission := range permissions {
		if !permission.Valid() {
			continue
		}
		if _, duplicate := seen[permission]; duplicate {
			continue
		}
		seen[permission] = struct{}{}
		cleaned = append(cleaned, permission)
	}

	slices.Sort(cleaned)
	return cleaned
}

// Strings converts a permission set for database storage.
func (s PermissionSet) Strings() []string {
	values := make([]string, len(s))
	for i, permission := range s {
		values[i] = string(permission)
	}
	return values
}

// PermissionsFromStrings converts stored values back into a permission set.
func PermissionsFromStrings(values []string) PermissionSet {
	permissions := make([]Permission, len(values))
	for i, value := range values {
		permissions[i] = Permission(value)
	}
	return Normalise(permissions)
}
