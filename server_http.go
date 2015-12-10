package oauth2

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/phuc0302/go-oauth2/utils"
)

// ServeHTTP handle HTTP request and HTTP response.
func (s *Server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	// Format reuest before process
	request.URL.Path = utils.FormatPath(request.URL.Path)
	request.Method = strings.ToUpper(request.Method)

	// Create context
	context := CreateRequestContext(request, response)
	defer RecoveryRequest(context, s.Development)

	// Validate http request methods
	if !s.methodsValidation.MatchString(request.Method) {
		context.OutputError(utils.Status405())
		return
	}

	// Should redirect request to static folder or not?
	isStaticRequest := false
	if request.Method == GET && len(s.StaticFolders) > 0 {
		for prefix, folder := range s.StaticFolders {
			if strings.HasPrefix(request.URL.Path, prefix) {
				newPath := strings.Replace(request.URL.Path, prefix, folder, 1)
				request.URL, _ = url.Parse(newPath)
				isStaticRequest = true
				break
			}
		}
	}

	if !isStaticRequest {
		s.serveRequest(context)
	} else {
		s.serveResource(context, request, response)
	}
}

// MARK: Struct's private functions
func (s *Server) serveRequest(context *RequestContext) {
	// FIX FIX FIX: Add priority here so that we can move the mosted used node to top

	for _, route := range s.routes {
		ok, pathQueries := route.Match(context.Method(), context.URLPath)
		if !ok {
			continue
		}

		// Validate authentication & roles if neccessary
		var securityContext *SecurityContext = nil
		if s.tokenStore != nil {
			securityContext = CreateSecurityContextWithRequestContext(context, s.tokenStore)
			for rule, _ := range s.userRoles {
				if rule.MatchString(context.URLPath) {
					if securityContext.AuthUser != nil {

					} else {
						//						context.OutputError(status)
						return
					}
					break
				}
			}
		}

		context.PathQueries = pathQueries
		route.InvokeHandler(context)
		return
	}

	context.OutputError(utils.Status503())
}

func (s *Server) serveResource(context *RequestContext, request *http.Request, response http.ResponseWriter) {
	resourcePath := request.URL.Path

	/* Condition validation: Check if file exist or not */
	if !utils.FileExisted(resourcePath) {
		context.OutputError(utils.Status404())
		return
	}

	// Open file as read only
	file, err := os.Open(resourcePath)
	defer file.Close()

	if err != nil {
		context.OutputError(utils.Status404())
		return
	}

	/* Condition validation: Only serve file, not directory */
	info, _ := file.Stat()
	if info.IsDir() {
		context.OutputError(utils.Status403())
		return
	}
	http.ServeContent(response, request, resourcePath, info.ModTime(), file)
}
