package layouts

import (
	"context"

	"github.com/svachaj/sambar-wall/modules/constants"
)

func hasRole(ctx context.Context, role string) bool {
	roles, _ := ctx.Value(constants.CTX_USER_ROLES).([]string)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// canSeeAttendance reports whether the signed-in user may open the attendance page (admins and instructors).
func canSeeAttendance(ctx context.Context) bool {
	return hasRole(ctx, constants.ROLE_SAMBAR_ADMIN) || hasRole(ctx, constants.ROLE_SAMBAR_INSTRUCTOR)
}

// isInstructorOnly reports whether the user is an instructor without admin rights
// (e.g. the shared instructor account), which has no use for the "Moje přihlášky" page.
func isInstructorOnly(ctx context.Context) bool {
	return hasRole(ctx, constants.ROLE_SAMBAR_INSTRUCTOR) && !hasRole(ctx, constants.ROLE_SAMBAR_ADMIN)
}
