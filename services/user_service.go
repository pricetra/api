package services

import (
	"context"
	"fmt"
	"strings"
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

const EMAIL_VERIFICATION_CODE_LEN = 6
const PASSWORD_RESET_CODE_LEN = 6
const PASSWORD_RESET_MAX_TRIES = 10

// Returns `false` if user email does not exist. Otherwise `true`
func (s Service) UserEmailExists(ctx context.Context, email string) bool {
	query := table.User.
		SELECT(table.User.Email.AS("email")).
		FROM(table.User).
		WHERE(table.User.Email.EQ(postgres.String(email))).
		LIMIT(1)
	var dest struct{ Email string }
	err := query.QueryContext(ctx, s.DbOrTxQueryable(), &dest)
	return err == nil
}

func (s Service) FindUserByEmail(ctx context.Context, email string) (gmodel.User, error) {
	qb := table.User.
		SELECT(table.User.AllColumns).
		WHERE(table.User.Email.EQ(postgres.String(email))).
		LIMIT(1)
	var user gmodel.User
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &user); err != nil {
		return gmodel.User{}, err
	}
	return user, nil
}

func (s Service) FindUserById(ctx context.Context, id int64) (gmodel.User, error) {
	qb := table.User.
		SELECT(table.User.AllColumns).
		WHERE(table.User.ID.EQ(postgres.Int(id))).
		LIMIT(1)
	var user gmodel.User
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &user); err != nil {
		return gmodel.User{}, err
	}
	return user, nil
}

func (s Service) FindAuthUserById(ctx context.Context, user_id int64, auth_state_id string) (user gmodel.User, err error) {
	auth_state_uuid, err := uuid.Parse(auth_state_id)
	if err != nil {
		return gmodel.User{}, fmt.Errorf("invalid uuid value")
	}

	qb := table.User.
		SELECT(
			table.User.AllColumns,
			table.AuthState.ID,
			table.AuthState.Platform,
			table.AuthState.DeviceType,
			table.AuthState.ExpoPushToken,
			table.Address.AllColumns,
			table.Country.Name,
		).
		FROM(
			table.User.
				LEFT_JOIN(table.AuthState, table.User.ID.EQ(table.AuthState.UserID)).
				LEFT_JOIN(table.Address, table.Address.ID.EQ(table.User.AddressID)).
				LEFT_JOIN(table.Country, table.Country.Code.EQ(table.Address.CountryCode)),
		).
		WHERE(
			table.User.ID.
				EQ(postgres.Int64(user_id)).
				AND(table.AuthState.ID.EQ(postgres.UUID(auth_state_uuid))),
		).
		LIMIT(1)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &user)
	return user, err
} 

func (s Service) CreateInternalUser(ctx context.Context, input gmodel.CreateAccountInput) (user gmodel.User, email_verification model.EmailVerification, err error) {
	if s.UserEmailExists(ctx, input.Email) {
		return gmodel.User{}, model.EmailVerification{}, fmt.Errorf("email already exists")
	}
	hashed_password, hash_err := s.HashPassword(input.Password)
	if hash_err != nil {
		return gmodel.User{}, model.EmailVerification{}, hash_err
	}

	// Create transaction
	tx, tx_err := s.DB.BeginTx(ctx, nil)
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
	s.TX = tx
	if email_verification, err = s.CreateEmailVerification(ctx, user); err != nil {
		return gmodel.User{}, model.EmailVerification{}, err
	}

	// Commit changes from transaction
	if err := tx.Commit(); err != nil {
		return gmodel.User{}, model.EmailVerification{}, err
	}
	return user, email_verification, nil
}

func (s Service) CreateOauthUser(ctx context.Context, input gmodel.CreateAccountInput, oauth_type model.UserAuthPlatformType) (gmodel.User, error) {
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
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &user); err != nil {
		return gmodel.User{}, fmt.Errorf("user entry could not be created. %s", err.Error())
	}
	return user, nil
}

func (s Service) GoogleAuthentication(
	ctx context.Context,
	access_token string,
	ip_address *string,
	device_type *model.AuthDeviceType,
) (gmodel.Auth, error) {
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
	if s.UserEmailExists(ctx, userinfo.Email) {
		user, _ = s.FindUserByEmail(ctx, userinfo.Email)
	} else {
		user, err = s.CreateOauthUser(ctx, gmodel.CreateAccountInput{
			Email: userinfo.Email,
			Name: userinfo.Name,
		}, model.UserAuthPlatformType_Google)
		if err != nil {
			return gmodel.Auth{}, err
		}
		new_user = true
	}
	auth_state, err := s.CreateAuthStateWithJwt(ctx, user.ID, model.UserAuthPlatformType_Google, ip_address, device_type)
	if err == nil && new_user {
		auth_state.IsNewUser = &new_user
	}
	return auth_state, err
}

func (s Service) LoginInternal(
	ctx context.Context,
	email string,
	password string,
	ip_address *string,
	device_type *model.AuthDeviceType,
) (gmodel.Auth, error) {
	db := s.DbOrTxQueryable()	
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
	if !s.VerifyPasswordHash(password, *verify_user.Password) {
		return gmodel.Auth{}, fmt.Errorf("incorrect email or password")
	}

	return s.CreateAuthStateWithJwt(ctx, verify_user.ID, model.UserAuthPlatformType_Internal, ip_address, device_type)
}

func (s Service) CreateAuthStateWithJwt(
	ctx context.Context, 
	user_id int64, 
	auth_platform model.UserAuthPlatformType, 
	ip_address *string,
	device_type *model.AuthDeviceType,
) (gmodel.Auth, error) {
	auth_state, auth_state_err := s.CreateAuthState(ctx, gmodel.User{ ID: user_id }, auth_platform, ip_address, device_type)
	if auth_state_err != nil {
		return gmodel.Auth{}, fmt.Errorf("could not create auth state")
	}

	user, err := s.FindAuthUserById(ctx, user_id, auth_state.ID.String())
	if err != nil {
		return gmodel.Auth{}, fmt.Errorf("internal error")
	}

	// Generate JWT
	jwt, err := s.GenerateJWT(s.Tokens.JwtKey, &user)
	if err != nil {
		return gmodel.Auth{}, fmt.Errorf("could not generate JWT")
	}
	return gmodel.Auth{
		Token: jwt,
		User: &user,
	}, nil
}

// Given an existing user, create an auth_state row
func (s Service) CreateAuthState(
	ctx context.Context,
	user gmodel.User,
	auth_platform model.UserAuthPlatformType,
	ip_address *string,
	device_type *model.AuthDeviceType,
) (model.AuthState, error) {
	if device_type == nil {
		new_device := model.AuthDeviceType_Unknown
		device_type = &new_device
	}
	query := table.AuthState.INSERT(
		table.AuthState.ID,
		table.AuthState.UserID,
		table.AuthState.IPAddress,
		table.AuthState.Platform,
		table.AuthState.DeviceType,
	).MODEL(model.AuthState{
		ID: uuid.New(),
		UserID: user.ID,
		IPAddress: ip_address,
		Platform: auth_platform,
		DeviceType: device_type,
	}).RETURNING(table.AuthState.AllColumns)

	var auth_state model.AuthState
	err := query.QueryContext(ctx, s.DbOrTxQueryable(), &auth_state)
	return auth_state, err
}

func (s Service) CreateEmailVerification(ctx context.Context, user gmodel.User) (model.EmailVerification, error) {
	code := randstr.Dec(EMAIL_VERIFICATION_CODE_LEN)

	query := table.EmailVerification.INSERT(
		table.EmailVerification.UserID,
		table.EmailVerification.Code,
	).MODEL(model.EmailVerification{
		UserID: user.ID,
		Code: code,
	}).RETURNING(table.EmailVerification.AllColumns)

	var email_verification model.EmailVerification
	err := query.QueryContext(ctx, s.DbOrTxQueryable(), &email_verification)
	return email_verification, err
}

func (s Service) ResendEmailVerification(ctx context.Context, user gmodel.User) (email_verification model.EmailVerification, err error) {
	if user.Active {
		return model.EmailVerification{}, fmt.Errorf("user already has a verified email address")
	}
	s.TX, err = s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.EmailVerification{}, err
	}
	defer s.TX.Rollback()

	_, err = table.EmailVerification.DELETE().
		WHERE(table.EmailVerification.UserID.EQ(postgres.Int(user.ID))).
		ExecContext(ctx, s.TX)
	if err != nil {
		return model.EmailVerification{}, fmt.Errorf("user email verification entry deletion failed")
	}

	email_verification, err = s.CreateEmailVerification(ctx, user)
	if err != nil {
		return model.EmailVerification{}, err
	}
	if err := s.TX.Commit(); err != nil {
		return model.EmailVerification{}, fmt.Errorf("could not commit changes")
	}
	return email_verification, nil
}

func (s Service) FindEmailVerificationByCode(ctx context.Context, verification_code string) (model.EmailVerification, error) {
	qb := table.EmailVerification.
		SELECT(table.EmailVerification.AllColumns).
		WHERE(table.EmailVerification.Code.EQ(postgres.String(verification_code))).
		LIMIT(1)
	var email_verification model.EmailVerification
	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &email_verification); err != nil {
		return model.EmailVerification{}, fmt.Errorf("invalid email verification code")
	}
	return email_verification, nil
}

func (s Service) VerifyUserEmail(ctx context.Context, verification_code string) (gmodel.User, error) {
	var err error
	s.TX, err = s.DB.BeginTx(ctx, nil)
	if err != nil {
		return gmodel.User{}, err
	}
	defer s.TX.Rollback()

	email_verification, err := s.FindEmailVerificationByCode(ctx, verification_code)
	if err != nil {
		return gmodel.User{}, err
	}

	if time.Until(email_verification.CreatedAt).Abs() > (10 * time.Minute) {
		// Delete verification entry since it's expired
		del_query := table.EmailVerification.
			DELETE().
			WHERE(table.EmailVerification.ID.EQ(postgres.Int(email_verification.ID)))
		if _, err := del_query.ExecContext(ctx, s.DB); err != nil {
			return gmodel.User{}, err
		}
		return gmodel.User{}, fmt.Errorf("verification code has expired")
	}

	update := table.User.
		UPDATE(table.User.Active, table.User.UpdatedAt).
		SET(postgres.Bool(true), postgres.DateT(time.Now())).
		WHERE(table.User.ID.EQ(postgres.Int(email_verification.UserID)))
	if _, err := update.ExecContext(ctx, s.TX); err != nil {
		return gmodel.User{}, fmt.Errorf("could not update user email verification status to verified")
	}

	// Remove email_verification row
	delete := table.EmailVerification.
		DELETE().
		WHERE(postgres.AND(
			table.EmailVerification.ID.EQ(postgres.Int(email_verification.ID)),
			table.EmailVerification.Code.EQ(postgres.String(verification_code)),
		))
	if _, err := delete.ExecContext(ctx, s.TX); err != nil {
		return gmodel.User{}, fmt.Errorf("could not delete email verification entry")
	}

	if err := s.TX.Commit(); err != nil {
		return gmodel.User{}, fmt.Errorf("could not commit changes")
	}
	s.TX = nil
	return s.FindUserById(ctx, email_verification.UserID)
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

func (s Service) VerifyJwt(ctx context.Context, authorization types.AuthorizationKeyType) (user gmodel.User, err error) {
	jwt_raw, err := authorization.GetToken()
	if err != nil {
		return gmodel.User{}, err
	}
	if s.Tokens == nil {
		return gmodel.User{}, fmt.Errorf("tokens value is nil")
	}

	claims, err := utils.GetJwtClaims(jwt_raw, s.Tokens.JwtKey)
	if err != nil {
		return gmodel.User{}, err
	}
	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return gmodel.User{}, fmt.Errorf("token expired")
	}

	auth_state_id := claims["authStateId"].(string)
	user_id := int64(claims["id"].(float64))
	email := claims["email"].(string)
	user, err = s.FindAuthUserById(ctx, user_id, auth_state_id)
	if err != nil || email != user.Email {
		return gmodel.User{}, fmt.Errorf("one or more invalid claim values")
	}
	return user, nil
}

func (Service) GetAuthUserFromContext(ctx context.Context) gmodel.User {
	val := ctx.Value(types.AuthUserKey)
	if val == nil {
		return gmodel.User{}
	}
	return val.(gmodel.User)
}

func (s Service) UpdateUserFull(ctx context.Context, user gmodel.User, input gmodel.UpdateUserFull) (updated_user gmodel.User, err error) {
	if input.AvatarBase64 != nil && !utils.IsValidBase64Image(*input.AvatarBase64) {
		return gmodel.User{}, fmt.Errorf("invalid base64 image")
	}

	u := model.User{}
	columns := postgres.ColumnList{}
	if input.Email != nil {
		columns = append(columns, table.User.Email)
		u.Email = *input.Email
	}
	if input.PhoneNumber != nil {
		columns = append(columns, table.User.PhoneNumber)
		u.PhoneNumber = input.PhoneNumber
	}
	if input.Name != nil {
		columns = append(columns, table.User.Name)
		u.Name = *input.Name
	}
	if input.AvatarFile != nil || input.AvatarBase64 != nil {
		columns = append(columns, table.User.Avatar)
		avatar_id := uuid.NewString()
		u.Avatar = &avatar_id
	}
	if input.BirthDate != nil {
		columns = append(columns, table.User.BirthDate)
		u.BirthDate = input.BirthDate
	}
	if input.Active != nil {
		columns = append(columns, table.User.Active)
		u.Active = *input.Active
	}
	// Don't update user if their role is already "SUPER_ADMIN"
	if input.Role != nil && user.Role != gmodel.UserRoleSuperAdmin {
		columns = append(columns, table.User.Role)
		role := model.UserRoleType_Consumer
		if err := role.Scan(input.Role.String()); err != nil {
			return gmodel.User{}, err
		}
		u.Role = role
	}
	var address gmodel.Address
	if input.Address != nil {
		columns = append(columns, table.User.AddressID)
		address_input, err := s.FullAddressToCreateAddress(ctx, *input.Address)
		if err != nil {
			return gmodel.User{}, err
		}
		address, err = s.FindOrCreateAddress(ctx, &user, address_input)
		if err != nil {
			return gmodel.User{}, err
		}
		u.AddressID = &address.ID
	}
	if len(columns) == 0 {
		return user, nil
	}
	columns = append(columns, table.User.UpdatedAt)
	u.UpdatedAt = time.Now()
	qb := table.User.
		UPDATE(columns).
		MODEL(u).
		WHERE(table.User.ID.EQ(postgres.Int(user.ID))).
		RETURNING(table.User.AllColumns)
	
	if err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &updated_user); err != nil {
		return gmodel.User{}, err
	}
	if user.AuthStateID == nil {
		if input.Address != nil {
			updated_user.Address = &address
		}
		return updated_user, nil
	}
	return s.FindAuthUserById(ctx, user.ID, *user.AuthStateID)
}

func (s Service) UpdateUser(ctx context.Context, user gmodel.User, input gmodel.UpdateUser) (updated_user gmodel.User, err error) {
	return s.UpdateUserFull(ctx, user, gmodel.UpdateUserFull{
		Name: input.Name,
		AvatarFile: input.AvatarFile,
		AvatarBase64: input.AvatarBase64,
		BirthDate: input.BirthDate,
		Bio: input.Bio,
		Address: input.Address,
	})
}

func (s Service) Logout(ctx context.Context, user gmodel.User, auth_state_id string) error {
	auth_state_uuid, err := uuid.Parse(auth_state_id)
	if err != nil {
		return err
	}

	db := s.DbOrTxQueryable()
	qb := table.AuthState.
		SELECT(table.AuthState.ID.AS("id")).
		FROM(table.AuthState).
		WHERE(
			postgres.AND(
				table.AuthState.ID.EQ(postgres.UUID(auth_state_uuid)),
				table.AuthState.UserID.EQ(postgres.Int64(user.ID)),
			),
		)
	var res struct { ID uuid.UUID }
	if err := qb.QueryContext(ctx, db, &res); err != nil {
		return err
	}

	_, err = table.AuthState.
		DELETE().
		WHERE(table.AuthState.ID.EQ(
			postgres.UUID(res.ID),
		)).
		ExecContext(ctx, s.DB)
	return err
}

func (s Service) CreatedAndUpdatedUserTable() (
	created_by_user *table.UserTable, 
	updated_by_user *table.UserTable, 
	columns []postgres.Projection,
) {
	created_by_user = table.User.AS("created_by_user")
	updated_by_user = table.User.AS("updated_by_user")
	columns = []postgres.Projection{
		created_by_user.ID,
		created_by_user.Name,
		created_by_user.Avatar,
		updated_by_user.ID,
		updated_by_user.Name,
		updated_by_user.Avatar,
	}
	return created_by_user, updated_by_user, columns
}

func (Service) RoleValue(role gmodel.UserRole) int {
	switch role {
		case gmodel.UserRoleSuperAdmin: return 4
		case gmodel.UserRoleAdmin: return 3
		case gmodel.UserRoleContributor: return 2
		default: return 1
	}
}

func (s Service) IsRoleAuthorized(minimum_required_role gmodel.UserRole, user_role gmodel.UserRole) bool {
	return s.RoleValue(user_role) >= s.RoleValue(minimum_required_role)
}

func (s Service) PaginatedUsers(ctx context.Context, paginator_input gmodel.PaginatorInput, filters *gmodel.UserFilter) (result gmodel.PaginatedUsers, err error) {
	sql_table := table.User
	where_clause := postgres.Bool(true)

	if filters != nil {
		if filters.ID != nil {
			where_clause = where_clause.
				AND(table.User.ID.EQ(postgres.Int(*filters.ID)))
		}
		if filters.Email != nil {
			where_clause = where_clause.AND(table.User.Email.LIKE(
				postgres.String(fmt.Sprintf("%s%%", *filters.Email)),
			))
		}
		if filters.Name != nil {
			where_clause = where_clause.AND(table.User.Name.LIKE(
				postgres.String(fmt.Sprintf("%%%s%%", *filters.Name)),
			))
		}
		if filters.Role != nil {
			var role model.UserRoleType
			if err := role.Scan(filters.Role.String()); err != nil {
				return gmodel.PaginatedUsers{}, err
			}
			where_clause = where_clause.
				AND(table.User.Role.EQ(postgres.RawString("$role", map[string]any{
					"$role": role.String(),
				})))
		}
	}

	paginator, err := s.Paginate(ctx, paginator_input, sql_table, table.User.ID, where_clause)
	if err != nil {
		return gmodel.PaginatedUsers{
			Users: []*gmodel.User{},
			Paginator: &gmodel.Paginator{},
		}, nil
	}
	qb := table.User.
		SELECT(table.User.AllColumns).
		FROM(sql_table).
		WHERE(where_clause).
		LIMIT(int64(paginator.Limit)).
		OFFSET(int64(paginator.Offset)).
		ORDER_BY(table.User.ID.DESC())

	if err := qb.QueryContext(ctx, s.DbOrTxQueryable(), &result.Users); err != nil {
		return gmodel.PaginatedUsers{}, err
	}
	result.Paginator = &paginator.Paginator
	return result, nil
}

func (s Service) LogoutAllForUser(ctx context.Context, user_id int64) error {
	qb := table.AuthState.
		DELETE().
		WHERE(table.AuthState.UserID.EQ(postgres.Int(user_id)))
	if _, err := qb.ExecContext(ctx, s.DB); err != nil {
		return err
	}
	return nil
}

func (s Service) CreatePasswordResetEntry(
	ctx context.Context,
	email string,
) (password_reset model.PasswordReset, user gmodel.User, err error) {
	s.TX, err = s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.PasswordReset{}, gmodel.User{}, err
	}
	defer s.TX.Rollback()

	if user, err = s.FindUserByEmail(ctx, email); err != nil {
		return model.PasswordReset{}, gmodel.User{}, fmt.Errorf("invalid email")
	}
	// Delete all existing rows for user...
	qb := table.PasswordReset.
		DELETE().
		WHERE(table.PasswordReset.UserID.EQ(postgres.Int(user.ID)))
	if _, err = qb.ExecContext(ctx, s.TX); err != nil {
		return model.PasswordReset{}, gmodel.User{}, err
	}

	code := strings.ToUpper(randstr.Base62(PASSWORD_RESET_CODE_LEN))
	query := table.PasswordReset.INSERT(
		table.PasswordReset.UserID,
		table.PasswordReset.Code,
	).MODEL(model.PasswordReset{
		UserID: user.ID,
		Code: code,
	}).RETURNING(table.PasswordReset.AllColumns)
	if err = query.QueryContext(ctx, s.DbOrTxQueryable(), &password_reset); err != nil {
		return model.PasswordReset{}, gmodel.User{}, err
	}
	if err := s.TX.Commit(); err != nil {
		return model.PasswordReset{}, gmodel.User{}, fmt.Errorf("could not commit changes")
	}
	return password_reset, user, nil
}

func (s Service) ValidatePasswordResetCode(
	ctx context.Context,
	email string,
	code string,
) (password_reset model.PasswordReset, err error) {
	s.TX, err = s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.PasswordReset{}, err
	}
	defer s.TX.Rollback()

	var user gmodel.User
	if user, err = s.FindUserByEmail(ctx, email); err != nil {
		return model.PasswordReset{}, fmt.Errorf("invalid email")
	}
	qb := table.PasswordReset.
		SELECT(table.PasswordReset.AllColumns).
		FROM(table.PasswordReset).
		WHERE(
			table.PasswordReset.Code.EQ(postgres.String(code)).
				AND(table.PasswordReset.UserID.EQ(postgres.Int(user.ID))),
		).
		LIMIT(1)
	if err = qb.QueryContext(ctx, s.TX, &password_reset); err != nil {
		// Update # of tries
		update_qb := table.PasswordReset.
			UPDATE(table.PasswordReset.Tries).
			SET(table.PasswordReset.Tries.SET(table.PasswordReset.Tries.ADD(postgres.Int(1)))).
			WHERE(table.PasswordReset.UserID.EQ(postgres.Int(user.ID))).
			RETURNING(table.PasswordReset.AllColumns)
		if _, err = update_qb.ExecContext(ctx, s.TX); err != nil {
			return model.PasswordReset{}, fmt.Errorf("something went wrong during update")
		}
		if err = s.TX.Commit(); err != nil {
			return model.PasswordReset{}, fmt.Errorf("could not complete transaction")
		}
		return model.PasswordReset{}, fmt.Errorf("invalid reset code")
	}

	delete_reset_entries := func() error {
		// Delete entry if tries limit is reached
		qb := table.PasswordReset.
			DELETE().
			WHERE(table.PasswordReset.ID.EQ(
				postgres.Int(password_reset.ID),
			))
		if _, err = qb.ExecContext(ctx, s.TX); err != nil {
			return err
		}
		if err = s.TX.Commit(); err != nil {
			return fmt.Errorf("could not complete action")
		}
		return nil
	}
	if time.Since(password_reset.CreatedAt) > (30 * time.Minute) {
		// Delete entry if tries limit is reached
		if err := delete_reset_entries(); err != nil {
			return model.PasswordReset{}, fmt.Errorf("something went wrong during delete action")
		}
		return model.PasswordReset{}, fmt.Errorf("password reset code has expired")
	}
	if password_reset.Tries > PASSWORD_RESET_MAX_TRIES {
		// Delete entry if tries limit is reached
		if err := delete_reset_entries(); err != nil {
			return model.PasswordReset{}, err
		}
		return model.PasswordReset{}, fmt.Errorf("maximum number of tries reached for verification")
	}

	return password_reset, nil
}

func (s Service) ResetPassword(
	ctx context.Context,
	email string,
	code string,
	new_password string,
) (user model.User, err error) {
	password_reset, err := s.ValidatePasswordResetCode(ctx, email, code)
	if err != nil {
		return model.User{}, err
	}
	
	s.TX, err = s.DB.BeginTx(ctx, nil)
	if err != nil {
		return model.User{}, err
	}
	defer s.TX.Rollback()

	new_hashed_password, err := s.HashPassword(new_password)
	if err != nil {
		return model.User{}, err
	}
	qb := table.User.
		UPDATE(
			table.User.Password,
			table.User.UpdatedAt,
		).
		MODEL(model.User{
			Password: &new_hashed_password,
			UpdatedAt: time.Now(),
		}).
		WHERE(table.User.ID.EQ(postgres.Int(password_reset.UserID))).
		RETURNING(table.User.AllColumns)
	if err = qb.QueryContext(ctx, s.TX, &user); err != nil {
		return model.User{}, err
	}
	// Delete password_reset rows for user
	password_reset_del := table.PasswordReset.
		DELETE().
		WHERE(table.PasswordReset.ID.EQ(
			postgres.Int(password_reset.ID),
		))
	if _, err = password_reset_del.ExecContext(ctx, s.TX); err != nil {
		return model.User{}, fmt.Errorf("could not delete reset entry")
	}
	// Logout of all devices for user
	if err = s.LogoutAllForUser(ctx, user.ID); err != nil {
		return model.User{}, fmt.Errorf("could not logout for user")
	}
	if err = s.TX.Commit(); err != nil {
		return model.User{}, fmt.Errorf("could not complete transaction")
	}
	return user, nil
}

func (s Service) AddExpoPushTokenToAuthState(
	ctx context.Context,
	auth_state_id string,
	expo_push_token string,
) error {
	auth_state_uuid, err := uuid.Parse(auth_state_id)
	if err != nil {
		return err
	}

	qb := table.AuthState.
		UPDATE(table.AuthState.ExpoPushToken).
		MODEL(model.AuthState{
			ExpoPushToken: &expo_push_token,
		}).
		WHERE(table.AuthState.ID.EQ(postgres.UUID(auth_state_uuid))).
		RETURNING(table.AuthState.AllColumns)
	var dest model.AuthState
	return qb.QueryContext(ctx, s.DbOrTxQueryable(), &dest)
}
