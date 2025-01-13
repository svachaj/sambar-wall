package constants

const (
	ROUTE_HOME = "/"

	ROUTE_AGREEMENT_START_PAGE  = "/souhlas-s-provoznim-radem"
	ROUTE_AGREEMENT_CHECK_EMAIL = "/agreement/check-email"
	ROUTE_AGREEMENT_FINALIZE    = "/agreement/finalize"

	ROUTE_LOGIN            = "/prihlaseni"
	ROUTE_LOGIN_STEP1      = "/sign-in-step1"
	ROUTE_LOGIN_STEP2      = "/sign-in-step2"
	ROUTE_LOGIN_MAGIC_LINK = "/sign-me-in"

	ROUTE_SIGN_OUT = "/sign-out"

	ROUTE_USER_ACCOUNT = "/ucet"

	ROUTE_COURSES                              = "/kurzy"
	ROUTE_COURSES_APPLICATION_FORM_PAGE        = "/prihlaska/:id"
	ROUTE_COURSES_APPLICATION_FORM             = "/prihlaska"
	ROUTE_COURSES_APPLICATION_FORM_EDIT        = "/prihlaska-edit"
	ROUTE_COURSES_APPLICATION_FORM_EDIT_CANCEL = "/prihlaska-edit-cancel"
	ROUTE_COURSES_APPLICATION_FORM_EDIT_ID     = "/prihlaska-edit/:id"
	ROUTE_COURSES_MY_APPLICATIONS              = "/moje-prihlasky"
	ROUTE_COURSES_APPLICATION_FORMS_REUSE      = "/prihlasky-opakovat"

	ROUTE_COURSES_APPLICATION_FORMS         = "/prihlasky"
	ROUTE_COURSES_APPLICATION_FORMS_SEARCH  = "/prihlasky-hledat"
	ROUTE_COURSES_APPLICATION_FORM_SET_PAID = "/prihlaska/:id"
)
