package oauth2

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/phuc0302/go-oauth2/util"
)

// TokenGrant describes a token grant controller.
type TokenGrant struct {
}

// HandleForm validates authentication form.
func (g *TokenGrant) HandleForm(c *RequestContext, s *OAuthContext) {
	oauthContext := new(OAuthContext)

	if err := g.validateForm(c, oauthContext); err == nil {
		g.finalizeToken(c, oauthContext)
	} else {
		c.OutputError(err)
	}
}

// validateForm validates general information
func (g *TokenGrant) validateForm(c *RequestContext, s *OAuthContext) *util.Status {
	// If client_id and client_secret are not include, try to look at the authorization header
	if c.QueryParams != nil && len(c.QueryParams["client_id"]) == 0 && len(c.QueryParams["client_secret"]) == 0 {
		c.QueryParams["client_id"], c.QueryParams["client_secret"], _ = c.request.BasicAuth()
	}

	// Bind
	var inputForm struct {
		GrantType    string `field:"grant_type"`
		ClientID     string `field:"client_id" validation:"^[0-9a-fA-F]{24}$"`
		ClientSecret string `field:"client_secret" validation:"^[0-9a-fA-F]{24}$"`
	}
	err := c.BindForm(&inputForm)

	/* Condition validation: Validate grant_type */
	if !grantsValidation.MatchString(inputForm.GrantType) {
		return util.Status400WithDescription(fmt.Sprintf(invalidParameter, "grant_type"))
	}

	/* Condition validation: Validate binding process */
	if err != nil {
		return util.Status400WithDescription(err.Error())
	}

	/* Condition validation: Check the store */
	recordClient := store.FindClientWithCredential(inputForm.ClientID, inputForm.ClientSecret)
	if recordClient == nil {
		return util.Status400WithDescription(fmt.Sprintf(invalidParameter, "client_id or client_secret"))
	}

	/* Condition validation: Check grant_type for client */
	clientGrantsValidation := regexp.MustCompile(fmt.Sprintf("^(%s)$", strings.Join(recordClient.GrantTypes(), "|")))
	if isGranted := clientGrantsValidation.MatchString(inputForm.GrantType); !isGranted {
		return util.Status400WithDescription("The \"grant_type\" is unauthorised for this \"client_id\".")
	}
	s.Client = recordClient

	// Choose authentication flow
	switch inputForm.GrantType {

	case AuthorizationCodeGrant:
		// TODO: Going to do soon
		//		g.handleAuthorizationCodeGrant(c, values, queryClient)
		break

		//	case ImplicitGrant:
		// TODO: Going to do soon
		//		break

	case ClientCredentialsGrant:
		// TODO: Going to do soon
		//		g.handleClientCredentialsGrant()
		break

	case PasswordGrant:
		return g.passwordFlow(c, s)

	case RefreshTokenGrant:
		return g.refreshTokenFlow(c, s)
	}
	return nil
}

func (t *TokenGrant) handleAuthorizationCodeGrant(c *RequestContext, values url.Values, client Client) {
	//	/* Condition validation: Validate redirect_uri */
	//	if len(queryClient.RedirectURI) == 0 {
	//		err := util.Status400WithDescription("Missing redirect_uri parameter.")
	//		c.OutputError(err)
	//		return
	//	}

	//	/* Condition validation: Check redirect_uri for client */
	//	isAllowRedirectURI := false
	//	for _, redirectURI := range recordClient.RedirectURIs {
	//		if redirectURI == queryClient.RedirectURI {
	//			isAllowRedirectURI = true
	//			break
	//		}
	//	}
	//	if !isAllowRedirectURI {
	//		err := util.Status400WithDescription("The redirect_uri had not been registered for this client_id.")
	//		c.OutputError(err)
	//		return
	//	}

	/* Condition validation: Validate code */
	authorizationCode := values.Get("code")
	if len(authorizationCode) == 0 {
		err := util.Status400()
		err.Description = "Missing code parameter."
		c.OutputError(err)
		return
	}

	//	t.store.FindAuthorizationCode(authorizationCode)
	// this.model.getAuthCode(code, function (err, authCode) {

	//   if (!authCode || authCode.clientId !== self.client.clientId) {
	//     return done(error('invalid_grant', 'Invalid code'));
	//   } else if (authCode.expires < self.now) {
	//     return done(error('invalid_grant', 'Code has expired'));
	//   }

	//   self.user = authCode.user || { id: authCode.userId };
	//   if (!self.user.id) {
	//     return done(error('server_error', false,
	//       'No user/userId parameter returned from getauthCode'));
	//   }

	//   done();
	// });
}

func (t *TokenGrant) handleClientCredentialsGrant() {
	// // Client credentials
	// var clientId = this.client.clientId,
	//   clientSecret = this.client.clientSecret;

	// if (!clientId || !clientSecret) {
	//   return done(error('invalid_client',
	//     'Missing parameters. "client_id" and "client_secret" are required'));
	// }

	// var self = this;
	// return this.model.getUserFromClient(clientId, clientSecret,
	//     function (err, user) {
	//   if (err) return done(error('server_error', false, err));
	//   if (!user) {
	//     return done(error('invalid_grant', 'Client credentials are invalid'));
	//   }

	//   self.user = user;
	//   done();
	// });
}

// passwordFlow implements user's authentication with user's credential.
func (g *TokenGrant) passwordFlow(c *RequestContext, s *OAuthContext) *util.Status {
	var passwordForm struct {
		Username string `field:"username" validation:"^\\w+$"`
		Password string `field:"password" validation:"^\\w+$"`
	}
	c.BindForm(&passwordForm)

	/* Condition validation: Validate username and password parameters */
	if len(passwordForm.Username) == 0 || len(passwordForm.Password) == 0 {
		return util.Status400WithDescription(fmt.Sprintf(invalidParameter, "username or password"))
	}

	/* Condition validation: Validate user's credentials */
	if recordUser := store.FindUserWithCredential(passwordForm.Username, passwordForm.Password); recordUser != nil {
		s.User = recordUser
		return nil
	}
	return util.Status400WithDescription(fmt.Sprintf(invalidParameter, "username or password"))
}

// useRefreshTokenFlow handle refresh token flow.
func (g *TokenGrant) refreshTokenFlow(c *RequestContext, s *OAuthContext) *util.Status {
	/* Condition validation: Validate refresh_token parameter */
	if queryToken := c.QueryParams["refresh_token"]; len(queryToken) > 0 {

		/* Condition validation: Validate refresh_token */
		refreshToken := store.FindRefreshToken(queryToken)

		if refreshToken == nil || refreshToken.ClientID() != s.Client.ClientID() {
			return util.Status400WithDescription(fmt.Sprintf(invalidParameter, "refresh_token"))
		} else if refreshToken.IsExpired() {
			return util.Status400WithDescription("\refresh_token\" is expired.")
		}
		s.User = store.FindUserWithID(refreshToken.UserID())

		// Delete current access token
		accessToken := store.FindAccessTokenWithCredential(refreshToken.ClientID(), refreshToken.UserID())
		store.DeleteAccessToken(accessToken)

		// Delete current refresh token
		store.DeleteRefreshToken(refreshToken)
		refreshToken = nil

		// Update security context
		s.RefreshToken = nil
		s.AccessToken = nil

		// Delete current refresh token
		return nil
	}
	return util.Status400WithDescription(fmt.Sprintf(invalidParameter, "refresh_token"))
}

// finalizeToken summary and return result to client.
func (g *TokenGrant) finalizeToken(c *RequestContext, s *OAuthContext) {
	now := time.Now()

	// Generate access token if neccessary
	if s.AccessToken == nil {
		accessToken := store.FindAccessTokenWithCredential(s.Client.ClientID(), s.User.UserID())
		if accessToken != nil && accessToken.IsExpired() {
			store.DeleteAccessToken(accessToken)
			accessToken = nil
		}

		if accessToken == nil {
			accessToken = store.CreateAccessToken(
				s.Client.ClientID(),
				s.User.UserID(),
				now,
				now.Add(Cfg.AccessTokenDuration),
			)
		}
		s.AccessToken = accessToken
	}

	// Generate refresh token if neccessary
	if Cfg.AllowRefreshToken && s.RefreshToken == nil {
		refreshToken := store.FindRefreshTokenWithCredential(s.Client.ClientID(), s.User.UserID())
		if refreshToken != nil && refreshToken.IsExpired() {
			store.DeleteRefreshToken(refreshToken)
			refreshToken = nil
		}

		if refreshToken == nil {
			refreshToken = store.CreateRefreshToken(
				s.Client.ClientID(),
				s.User.UserID(),
				now,
				now.Add(Cfg.RefreshTokenDuration),
			)
		}
		s.RefreshToken = refreshToken
	}

	// Generate response token
	tokenResponse := &TokenResponse{
		TokenType:   "Bearer",
		AccessToken: s.AccessToken.Token(),
		ExpiresIn:   s.AccessToken.ExpiredTime().Unix() - time.Now().UTC().Unix(),
		Roles:       s.User.UserRoles(),
	}

	// Only add refresh_token if allowed
	if Cfg.AllowRefreshToken {
		tokenResponse.RefreshToken = s.RefreshToken.Token()
	}
	c.OutputJSON(util.Status200(), tokenResponse)
}
