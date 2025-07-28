package mocks

import (
	"github.com/novel/internal/model"
	"github.com/novel/internal/repository"
	"github.com/stretchr/testify/mock"
)

// UserRepositoryMock 是一个 UserRepository 的模拟实现
type UserRepositoryMock struct {
	mock.Mock
}

// 确保 UserRepositoryMock 实现了 UserRepository 接口
var _ repository.UserRepository = (*UserRepositoryMock)(nil)

// --- 为接口中的每一个方法，都创建一个对应的模拟方法 ---

func (m *UserRepositoryMock) Create(user *model.User) error {
	// m.Called 会记录这次调用，并返回我们预设的结果
	args := m.Called(user)
	return args.Error(0)
}

func (m *UserRepositoryMock) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *UserRepositoryMock) FindByID(id uint) (*model.User, error) {
	args := m.Called(id)
	// 如果第一个返回值不是nil，则进行类型断言
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *UserRepositoryMock) FindByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *UserRepositoryMock) FindByIDWithRatings(id uint) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
