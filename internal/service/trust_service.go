package service

import (
	"github.com/novel/internal/repository"
	"log"
	"time"
)

// TrustService 定义了计算用户信誉的接口 (最终版)
type TrustService interface {
	GetUserTrustScore(userID uint) (float64, error)
	UpdateTrustScoreOnNewRating(userID uint, ratingWeight float64) error
	UpdateTrustScoreOnVote(voterID, authorID uint, voteChange int) error
}

// trustService 结构体实现了 TrustService 接口 (最终版)
type trustService struct {
	userRepo  repository.UserRepository
	novelRepo repository.NovelRepository // 注入 novelRepo 以备全量计算使用
}

// NewTrustService 构造函数 (最终版)
func NewTrustService(userRepo repository.UserRepository, novelRepo repository.NovelRepository) TrustService {
	return &trustService{
		userRepo:  userRepo,
		novelRepo: novelRepo,
	}
}

// GetUserTrustScore 获取用户的信誉分
func (s *trustService) GetUserTrustScore(userID uint) (float64, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return 1.0, err
	}
	return user.TrustScore, nil
}

// --- 增量计算方法 ---

func (s *trustService) UpdateTrustScoreOnNewRating(userID uint, ratingWeight float64) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	var scoreChange float64
	if ratingWeight > 0.8 {
		scoreChange = 0.1
	} else {
		scoreChange = 0.02
	}
	user.TrustScore = s.applyLimits(user.TrustScore + scoreChange)
	return s.userRepo.Update(user)
}

func (s *trustService) UpdateTrustScoreOnVote(voterID, authorID uint, voteChange int) error {
	author, err := s.userRepo.FindByID(authorID)
	if err != nil {
		return err
	}
	authorScoreChange := float64(voteChange) * 0.01
	author.TrustScore = s.applyLimits(author.TrustScore + authorScoreChange)
	if err := s.userRepo.Update(author); err != nil {
		return err
	}
	if voterID == authorID {
		return nil
	}
	voter, err := s.userRepo.FindByID(voterID)
	if err != nil {
		return err
	}
	voterScoreChange := 0.001
	voter.TrustScore = s.applyLimits(voter.TrustScore + voterScoreChange)
	return s.userRepo.Update(voter)
}

// RecalculateAndSaveUserTrustScore 全量重新计算一个用户的信誉分 (作为内部工具)
// 注意：这个方法不再是接口的一部分，但实现被保留了下来，供内部定时任务调用
func (s *trustService) RecalculateAndSaveUserTrustScore(userID uint) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("FATAL: Recovered in RecalculateUserTrustScore for userID %d. Panic: %v", userID, r)
		}
	}()

	user, err := s.userRepo.FindByIDWithRatings(userID)
	if err != nil {
		return err
	}

	var score float64 = 1.0
	daysSinceRegistration := time.Since(user.CreatedAt).Hours() / 24
	score += (daysSinceRegistration / 30) * 0.05
	var totalUpvotes, highQualityComments int
	for _, rating := range user.Ratings {
		totalUpvotes += rating.UpvotesCount
		if rating.Weight > 0.8 {
			highQualityComments++
		}
	}
	score += float64(highQualityComments) * 0.1
	score += float64(totalUpvotes) * 0.01

	user.TrustScore = s.applyLimits(score)
	return s.userRepo.Update(user)
}

// applyLimits 辅助函数
func (s *trustService) applyLimits(score float64) float64 {
	if score > 1.5 {
		return 1.5
	}
	if score < 0.8 {
		return 0.8
	}
	return score
}
