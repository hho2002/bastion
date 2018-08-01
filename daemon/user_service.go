package daemon

import (
	"time"

	"github.com/asdine/storm"
	"github.com/jinzhu/copier"
	"github.com/yankeguo/bastion/daemon/models"
	"github.com/yankeguo/bastion/types"
	"github.com/yankeguo/bastion/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errInvalidAuthentication = status.Error(codes.InvalidArgument, "user not found or invalid password")
)

func (d *Daemon) ListUsers(c context.Context, req *types.ListUsersRequest) (res *types.ListUsersResponse, err error) {
	var users []models.User
	if err = d.DB.All(&users); err != nil {
		err = errFromStorm(err)
		return
	}
	ret := make([]*types.User, 0, len(users))
	for _, u := range users {
		ret = append(ret, u.ToGRPCUser())
	}
	res = &types.ListUsersResponse{Users: ret}
	return
}

func (d *Daemon) CreateUser(c context.Context, req *types.CreateUserRequest) (res *types.CreateUserResponse, err error) {
	// fix request
	if err = req.Validate(); err != nil {
		return
	}

	// inside a transaction
	u := models.User{}
	err = d.Transaction(true, func(db storm.Node) (err error) {
		// find existing
		if err = checkDuplicated(db, "User", "account", req.Account); err != nil {
			return
		}
		// assign values
		copier.Copy(&u, req)
		// create password
		if u.PasswordDigest, err = utils.BcryptGenerate(req.Password); err != nil {
			err = errInternal
			return
		}
		// assign created_at / updated_at and save
		u.CreatedAt = time.Now().Unix()
		u.UpdatedAt = u.CreatedAt
		if err = db.Save(&u); err != nil {
			err = errFromStorm(err)
			return
		}
		return
	})
	// return if err != nil
	if err != nil {
		return
	}
	// build response
	res = &types.CreateUserResponse{User: u.ToGRPCUser()}
	return
}

func (d *Daemon) TouchUser(c context.Context, req *types.TouchUserRequest) (res *types.TouchUserResponse, err error) {
	u := models.User{}
	// find by account
	if err = d.DB.One("Account", req.Account, &u); err != nil {
		err = errFromStorm(err)
		return
	}
	// update viewed_at
	u.ViewedAt = time.Now().Unix()
	// save
	if err = d.DB.UpdateField(&u, "ViewedAt", u.ViewedAt); err != nil {
		err = errFromStorm(err)
		return
	}
	// build response
	res = &types.TouchUserResponse{User: u.ToGRPCUser()}
	return
}

func (d *Daemon) UpdateUser(c context.Context, req *types.UpdateUserRequest) (res *types.UpdateUserResponse, err error) {
	// validate request
	if err = req.Validate(); err != nil {
		return
	}
	// find user by account
	u := models.User{}
	if err = d.DB.One("Account", req.Account, &u); err != nil {
		err = errFromStorm(err)
		return
	}
	// update user
	if req.UpdateIsBlocked {
		u.IsBlocked = req.IsBlocked
	}
	if req.UpdateIsAdmin {
		u.IsAdmin = req.IsAdmin
	}
	if req.UpdateNickname {
		u.Nickname = req.Nickname
	}
	if req.UpdatePassword {
		if u.PasswordDigest, err = utils.BcryptGenerate(req.Password); err != nil {
			err = errInternal
			return
		}
	}
	// update updated_at
	u.UpdatedAt = time.Now().Unix()
	// save
	if err = d.DB.Update(&u); err != nil {
		err = errFromStorm(err)
		return
	}
	// build response
	res = &types.UpdateUserResponse{User: u.ToGRPCUser()}
	return
}

func (d *Daemon) AuthenticateUser(c context.Context, req *types.AuthenticateUserRequest) (res *types.AuthenticateUserResponse, err error) {
	u := models.User{}
	// find by account
	if err = d.DB.One("Account", req.Account, &u); err != nil {
		// hide not found error
		if err == storm.ErrNotFound {
			err = errInvalidAuthentication
		} else {
			err = errDatabase
		}
		return
	}
	// validate password
	if !utils.BcryptValidate(u.PasswordDigest, req.Password) {
		err = errInvalidAuthentication
		return
	}
	// update viewed_at
	u.ViewedAt = time.Now().Unix()
	if err = d.DB.UpdateField(&u, "ViewedAt", u.ViewedAt); err != nil {
		err = errFromStorm(err)
		return
	}
	// build response
	res = &types.AuthenticateUserResponse{User: u.ToGRPCUser()}
	return
}