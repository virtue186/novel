package service

import (
	"errors"
	"github.com/novel/internal/dto"
	"github.com/novel/internal/model"
	"github.com/novel/internal/pkg/config"
	"github.com/novel/internal/repository"
	"log"
)

// NovelService 定义了与小说相关的业务逻辑接口
type NovelService interface {
	GetNovelWithCalculatedScores(id uint) (*NovelScoreDetails, error)
	CreateRatingForNovel(novelID uint, score int) (*model.Rating, error)
	GetRankedNovels(query *dto.PaginationQuery) (*dto.PaginatedResponse, error)
}

// novelService 结构体实现了 NovelService 接口
type novelService struct {
	repo repository.NovelRepository
	m    float64 // IMDb 公式参数 m
	c    float64 // IMDb 公式参数 c
}

// NovelScoreDetails 是一个新的 DTO，用于封装小说及其各种计算分数
type NovelScoreDetails struct {
	Novel          *model.Novel
	RatingsCount   int
	SimpleAvgScore float64
	WeightedScore  float64 // 我们最终的目标
}

// NewNovelService 是 novelService 的构造函数
func NewNovelService(repo repository.NovelRepository, cfg *config.AlgorithmConfig) NovelService {
	return &novelService{
		repo: repo,
		m:    cfg.ImdbM,
		c:    cfg.ImdbC,
	}
}

// GetNovelWithCalculatedScores 现在直接从数据库读取预计算的分数
func (s *novelService) GetNovelWithCalculatedScores(id uint) (*NovelScoreDetails, error) {
	novel, err := s.repo.FindByID(id) // 不再需要预加载 Ratings
	if err != nil {
		return nil, err
	}

	// 将数据库中的数据填充到 DTO 中
	details := &NovelScoreDetails{
		// 注意：这里的 Novel 对象不包含详细的 Ratings 列表，这会使API响应更轻快
		// 如果需要，可以另开一个API来获取评论列表
		Novel:          novel,
		RatingsCount:   novel.RatingsCount,
		SimpleAvgScore: 0, // 简单平均分可以考虑不再计算和返回，或按需计算
		WeightedScore:  novel.WeightedScore,
	}

	return details, nil
}

// CreateRatingForNovel 在创建评分后，异步触发分数更新
func (s *novelService) CreateRatingForNovel(novelID uint, score int) (*model.Rating, error) {
	// ... (业务校验和创建 rating 的逻辑保持不变) ...
	_, err := s.repo.FindByID(novelID)
	if err != nil {
		return nil, err
	}

	rating := &model.Rating{
		NovelID: novelID,
		Score:   score,
	}

	if err := s.repo.CreateRating(rating); err != nil {
		return nil, errors.New("failed to create rating in repository")
	}

	// 异步触发分数重新计算，不会阻塞当前请求
	go s.recalculateAndUpdateNovelScores(novelID)

	return rating, nil
}

// recalculateAndUpdateNovelScores 是核心的私有方法
func (s *novelService) recalculateAndUpdateNovelScores(novelID uint) {
	// 在一个新的 goroutine 中，我们最好用 recover 来防止 panic 导致整个程序崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in recalculateAndUpdateNovelScores: %v", r)
		}
	}()

	novel, err := s.repo.FindByIDWithRatings(novelID)
	if err != nil {
		log.Printf("Error finding novel %d for recalculation: %v", novelID, err)
		return
	}

	// --- 这里是我们的核心算法逻辑 ---
	var totalScore int
	for _, rating := range novel.Ratings {
		totalScore += rating.Score
	}
	ratingsCount := len(novel.Ratings)
	var simpleAvgScore float64
	if ratingsCount > 0 {
		simpleAvgScore = float64(totalScore) / float64(ratingsCount)
	}
	v := float64(ratingsCount)
	R := simpleAvgScore
	m := s.m
	c := s.c
	weightedScore := (v/(v+m))*R + (m/(v+m))*c
	// --- 算法逻辑结束 ---

	// 更新 novel 对象中的预计算字段
	novel.WeightedScore = weightedScore
	novel.RatingsCount = ratingsCount

	// 通过 Repository 将更新持久化到数据库
	if err := s.repo.Update(novel); err != nil {
		log.Printf("Error updating novel %d scores: %v", novelID, err)
	}

	log.Printf("Successfully recalculated scores for novel %d. New score: %.2f", novelID, weightedScore)
}

// GetRankedNovels 实现了获取排序和分页后的小说列表的业务逻辑
func (s *novelService) GetRankedNovels(query *dto.PaginationQuery) (*dto.PaginatedResponse, error) {
	// 可以在这里增加业务逻辑，例如，限制 PageSize 的最大值
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 调用 Repository 获取数据
	novels, total, err := s.repo.FindAll(query)
	if err != nil {
		return nil, err
	}

	// 封装成标准的分页响应 DTO
	return &dto.PaginatedResponse{
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
		Data:     novels,
	}, nil
}
