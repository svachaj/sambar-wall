package constants

const ROLE_SAMBAR_ADMIN = "SADM"
const ROLE_SAMBAR_RECEPTION = "recepce"
const ROLE_SAMBAR_INSTRUCTOR = "instruktor"

type contextKey string

// CTX_USER_ROLES is the request context key holding the signed-in user's roles ([]string).
const CTX_USER_ROLES contextKey = "userRoles"
