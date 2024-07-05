// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.747
package courses

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	types "github.com/svachaj/sambar-wall/db/types"
	"strconv"
)

func CourseTile(course types.CourseType) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"w-full max-w-5xl border border-neutral-600 shadow-xl rounded-xl bg-neutral-900 mb-10\"><div class=\"py-3 bg-slate-800 rounded-t-xl sticky top-[76px] z-10\"><h2 class=\"px-4 sm:text-xl text-lg text-center font-semibold leading-7 text-white sm:px-6 lg:px-8\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(course.Name)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `modules/courses/templates/course-tile.templ`, Line: 11, Col: 115}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</h2></div><div class=\"flex items-center justify-center py-4\"><h3 class=\"px-4 text-base text-justify leading-7 text-gray-200 sm:px-6 lg:px-8 max-w-2xl\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var3 string
		templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(course.Description)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `modules/courses/templates/course-tile.templ`, Line: 14, Col: 113}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</h3></div><table class=\"mt-2 w-full whitespace-nowrap text-left\"><colgroup><col class=\"w-24\"> <col class=\"w-24\"> <col class=\"w-fit\"> <col class=\"lg:w-1/12\"> <col class=\"lg:w-1/12\"> <col class=\"w-24\"></colgroup> <thead class=\"border-b border-white/10 text-sm leading-6 text-white\"><tr><th scope=\"col\" class=\"py-2 pl-4 pr-8 font-semibold sm:pl-6 lg:pl-8 text-center\">Den</th><th scope=\"col\" class=\"hidden py-2 pl-0 pr-8 font-semibold sm:table-cell text-center\">Čas</th><th scope=\"col\" class=\"hidden py-2 pl-0 pr-8 font-semibold sm:table-cell text-center\">Typ kurzu</th><th scope=\"col\" class=\"hidden py-2 pl-0 pr-8 font-semibold sm:table-cell text-center\">Věk</th><th scope=\"col\" class=\"hidden py-2 pl-0 pr-8 font-semibold sm:table-cell text-center\">Cena</th><th scope=\"col\" class=\"hidden py-2 pl-0 pr-8 font-semibold sm:table-cell text-center\"></th></tr></thead> <tbody class=\"divide-y divide-white/5\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		for _, course := range course.Courses {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<tr class=\"hover:bg-slate-800\"><td class=\"py-4 pl-4 pr-8 sm:pl-6 lg:pl-8\"><div class=\"flex items-center gap-x-4\"><div class=\"truncate text-sm font-medium leading-6 text-white\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var4 string
			templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(course.Days)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `modules/courses/templates/course-tile.templ`, Line: 40, Col: 84}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div></div></td><td class=\"hidden py-4 pl-0 pr-4 sm:table-cell sm:pr-8 text-white text-sm\"><div class=\"w-full text-center\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var5 string
			templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(course.TimeFrom.Format("15:04"))
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `modules/courses/templates/course-tile.templ`, Line: 45, Col: 41}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" - ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var6 string
			templ_7745c5c3_Var6, templ_7745c5c3_Err = templ.JoinStringErrs(course.TimeTo.Format("15:04"))
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `modules/courses/templates/course-tile.templ`, Line: 45, Col: 77}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var6))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div></td><td class=\"hidden py-4 pl-0 pr-4 sm:table-cell sm:pr-8 text-white text-sm\"><div class=\"w-full text-center\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var7 string
			templ_7745c5c3_Var7, templ_7745c5c3_Err = templ.JoinStringErrs(course.Name)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `modules/courses/templates/course-tile.templ`, Line: 50, Col: 21}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var7))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div></td><td class=\"hidden py-4 pl-0 pr-4 sm:table-cell sm:pr-8 text-white text-sm\"><div class=\"w-full text-center\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var8 string
			templ_7745c5c3_Var8, templ_7745c5c3_Err = templ.JoinStringErrs(course.AgeGroup)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `modules/courses/templates/course-tile.templ`, Line: 55, Col: 25}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var8))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div></td><td class=\"hidden py-4 pl-0 pr-4 sm:table-cell sm:pr-8 text-white text-sm\"><div class=\"w-full text-center\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var9 string
			templ_7745c5c3_Var9, templ_7745c5c3_Err = templ.JoinStringErrs(strconv.FormatFloat(course.Price, 'f', 2, 64))
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `modules/courses/templates/course-tile.templ`, Line: 60, Col: 55}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var9))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" Kč</div></td><td class=\"hidden py-4 pl-0 pr-4 sm:table-cell sm:pr-8 text-white text-sm\"><div class=\"w-full text-center\"><button class=\"bg-primary-600 text-white rounded-md px-2 py-2 hover:bg-primary-400 w-40\">Přihlásit se</button></div></td></tr>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</tbody></table></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}
