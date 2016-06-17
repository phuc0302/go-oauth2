package oauth2

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/phuc0302/go-oauth2/mongo"
)

// DefaultFactory describes a default factory object.
type DefaultFactory struct {
}

// CreateRequestContext creates new request context.
func (d *DefaultFactory) CreateRequestContext(request *http.Request, response http.ResponseWriter) *Request {
	context := &Request{
		Path:     request.URL.Path,
		request:  request,
		response: response,
	}

	// Format request headers
	if len(request.Header) > 0 {
		context.Header = make(map[string]string)

		for k, v := range request.Header {
			if header := strings.ToLower(k); header == "authorization" {
				context.Header[header] = v[0]
			} else {
				context.Header[header] = strings.ToLower(v[0])
			}
		}
	}

	// Parse body context if neccessary
	var params url.Values
	switch context.request.Method {

	case GET:
		params = request.URL.Query()
		break

	case POST, PATCH:
		if contentType := context.Header["content-type"]; contentType == "application/x-www-form-urlencoded" {
			err := request.ParseForm()
			if err == nil {
				params = request.Form
			}
		} else if strings.HasPrefix(contentType, "multipart/form-data; boundary") {
			err := request.ParseMultipartForm(Cfg.MultipartSize)
			if err == nil {
				params = request.MultipartForm.Value
			}
		}
		break

	default:
		break
	}

	// Process params
	if len(params) > 0 {
		context.QueryParams = make(map[string]string)

		for k, v := range params {
			context.QueryParams[k] = v[0]
		}
	}
	return context
}

// CreateSecurityContext creates new security context.
func (d *DefaultFactory) CreateSecurityContext(c *Request) *Security {
	tokenString := c.Header["authorization"]

	/* Condition validation: Validate existing of authorization header */
	if isBearer := bearerFinder.MatchString(tokenString); !isBearer {
		tokenString = c.QueryParams["access_token"]
		if len(tokenString) <= 0 {
			return nil
		} else {
			delete(c.QueryParams, "access_token")
		}
	} else {
		tokenString = tokenString[7:]
	}

	/* Condition validation: Validate expiration time */
	if accessToken := TokenStore.FindAccessToken(tokenString); accessToken == nil || accessToken.IsExpired() {
		return nil
	} else {
		client := TokenStore.FindClientWithID(accessToken.ClientID())
		user := TokenStore.FindUserWithID(accessToken.UserID())
		securityContext := &Security{
			AuthClient:      client,
			AuthUser:        user,
			AuthAccessToken: accessToken,
		}
		return securityContext
	}
}

// CreateRoute creates new route component.
func (d *DefaultFactory) CreateRoute(urlPattern string) IRoute {
	regexPattern := pathFinder.ReplaceAllStringFunc(urlPattern, func(m string) string {
		return fmt.Sprintf(`(?P<%s>[^/#?]+)`, m[1:len(m)-1])
	})
	regexPattern += "/?"

	route := DefaultRoute{
		path:     urlPattern,
		handlers: map[string]interface{}{},
		regex:    regexp.MustCompile(regexPattern),
	}
	return &route
}

// CreateRouter creates new router component.
func (d *DefaultFactory) CreateRouter() IRouter {
	return &DefaultRouter{}
}

// CreateStore creates new store component.
func (d *DefaultFactory) CreateStore() IStore {
	mongo.ConnectMongo()
	return &DefaultMongoStore{}
}
