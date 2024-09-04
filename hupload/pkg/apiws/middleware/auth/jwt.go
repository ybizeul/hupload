package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	JWTAuthNoAuthorizationHeader = errors.New("JWTAuthMiddleware: no Authorization header")
	JWTAuthNoHMACSecret          = fmt.Errorf("JWTAuthMiddleware: no hmac secret")
	JWTAuthUnableToCreateToken   = errors.New("JWTAuthMiddleware: unable to create token")
	JWTAuthNoSubClaim            = errors.New("JWTAuthMiddleware: no sub in claim")

	shortTokenMinutesExpire = 5 * time.Minute
	longTokenMinutesExpire  = 20 * time.Minute
)

//var HMACSecret string

type JWTAuthMiddleware struct {
	HMACSecret string
}

func NewJWTAuthMiddleware(hmac string) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{HMACSecret: hmac}
}

func (j JWTAuthMiddleware) Middleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check JWT cookies
		shortCookie, _ := r.Cookie("X-Token")
		longCookie, _ := r.Cookie("X-Token-Refresh")

		upstreamUser, ok := AuthForRequest(r)

		if (shortCookie == nil && longCookie == nil) || (upstreamUser != "" && ok) {
			// Check that authentication has been previoulsy approved
			// If request is already authenticated, generate a JWT token
			if ok {
				short, long, err := j.generateTokens(upstreamUser)
				if err != nil {
					ServeNextError(next, w, r, err)
					return
				}

				http.SetCookie(w, &http.Cookie{Name: "X-Token", Value: short, Path: "/", Expires: time.Now().Add(shortTokenMinutesExpire)})
				http.SetCookie(w, &http.Cookie{Name: "X-Token-Refresh", Value: long, Path: "/", Expires: time.Now().Add(longTokenMinutesExpire)})

				ServeNextAuthenticated(upstreamUser, next, w, r)
				return
			}

			// Delete cookies
			http.SetCookie(w, &http.Cookie{Name: "X-Token", Value: "deleted", Path: "/", Expires: time.Unix(0, 0)})
			http.SetCookie(w, &http.Cookie{Name: "X-Token-Refresh", Value: "deleted", Path: "/", Expires: time.Unix(0, 0)})

			ServeNextError(next, w, r, JWTAuthNoAuthorizationHeader)
			return
		}

		var tokenString string
		if shortCookie == nil {
			tokenString = longCookie.Value
		} else {
			tokenString = shortCookie.Value
		}

		// Parse token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(j.HMACSecret), nil
		})

		if err != nil {
			http.SetCookie(w, &http.Cookie{Name: "X-Token", Value: "deleted", Path: "/", Expires: time.Unix(0, 0)})
			http.SetCookie(w, &http.Cookie{Name: "X-Token-Refresh", Value: "deleted", Path: "/", Expires: time.Unix(0, 0)})
			ServeNextError(next, w, r, fmt.Errorf("Unable to parse token: %w", err))
			return
		}

		if !token.Valid {
			http.SetCookie(w, &http.Cookie{Name: "X-Token", Value: "deleted", Path: "/", Expires: time.Unix(0, 0)})
			http.SetCookie(w, &http.Cookie{Name: "X-Token-Refresh", Value: "deleted", Path: "/", Expires: time.Unix(0, 0)})
			ServeNextError(next, w, r, errors.New("Invalid token"))
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			user, ok := claims["sub"].(string)
			if !ok || user == "" {
				ServeNextError(next, w, r, JWTAuthNoSubClaim)
				return
			}
			_, ok = claims["refresh"]
			if ok {
				if !ok {
					ServeNextError(next, w, r, JWTAuthNoSubClaim)
				}
				short, long, err := j.generateTokens(user)
				if err != nil {
					ServeNextError(next, w, r, err)
				}
				http.SetCookie(w, &http.Cookie{Name: "X-Token", Value: short, Path: "/", Expires: time.Now().Add(shortTokenMinutesExpire)})
				http.SetCookie(w, &http.Cookie{Name: "X-Token-Refresh", Value: long, Path: "/", Expires: time.Now().Add(longTokenMinutesExpire)})
			}
			ServeNextAuthenticated(user, next, w, r)
		} else {
			slog.Error("jwt decoding returned an invalid claim")
		}
	})
}

func (j *JWTAuthMiddleware) generateTokens(user string) (string, string, error) {
	var (
		err   error
		t     *jwt.Token
		short string
		long  string
	)

	// Generate short lived token
	t = jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss": "hupload",
			"sub": user,
			"exp": time.Now().Add(time.Minute * 5).Unix(),
		})

	short, err = t.SignedString([]byte(j.HMACSecret))

	if err != nil {
		return "", "", err
	}

	// Generate short lived token
	t = jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss":     "hupload",
			"sub":     user,
			"refresh": "true",
			"exp":     time.Now().Add(time.Minute * 20).Unix(),
		})

	long, err = t.SignedString([]byte(j.HMACSecret))

	if err != nil {
		return "", "", err
	}

	return short, long, nil
}
