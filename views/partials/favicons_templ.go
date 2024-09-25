// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.778
package partials

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func Favicons() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
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
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<link rel=\"icon\" type=\"image/png\" sizes=\"16x16\" href=\"/favicon/favicon-16x16.png\"><link rel=\"icon\" type=\"image/png\" sizes=\"32x32\" href=\"/favicon/favicon-32x32.png\"><link rel=\"icon\" type=\"image/png\" sizes=\"96x96\" href=\"/favicon/favicon-96x96.png\"><link rel=\"icon\" type=\"image/png\" sizes=\"192x192\" href=\"/favicon/favicon-192x192.png\"><link rel=\"apple-touch-icon\" sizes=\"57x57\" href=\"/favicon/favicon-57x57.png\"><link rel=\"apple-touch-icon\" sizes=\"60x60\" href=\"/favicon/favicon-60x60.png\"><link rel=\"apple-touch-icon\" sizes=\"72x72\" href=\"/favicon/favicon-72x72.png\"><link rel=\"apple-touch-icon\" sizes=\"76x76\" href=\"/favicon/favicon-76x76.png\"><link rel=\"apple-touch-icon\" sizes=\"114x114\" href=\"/favicon/favicon-114x114.png\"><link rel=\"apple-touch-icon\" sizes=\"120x120\" href=\"/favicon/favicon-120x120.png\"><link rel=\"apple-touch-icon\" sizes=\"144x144\" href=\"/favicon/favicon-144x144.png\"><link rel=\"apple-touch-icon\" sizes=\"152x152\" href=\"/favicon/favicon-152x152.png\"><link rel=\"apple-touch-icon\" sizes=\"180x180\" href=\"/favicon/favicon-180x180.png\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
