package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"math/rand"

	"github.com/samyak-jain/agora_backend/graph/generated"
	"github.com/samyak-jain/agora_backend/graph/model"
	"github.com/samyak-jain/agora_backend/middleware"
	"github.com/samyak-jain/agora_backend/models"
	"github.com/samyak-jain/agora_backend/utils"
	"github.com/satori/go.uuid"
)

func (r *mutationResolver) CreateChannel(ctx context.Context, channel string, password *model.PasswordInput, enableLink *bool) (*model.ShareResponse, error) {
	authUser := middleware.GetUserFromContext(ctx)
	if authUser == nil {
		return nil, errors.New("Invalid Token")
	}

	var hostPhrase string
	var viewPhrase string

	if *enableLink {
		hostPhrase = uuid.NewV4().String()
		viewPhrase = uuid.NewV4().String()
	}

	if !r.DB.Where("name = ?", channel).RecordNotFound() {
		return nil, errors.New("Channel name already taken")
	}

	r.DB.NewRecord(models.Channel{
		Name:             channel,
		HostPassword:     password.Host,
		ViewerPassword:   password.View,
		HostPassphrase:   hostPhrase,
		ViewerPassphrase: viewPhrase,
		Creator:          *authUser,
	})

	passwordResponse := model.Password(*password)

	return &model.ShareResponse{
		Password: &passwordResponse,
		Passphrase: &model.Passphrase{
			Host: hostPhrase,
			View: viewPhrase,
		},
	}, nil
}

func (r *mutationResolver) UpdateUserName(ctx context.Context, name string) (*model.User, error) {
	authUser := middleware.GetUserFromContext(ctx)
	if authUser == nil {
		return nil, errors.New("Invalid Token")
	}

	user := &models.User{Token: authUser.Token}
	if err := r.DB.Model(&user).Update("name", name).Error; err != nil {
		return nil, err
	}

	return &model.User{
		Name:  name,
		Email: authUser.Email,
	}, nil
}

func (r *queryResolver) JoinChannel(ctx context.Context, channel string, password string) (*model.Session, error) {
	uid := int(rand.Uint32())
	rtcToken, err := utils.GetRtcToken(channel, uid)
	if err != nil {
		return nil, err
	}

	rtmToken, err := utils.GetRtmToken(string(uid))
	if err != nil {
		return nil, err
	}

	var channelData models.Channel
	if err := r.DB.Where("name = ?", channel).First(channelData).Error; err != nil {
		return nil, err
	}

	var host bool
	if password == channelData.HostPassword {
		host = true
	} else if password == channelData.ViewerPassword {
		host = false
	} else {
		return nil, errors.New("Invalid Password")
	}

	return &model.Session{
		Channel: &channel,
		Rtc:     rtcToken,
		Rtm:     rtmToken,
		UID:     uid,
		IsHost:  host,
	}, nil
}

func (r *queryResolver) JoinChannelWithPassphrase(ctx context.Context, passphrase *model.PassphraseInput) (*model.Session, error) {
	var channelData models.Channel
	var host bool
	if r.DB.Where("host_passphrase = ?", passphrase).First(channelData).RecordNotFound() {
		if r.DB.Where("viewer_passphrase = ?", passphrase).First(channelData).RecordNotFound() {
			return nil, errors.New("Invalid passphrase")
		}

		host = false
	} else {
		host = true
	}

	uid := int(rand.Uint32())
	rtcToken, err := utils.GetRtcToken(channelData.Name, uid)
	if err != nil {
		return nil, err
	}

	rtmToken, err := utils.GetRtmToken(string(uid))
	if err != nil {
		return nil, err
	}

	return &model.Session{
		Channel: &channelData.Name,
		Rtc:     rtcToken,
		Rtm:     rtmToken,
		UID:     uid,
		IsHost:  host,
	}, nil
}

func (r *queryResolver) Share(ctx context.Context, channel string) (*model.ShareResponse, error) {
	authUser := middleware.GetUserFromContext(ctx)
	if authUser == nil {
		return nil, errors.New("Invalid Token")
	}

	var channelData models.Channel
	var userData models.User

	if err := r.DB.Where("name = ?", channel).First(&channelData).Related(&userData).Error; err != nil {
		return nil, err
	}

	if userData.Token != authUser.Token {
		return nil, errors.New("Unauthorized Access")
	}

	return &model.ShareResponse{
		Password:   &model.Password{Host: channelData.HostPassword, View: channelData.ViewerPassword},
		Passphrase: &model.Passphrase{Host: channelData.HostPassphrase, View: channelData.ViewerPassphrase},
	}, nil
}

func (r *queryResolver) GetUser(ctx context.Context) (*model.User, error) {
	authUser := middleware.GetUserFromContext(ctx)
	if authUser == nil {
		return nil, errors.New("Invalid Token")
	}

	return &model.User{
		Name:  authUser.Name,
		Email: authUser.Email,
	}, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }