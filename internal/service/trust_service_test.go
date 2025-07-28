package service

import (
	"github.com/novel/internal/model"
	"github.com/novel/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"testing"
)

// TestUpdateTrustScoreOnNewRating 是我们的第一个测试用例
func TestUpdateTrustScoreOnNewRating(t *testing.T) {
	// --- 1. Arrange (准备阶段) ---

	// 创建一个 userRepository 的模拟实例
	mockUserRepo := new(mocks.UserRepositoryMock)

	// 创建我们要测试的 trustService 实例，并注入 mock repository
	// 注意：因为我们只测试这个方法，所以 novelRepo 可以暂时传 nil
	trustSvc := NewTrustService(mockUserRepo, nil)

	// 定义测试用的数据
	testUserID := uint(1)
	initialTrustScore := 1.0
	// 这是一个高质量评论的权重
	highQualityRatingWeight := 0.9

	// 定义我们期望的 User 对象
	expectedUser := &model.User{
		Model:      gorm.Model{ID: testUserID},
		TrustScore: initialTrustScore,
	}

	// 设置 Mock 的期望行为：
	// 当 FindByID 方法被以 testUserID 作为参数调用时，
	// 我们期望它返回我们定义的 expectedUser 和 nil 错误。
	mockUserRepo.On("FindByID", testUserID).Return(expectedUser, nil)

	// 当 Update 方法被调用时，我们期望传入的对象是 *model.User 类型，
	// 并且我们让这次调用返回 nil 错误，表示更新成功。
	mockUserRepo.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	// --- 2. Act (执行阶段) ---

	// 调用我们要测试的方法
	err := trustSvc.UpdateTrustScoreOnNewRating(testUserID, highQualityRatingWeight)

	// --- 3. Assert (断言阶段) ---

	// 使用 testify/assert 来进行断言，代码更清晰
	// 断言：我们期望整个过程没有错误发生
	assert.NoError(t, err)

	// 断言：我们期望 mockUserRepo 的所有预设期望都已经被满足了
	mockUserRepo.AssertExpectations(t)

	// 进阶断言：我们可以捕获 Update 方法被调用时传入的参数，并检查它
	// 获取被捕获的调用参数
	capturedUser := mockUserRepo.Calls[1].Arguments.Get(0).(*model.User)

	// 计算期望的最终分数
	expectedScoreChange := 0.1
	expectedFinalScore := initialTrustScore + expectedScoreChange

	// 断言：传入 Update 方法的 user 对象的 TrustScore 是否是我们期望的值
	assert.Equal(t, expectedFinalScore, capturedUser.TrustScore)
}
