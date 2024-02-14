package agreement

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/middlewares"
	"github.com/svachaj/sambar-wall/utils"

	agreementTemplates "github.com/svachaj/sambar-wall/modules/agreement/templates"
)

type IAgreementHandlers interface {
	AgreementStartPage(c echo.Context) error
	CheckEmail(c echo.Context) error
}

type AgreementHandlers struct {
	db *sqlx.DB
}

func NewAgreementHandlers(db *sqlx.DB) IAgreementHandlers {
	return &AgreementHandlers{db: db}
}

func (h *AgreementHandlers) AgreementStartPage(c echo.Context) error {

	isAuthenticated, _ := middlewares.IsAuthenticated(&c)

	step1Page := Step1Page(h.db, isAuthenticated)

	return utils.HTML(c, step1Page)
}

func (h *AgreementHandlers) CheckEmail(c echo.Context) error {

	email := c.FormValue("email")

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM t_system_wall_user WHERE isenabled = 'true'f AND lower(email) = '%v'", strings.ToLower(email))
	err := h.db.Get(&count, query)

	if err != nil {
		log.Error().Msgf("CheckEmail error: %v", err)
		return c.HTML(200, "Doslo k chybe omlouváme se...")
	}

	// if count > 0 then user is already registered
	if count > 0 {
		return c.HTML(200, "User with this email is already registered.")
	}

	return c.HTML(200, `
						  <div class="rounded-md bg-yellow-50 p-4" id="toast-success" x-init="setTimeout(()=> removeElement('#toast-success'), 2000)">
						  <div class="flex">
							<div class="flex-shrink-0">
							  <svg class="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
								<path fill-rule="evenodd" d="M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.875c.673 1.167-.17 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 5a.75.75 0 01.75.75v3.5a.75.75 0 01-1.5 0v-3.5A.75.75 0 0110 5zm0 9a1 1 0 100-2 1 1 0 000 2z" clip-rule="evenodd" />
							  </svg>
							</div>
							<div class="ml-3">
							  <h3 class="text-sm font-medium text-yellow-800">Attention needed</h3>
							  <div class="mt-2 text-sm text-yellow-700">
								<p>Lorem ipsum dolor sit amet consectetur adipisicing elit. Aliquid pariatur, ipsum similique veniam quo totam eius aperiam dolorum.</p>
							  </div>
							</div>
						  </div>
						</div>`)
}

func Step1Page(db *sqlx.DB, isAuthenticated bool) templ.Component {
	step1Page := agreementTemplates.AgreementPage()

	return step1Page
}
