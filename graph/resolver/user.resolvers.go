package gresolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/graph/gmodel"
)

// CreateAccount is the resolver for the createAccount field.
func (r *mutationResolver) CreateAccount(ctx context.Context, input gmodel.CreateAccountInput) (*gmodel.User, error) {
	validation_err := r.AppContext.StructValidator.StructCtx(ctx, input)
	if validation_err != nil {
		return nil, validation_err
	}

	new_user, email_verification, err := r.Service.CreateInternalUser(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("could not create user. %s", err.Error())
	}

	email_res, err := r.Service.SendEmailVerification(ctx, new_user, email_verification)
	if err != nil {
		return nil, err
	}
	if email_res.StatusCode == http.StatusBadRequest {
		return nil, fmt.Errorf("could not send email. %s", email_res.Body)
	}
	return &new_user, nil
}

// VerifyEmail is the resolver for the verifyEmail field.
func (r *mutationResolver) VerifyEmail(ctx context.Context, verificationCode string) (*gmodel.User, error) {
	user, err := r.Service.VerifyUserEmail(ctx, verificationCode)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ResendEmailVerificationCode is the resolver for the resendEmailVerificationCode field.
func (r *mutationResolver) ResendEmailVerificationCode(ctx context.Context, email string) (bool, error) {
	user, err := r.Service.FindUserByEmail(ctx, email)
	if err != nil {
		return false, fmt.Errorf("no user found with the provided email address")
	}
	email_verification, err := r.Service.ResendEmailVerification(ctx, user)
	if err != nil {
		return false, err
	}

	email_res, email_err := r.Service.SendEmailVerification(ctx, user, email_verification)
	if email_err != nil {
		return false, email_err
	}
	if email_res.StatusCode == http.StatusBadRequest {
		return false, fmt.Errorf("could not send email. %s", email_res.Body)
	}
	log.Printf("%+v\n", email_res.Body)
	return email_verification.ID > 0, nil
}

// UpdateProfile is the resolver for the updateProfile field.
func (r *mutationResolver) UpdateProfile(ctx context.Context, input gmodel.UpdateUser) (*gmodel.User, error) {
	user := r.Service.GetAuthUserFromContext(ctx)
	updated_user, err := r.Service.UpdateUser(ctx, user, input)
	if err != nil {
		return nil, err
	}

	if updated_user.Avatar != nil && input.AvatarFile != nil {
		_, err := r.Service.GraphImageUpload(ctx, *input.AvatarFile, uploader.UploadParams{
			PublicID: *updated_user.Avatar,
			Tags:     []string{"USER_PROFILE"},
		})
		if err != nil {
			return nil, fmt.Errorf("could not upload avatar to CDN")
		}
	}
	// Delete old avatar
	if user.Avatar != nil {
		r.Service.DeleteImageUpload(ctx, *user.Avatar)
	}
	return &updated_user, nil
}

// Logout is the resolver for the logout field.
func (r *mutationResolver) Logout(ctx context.Context) (bool, error) {
	user := r.Service.GetAuthUserFromContext(ctx)
	if user.AuthStateID == nil {
		return false, fmt.Errorf("no auth state attached to user")
	}
	err := r.Service.Logout(ctx, user, *user.AuthStateID)
	return err == nil, err
}

// Login is the resolver for the login field.
func (r *queryResolver) Login(ctx context.Context, email string, password string, ipAddress *string, device *gmodel.AuthDeviceType) (*gmodel.Auth, error) {
	var mapped_device model.AuthDeviceType
	if device == nil {
		mapped_device = model.AuthDeviceType_Unknown
	} else {
		mapped_device.Scan(device.String())
	}
	auth, err := r.Service.LoginInternal(ctx, email, password, ipAddress, &mapped_device)
	return &auth, err
}

// GoogleOAuth is the resolver for the googleOAuth field.
func (r *queryResolver) GoogleOAuth(ctx context.Context, accessToken string, ipAddress *string, device *gmodel.AuthDeviceType) (*gmodel.Auth, error) {
	var mapped_device model.AuthDeviceType
	if device == nil {
		mapped_device = model.AuthDeviceType_Unknown
	} else {
		mapped_device.Scan(device.String())
	}
	auth, err := r.Service.GoogleAuthentication(ctx, accessToken, ipAddress, &mapped_device)
	return &auth, err
}

// Me is the resolver for the me field.
func (r *queryResolver) Me(ctx context.Context) (*gmodel.User, error) {
	auth_user := r.Service.GetAuthUserFromContext(ctx)
	return &auth_user, nil
}
