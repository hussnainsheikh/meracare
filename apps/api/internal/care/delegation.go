package care

import "slices"

// Permission delegation.
//
// The rule is simple and absolute: nobody can grant access they do not
// themselves hold. Without it, `members.invite` would be a path to every other
// permission — invite an accomplice with rights you lack, or invite yourself
// back with more (docs/02, "Least Privilege").
//
// This is enforced in the domain rather than at the edge, so every caller that
// creates or changes a membership goes through the same check.

// InvitableRoles are the roles a member can be invited into.
//
// `senior` is absent deliberately: a circle has exactly one senior, and that
// seat belongs to the person the circle exists for. It is established when the
// profile is created, never handed out by invitation.
var InvitableRoles = []Role{RoleFamilyMember, RoleProfessionalCaregiver}

// CanInviteRole reports whether role may be granted by invitation.
func CanInviteRole(role Role) bool { return slices.Contains(InvitableRoles, role) }

// DelegationResult describes the outcome of checking a requested grant.
type DelegationResult struct {
	// Granted is the permission set the recipient will actually receive.
	Granted PermissionSet
	// Refused lists requested permissions the granter does not hold. A
	// non-empty value means the request was an escalation attempt and must be
	// rejected rather than quietly narrowed.
	Refused PermissionSet
}

// OK reports whether every requested permission could be granted.
func (r DelegationResult) OK() bool { return len(r.Refused) == 0 }

// Delegate works out what granter may confer on somebody else.
//
// requested may be empty, in which case the role's defaults are used. Defaults
// are still filtered through the granter's own permissions: a member must not
// be able to confer more than they hold merely by declining to specify a set.
//
// Unrecognised permission names are dropped before the comparison, so an
// invented name is never treated as an escalation the granter could satisfy.
func Delegate(granter PermissionSet, role Role, requested []Permission) DelegationResult {
	wanted := Normalise(requested)
	if len(wanted) == 0 {
		wanted = Normalise(DefaultPermissions(role))
	}

	result := DelegationResult{
		Granted: make(PermissionSet, 0, len(wanted)),
		Refused: make(PermissionSet, 0),
	}

	for _, permission := range wanted {
		if granter.Has(permission) {
			result.Granted = append(result.Granted, permission)
			continue
		}
		result.Refused = append(result.Refused, permission)
	}

	return result
}

// DelegateDefaults returns the permissions granter may confer on role when no
// explicit set is requested. It never refuses: the caller asked for "whatever
// is appropriate", so narrowing is the correct answer rather than an error.
func DelegateDefaults(granter PermissionSet, role Role) PermissionSet {
	result := Delegate(granter, role, nil)
	return result.Granted
}
