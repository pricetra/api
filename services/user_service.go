package services

import (
	"context"
	"fmt"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
	"github.com/pricetra/api/types"
	"github.com/pricetra/api/utils"
	"github.com/thanhpk/randstr"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

const EMAIL_VERIFICATION_CODE_LEN = 10

// Returns `false` if user email does not exist. Otherwise `true`
func (service Service) UserEmailExists(ctx context.Context, email string) bool {
	query := table.User.
		SELECT(table.User.Email.AS("email")).
		FROM(table.User).
		WHERE(table.User.Email.EQ(postgres.String(email))).
		LIMIT(1)
	var dest struct{ Email string }
	err := query.QueryContext(ctx, service.DbOrTxQueryable(), &dest)
	return err == nil
}

func (service Service) FindUserByEmail(ctx context.Context, email string) (gmodel.User, error) {
	qb := table.User.
		SELECT(table.User.AllColumns).
		WHERE(table.User.Email.EQ(postgres.String(email))).
		LIMIT(1)
	var user gmodel.User
	if err := qb.QueryContext(ctx, service.DbOrTxQueryable(), &user); err != nil {
		return gmodel.User{}, err
	}
	return user, nil
}

func (service Service) FindUserById(ctx context.Context, id int64) (gmodel.User, error) {
	qb := table.User.
		SELECT(table.User.AllColumns).
		WHERE(table.User.ID.EQ(postgres.Int(id))).
		LIMIT(1)
	var user gmodel.User
	if err := qb.QueryContext(ctx, service.DbOrTxQueryable(), &user); err != nil {
		return gmodel.User{}, err
	}
	return user, nil
}

func (service Service) CreateInternalUser(ctx context.Context, input gmodel.CreateAccountInput) (user gmodel.User, email_verification model.EmailVerification, err error) {
	if service.UserEmailExists(ctx, input.Email) {
		return gmodel.User{}, model.EmailVerification{}, fmt.Errorf("email already exists")
	}
	hashed_password, hash_err := service.HashPassword(input.Password)
	if hash_err != nil {
		return gmodel.User{}, model.EmailVerification{}, hash_err
	}

	// Create transaction
	tx, tx_err := service.DB.BeginTx(ctx, nil)
	if tx_err != nil {
		return gmodel.User{}, model.EmailVerification{}, tx_err
	}
	defer tx.Rollback()

	qb := table.User.
		INSERT(
			table.User.Email,
			table.User.Name,
			table.User.Password,
			table.User.Active,
			table.User.PhoneNumber,
		).
		MODEL(model.User{
			Email: input.Email,
			Name: input.Name,
			Password: &hashed_password,
			Active: false,
			PhoneNumber: input.PhoneNumber,
		}).
		RETURNING(table.User.AllColumns)
	if err := qb.QueryContext(ctx, tx, &user); err != nil {
		return gmodel.User{}, model.EmailVerification{}, fmt.Errorf("user entry could not be created. %s", err.Error())
	}
	service.TX = tx
	if email_verification, err = service.CreateEmailVerification(ctx, user); err != nil {
		return gmodel.User{}, model.EmailVerification{}, err
	}

	// Commit changes from transaction
	if err := tx.Commit(); err != nil {
		return gmodel.User{}, model.EmailVerification{}, err
	}
	return user, email_verification, nil
}

func (service Service) CreateOauthUser(ctx context.Context, input gmodel.CreateAccountInput, oauth_type model.UserAuthPlatformType) (gmodel.User, error) {
	var user gmodel.User
	qb := table.User.
		INSERT(
			table.User.Email,
			table.User.Name,
			table.User.Active,
			table.User.PhoneNumber,
		).
		MODEL(model.User{
			Email: input.Email,
			Name: input.Name,
			Active: true,
			PhoneNumber: input.PhoneNumber,
		}).
		RETURNING(table.User.AllColumns)
	if err := qb.QueryContext(ctx, service.DbOrTxQueryable(), &user); err != nil {
		return gmodel.User{}, fmt.Errorf("user entry could not be created. %s", err.Error())
	}
	return user, nil
}

func (service Service) GoogleAuthentication(ctx context.Context, access_token string, ip_address *string) (gmodel.Auth, error) {
	oauth_service, err := oauth2.NewService(ctx, option.WithoutAuthentication())
	if err != nil {
		return gmodel.Auth{}, fmt.Errorf("could not create service")
	}
	userinfo_service := oauth2.NewUserinfoService(oauth_service)
	userinfo, err := userinfo_service.Get().Do(googleapi.QueryParameter("access_token", access_token))
	if err != nil {
		return gmodel.Auth{}, fmt.Errorf("invalid access token")
	}

	var user gmodel.User
	new_user := false
	if service.UserEmailExists(ctx, userinfo.Email) {
		user, _ = service.FindUserByEmail(ctx, userinfo.Email)
	} else {
		user, err = service.CreateOauthUser(ctx, gmodel.CreateAccountInput{
			Email: userinfo.Email,
			Name: userinfo.Name,
		}, model.UserAuthPlatformType_Google)
		if err != nil {
			return gmodel.Auth{}, err
		}
		new_user = true
	}
	auth_state, err := service.CreateAuthStateWithJwt(ctx, user.ID, model.UserAuthPlatformType_Google, ip_address)
	if err == nil && new_user {
		auth_state.IsNewUser = &new_user
	}
	return auth_state, err
}

func (service Service) LoginInternal(ctx context.Context, email string, password string, ip_address *string) (gmodel.Auth, error) {
	db := service.DbOrTxQueryable()	
	query := table.User.
		SELECT(table.User.AllColumns).
		WHERE(table.User.Email.EQ(postgres.String(email))).
		LIMIT(1)
	var verify_user model.User
	if err := query.QueryContext(ctx, db, &verify_user); err != nil {
		return gmodel.Auth{}, fmt.Errorf("incorrect email or password")
	}
	if verify_user.Password == nil {
		return gmodel.Auth{}, fmt.Errorf("password has not been set for this account. try a different authentication method")
	}
	if !service.VerifyPasswordHash(password, *verify_user.Password) {
		return gmodel.Auth{}, fmt.Errorf("incorrect email or password")
	}

	return service.CreateAuthStateWithJwt(ctx, verify_user.ID, model.UserAuthPlatformType_Internal, ip_address)
}

func (service Service) CreateAuthStateWithJwt(
	ctx context.Context, 
	user_id int64, 
	auth_platform model.UserAuthPlatformType, 
	ip_address *string,
) (gmodel.Auth, error) {
	auth_state, auth_state_err := service.CreateAuthState(ctx, gmodel.User{ ID: user_id }, auth_platform, ip_address)
	if auth_state_err != nil {
		return gmodel.Auth{}, fmt.Errorf("could not create auth state")
	}

	qb := table.User.
		SELECT(
			table.User.AllColumns,
			table.AuthState.ID, 
			table.AuthState.Platform,
		).
		FROM(table.User.LEFT_JOIN(
			table.AuthState, 
			table.User.ID.EQ(table.AuthState.UserID),
		)).
		WHERE(postgres.AND(
			table.User.ID.EQ(postgres.Int(user_id)),
			table.AuthState.ID.EQ(postgres.Int(auth_state.ID)),
		)).LIMIT(1)
	var user gmodel.User
	if err := qb.QueryContext(ctx, service.DbOrTxQueryable(), &user); err != nil {
		return gmodel.Auth{}, fmt.Errorf("internal error")
	}

	// Generate JWT
	jwt, err := service.GenerateJWT(service.Tokens.JwtKey, &user)
	if err != nil {
		return gmodel.Auth{}, fmt.Errorf("could not generate JWT")
	}
	return gmodel.Auth{
		Token: jwt,
		User: &user,
	}, nil
}

// Given an existing user, create an auth_state row
func (service Service) CreateAuthState(ctx context.Context, user gmodel.User, auth_platform model.UserAuthPlatformType, ip_address *string) (model.AuthState, error) {
	query := table.AuthState.INSERT(
		table.AuthState.UserID,
		table.AuthState.IPAddress,
		table.AuthState.Platform,
	).MODEL(model.AuthState{
		UserID: user.ID,
		IPAddress: ip_address,
		Platform: auth_platform,
	}).RETURNING(table.AuthState.AllColumns)

	var auth_state model.AuthState
	err := query.QueryContext(ctx, service.DbOrTxQueryable(), &auth_state)
	return auth_state, err
}

func (service Service) CreateEmailVerification(ctx context.Context, user gmodel.User) (model.EmailVerification, error) {
	code := randstr.String(EMAIL_VERIFICATION_CODE_LEN)

	query := table.EmailVerification.INSERT(
		table.EmailVerification.UserID,
		table.EmailVerification.Code,
	).MODEL(model.EmailVerification{
		UserID: user.ID,
		Code: code,
	}).RETURNING(table.EmailVerification.AllColumns)

	var email_verification model.EmailVerification
	err := query.QueryContext(ctx, service.DbOrTxQueryable(), &email_verification)
	return email_verification, err
}

func (service Service) ResendEmailVerification(ctx context.Context, user gmodel.User) (email_verification model.EmailVerification, err error) {
	if user.Active {
		return model.EmailVerification{}, fmt.Errorf("user already has a verified email address")
	}
	service.TX, err = service.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.EmailVerification{}, err
	}
	defer service.TX.Rollback()

	_, err = table.EmailVerification.DELETE().
		WHERE(table.EmailVerification.UserID.EQ(postgres.Int(user.ID))).
		ExecContext(ctx, service.TX)
	if err != nil {
		return model.EmailVerification{}, fmt.Errorf("user email verification entry deletion failed")
	}

	email_verification, err = service.CreateEmailVerification(ctx, user)
	if err != nil {
		return model.EmailVerification{}, err
	}
	if err := service.TX.Commit(); err != nil {
		return model.EmailVerification{}, fmt.Errorf("could not commit changes")
	}
	return email_verification, nil
}

func (service Service) FindEmailVerificationByCode(ctx context.Context, verification_code string) (model.EmailVerification, error) {
	qb := table.EmailVerification.
		SELECT(table.EmailVerification.AllColumns).
		WHERE(table.EmailVerification.Code.EQ(postgres.String(verification_code))).
		LIMIT(1)
	var email_verification model.EmailVerification
	if err := qb.QueryContext(ctx, service.DbOrTxQueryable(), &email_verification); err != nil {
		return model.EmailVerification{}, fmt.Errorf("invalid email verification code")
	}
	return email_verification, nil
}

func (service Service) VerifyUserEmail(ctx context.Context, verification_code string) (gmodel.User, error) {
	var err error
	service.TX, err = service.DB.BeginTx(ctx, nil)
	if err != nil {
		return gmodel.User{}, err
	}
	defer service.TX.Rollback()

	email_verification, err := service.FindEmailVerificationByCode(ctx, verification_code)
	if err != nil {
		return gmodel.User{}, err
	}

	if time.Until(email_verification.CreatedAt).Abs() > time.Hour {
		// Delete verification entry since it's expired
		del_query := table.EmailVerification.
			DELETE().
			WHERE(table.EmailVerification.ID.EQ(postgres.Int(email_verification.ID)))
		if _, err := del_query.ExecContext(ctx, service.DB); err != nil {
			return gmodel.User{}, err
		}
		return gmodel.User{}, fmt.Errorf("verification code has expired")
	}

	update := table.User.
		UPDATE(table.User.Active, table.User.UpdatedAt).
		SET(postgres.Bool(true), postgres.DateT(time.Now())).
		WHERE(table.User.ID.EQ(postgres.Int(email_verification.UserID)))
	if _, err := update.ExecContext(ctx, service.TX); err != nil {
		return gmodel.User{}, fmt.Errorf("could not update user email verification status to verified")
	}

	// Remove email_verification row
	delete := table.EmailVerification.
		DELETE().
		WHERE(postgres.AND(
			table.EmailVerification.ID.EQ(postgres.Int(email_verification.ID)),
			table.EmailVerification.Code.EQ(postgres.String(verification_code)),
		))
	if _, err := delete.ExecContext(ctx, service.TX); err != nil {
		return gmodel.User{}, fmt.Errorf("could not delete email verification entry")
	}

	if err := service.TX.Commit(); err != nil {
		return gmodel.User{}, fmt.Errorf("could not commit changes")
	}
	service.TX = nil
	return service.FindUserById(ctx, email_verification.UserID)
}

func (Service) HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func (Service) VerifyPasswordHash(password string, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// Generates a JWT with claims, signed with key
func (Service) GenerateJWT(key string, user *gmodel.User) (string, error) {
	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": user.ID,
		"name": user.Name,
		"email": user.Email,
		"authPlatform": (*user.AuthPlatform).String(),
		"authStateId": *user.AuthStateID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})
	token, err := jwt.SignedString([]byte(key))
	if err != nil {
		return "", err
	}
	return token, nil
}

func (service Service) VerifyJwt(ctx context.Context, authorization types.AuthorizationKeyType) (gmodel.User, error) {
	jwt_raw, err := authorization.GetToken()
	if err != nil {
		return gmodel.User{}, err
	}

	claims, err := utils.GetJwtClaims(jwt_raw, service.Tokens.JwtKey)
	if err != nil {
		return gmodel.User{}, err
	}
	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return gmodel.User{}, fmt.Errorf("token expired")
	}

	authStateId := int64(claims["authStateId"].(float64))
	userId := int64(claims["id"].(float64))
	email := claims["email"].(string)
	query := table.User.
		SELECT(
			table.User.AllColumns,
			table.AuthState.ID,
			table.AuthState.Platform,
		).
		FROM(table.User.LEFT_JOIN(
			table.AuthState, 
			table.User.ID.EQ(table.AuthState.UserID),
		)).
		WHERE(
			table.User.ID.
				EQ(postgres.Int64(userId)).
				AND(table.User.Email.EQ(postgres.String(email))).
				AND(table.AuthState.ID.EQ(postgres.Int64(authStateId))),
		).
		LIMIT(1)
	var user gmodel.User
	if err := query.QueryContext(ctx, service.DbOrTxQueryable(), &user); err != nil {
		return gmodel.User{}, fmt.Errorf("one or more invalid claim values")
	}
	return user, nil
}

func (Service) GetAuthUserFromContext(ctx context.Context) gmodel.User {
	return ctx.Value(types.AuthUserKey).(gmodel.User)
}

func (service Service) UpdateUser(ctx context.Context, user gmodel.User, input gmodel.UpdateUser) (updated_user gmodel.User, err error) {
	columns := postgres.ColumnList{}
	if input.Name != nil {
		columns = append(columns, table.User.Name)
		user.Name = *input.Name
	}
	if input.Avatar != nil {
		columns = append(columns, table.User.Avatar)
		avatar_id := uuid.New().String()
		user.Avatar = &avatar_id
	}
	if input.BirthDate != nil {
		columns = append(columns, table.User.BirthDate)
		user.BirthDate = input.BirthDate
	}
	user.UpdatedAt = time.Now()
	qb := table.User.
		UPDATE(columns, table.User.UpdatedAt).
		MODEL(user).
		WHERE(table.User.ID.EQ(postgres.Int(user.ID))).
		RETURNING(table.User.AllColumns)
	
	if err = qb.QueryContext(ctx, service.DbOrTxQueryable(), &updated_user); err != nil {
		return gmodel.User{}, err
	}
	return updated_user, nil
}
