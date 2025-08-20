package tests

import (
	"testing"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
	"github.com/pricetra/api/utils"
)

func TestUser(t *testing.T) {
	var user1 gmodel.User
	user1_input := gmodel.CreateAccountInput{
		Email: "user1@email.com",
		Name: "User 1",
		Password: "password123",
	}
	var (
		user1_auth gmodel.Auth
		user1_auth2 gmodel.Auth
	)
	
	t.Run("create user", func(t *testing.T) {
		var err error
		var email_verification model.EmailVerification
		user1, email_verification, err = service.CreateInternalUser(ctx, user1_input)
		if err != nil {
			t.Fatal(err)
		}
		if user1.Active {
			t.Fatal("User.Active should be set to false")
		}

		t.Parallel()

		t.Run("password must be hashed", func(t *testing.T) {
			var u model.User
			qb := table.User.SELECT(table.User.AllColumns).
				FROM(table.User).
				WHERE(table.User.ID.EQ(postgres.Int(user1.ID))).
				LIMIT(1)
			if err := qb.QueryContext(ctx, db, &u); err != nil {
				t.Fatal(err)
			}

			if *u.Password == user1_input.Password {
				t.Fatal("password should not be stored in plain text")
			}

			new_hash_pass, err := service.HashPassword(user1_input.Password)
			if err != nil {
				t.Fatal("could not hash password", err)
			}
			if *u.Password == new_hash_pass {
				t.Fatal("hashes of the same password should not result in the same hash")
			}
		})

		t.Run("should create email verification entry", func(t *testing.T) {
			ev_check, err := service.FindEmailVerificationByCode(ctx, email_verification.Code)
			if err != nil {
				t.Fatal("should find verification entry using the code")
			}
			if ev_check.ID != email_verification.ID {
				t.Fatal("verification ids must match")
			}

			t.Run("email verification helper methods", func(t *testing.T) {
				service.TX, err = db.BeginTx(ctx, nil)
				if err != nil {
					t.Fatal(err)
				}

				u, err := service.VerifyUserEmail(ctx, email_verification.Code)
				if err != nil {
					t.Fatal(err)
				}
				if u.ID != email_verification.UserID {
					t.Fatal("user id is incorrect")
				}
				if !u.Active {
					t.Fatal("user should be set to active")
				}
				if _, err := service.FindEmailVerificationByCode(ctx, email_verification.Code); err == nil {
					t.Fatal("verification entry should be deleted")
				}
				service.TX.Rollback()
				service.TX = nil
			})
		})

		t.Run("duplicate user", func(t *testing.T) {
			_, _, err := service.CreateInternalUser(ctx, gmodel.CreateAccountInput{
				Email: user1.Email,
				Password: "abc123",
			})
	
			if err == nil {
				t.Fatal("user email is duplicate. should not create user.")
			}
		})
	})

	t.Run("login", func(t *testing.T) {
		t.Run("correct input", func(t *testing.T) {
			t.Run("inactive user", func(t *testing.T) {
				_, err := service.LoginInternal(ctx, user1_input.Email, user1_input.Password, nil, nil)
				if err != nil {
					t.Fatal("should still login, even if user is inactive")
				}
			})

			t.Run("active user", func(t *testing.T) {
				var err error
				// activate user account first...
				_, err = table.User.UPDATE(table.User.Active).
					SET(postgres.Bool(true)).
					WHERE(table.User.ID.EQ(postgres.Int(user1.ID))).
					ExecContext(ctx, db)
				if err != nil {
					t.Fatal("could not update user active status", err.Error())
				}
				user1.Active = true

				user1_auth, err = service.LoginInternal(ctx, user1_input.Email, user1_input.Password, nil, nil)
				if err != nil {
					t.Fatal("could not login", err.Error())
				}
				user1.AuthStateID = user1_auth.User.AuthStateID
				user1.AuthPlatform = user1_auth.User.AuthPlatform
				if user1_auth.User.ID != user1.ID {
					t.Fatal("returned user does not match", user1_auth.User, user1)
				}

				if *user1_auth.User.AuthDevice != gmodel.AuthDeviceTypeUnknown {
					t.Fatal("default auth platform type should be 'unknown'", user1_auth.User.AuthPlatform)
				}

				t.Parallel()
				
				t.Run("verify jwt", func(t *testing.T) {
					claims, err := utils.GetJwtClaims(user1_auth.Token, app.Tokens.JwtKey)
					if err != nil {
						t.Fatal("invalid jwt", err.Error())
					}

					claims_user_id, ok := claims["id"].(float64)
					if !ok {
						t.Fatal("could not convert claims.id to float64")
					}
					if int64(claims_user_id) != user1.ID {
						t.Fatal("jwt claim user.id does not match")
					}
				})

				t.Run("check if auth_state entry exists", func(t *testing.T) {
					var auth_state model.AuthState
					qb := table.AuthState.SELECT(table.AuthState.AllColumns).
						FROM(table.AuthState).
						WHERE(table.AuthState.ID.EQ(postgres.String(*user1_auth.User.AuthStateID))).
						LIMIT(1)
					if err := qb.QueryContext(ctx, db, &auth_state); err != nil {
						t.Fatal("could not find auth_state entry", err.Error())
					}

					if auth_state.UserID != user1.ID {
						t.Fatal("auth_state.user_id does not match")
					}
				})

				t.Run("new login should create new auth_state entry", func(t *testing.T) {
					device := model.AuthDeviceType_Android
					user1_auth2, err = service.LoginInternal(ctx, user1_input.Email, user1_input.Password, nil, &device)
					if err != nil {
						t.Fatal("could not login", err.Error())
					}
					if user1_auth2.User.AuthStateID == user1_auth.User.AuthStateID {
						t.Fatal("auth_state.id should not match")
					}
					if *user1_auth2.User.AuthDevice != gmodel.AuthDeviceTypeAndroid {
						t.Fatal("auth device is not 'android'")
					}

					var logins []int
					qb := table.AuthState.SELECT(table.AuthState.ID).
						FROM(table.AuthState).
						WHERE(table.AuthState.UserID.EQ(postgres.Int(user1.ID)))
					if err := qb.QueryContext(ctx, db, &logins); err != nil {
						t.Fatal("could not query auth state for user", err.Error())
					}
					if len(logins) != 3 {
						t.Fatal("auth_state for user should be 3")
					}
				})
			})
		})

		t.Run("incorrect input", func(t *testing.T) {
			device := model.AuthDeviceType_Web
			_, err := service.LoginInternal(ctx, user1_input.Email, "somerandompassword", nil, &device)
			if err == nil {
				t.Fatal("login should fail. password is incorrect")
			}
		})
	})

	t.Run("update user", func(t *testing.T) {
		user, _, err := service.CreateInternalUser(ctx, gmodel.CreateAccountInput{
			Name: "Joe Doe",
			Email: "joe_doe@prictra.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatal(err)
		}

		t.Run("update name only", func(t *testing.T) {
			new_name := "Joe Diaz Doe"
			updated_user, err := service.UpdateUser(ctx, user, gmodel.UpdateUser{
				Name: &new_name,
			})
			if err != nil {
				t.Fatal(err)
			}
			if updated_user.Name == user.Name {
				t.Fatal("name was not updated")
			}

			find_updated_user, err := service.FindUserById(ctx, updated_user.ID)
			if err != nil {
				t.Fatal(err)
			}
			if find_updated_user.Name != updated_user.Name {
				t.Fatal("updated name and found name were different")
			}
		})

		t.Run("update avatar only", func(t *testing.T) {
			img := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAC0lEQVR42mP8//8/AwAI/wP+vQAAAABJRU5ErkJggg=="
			updated_user, err := service.UpdateUser(ctx, user, gmodel.UpdateUser{
				AvatarBase64: &img,
			})
			if err != nil {
				t.Fatal(err)
			}
			user_avatar := ""
			if user.Avatar != nil {
				user_avatar = *user.Avatar
			}
			if *updated_user.Avatar == user_avatar {
				t.Fatal("avatar was not updated")
			}
			if uuid.Validate(*updated_user.Avatar) != nil {
				t.Fatal("avatar should be a valid uuid")
			}

			find_updated_user, err := service.FindUserById(ctx, updated_user.ID)
			if err != nil {
				t.Fatal(err)
			}
			if *find_updated_user.Avatar != *updated_user.Avatar {
				t.Fatal("updated avatar and found avatar were different")
			}
		})
	})
}
