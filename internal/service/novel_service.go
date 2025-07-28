package service

import (
	"errors"
	"github.com/novel/internal/dto"
	"github.com/novel/internal/model"
	"github.com/novel/internal/pkg/config"
	"github.com/novel/internal/repository"
	"gorm.io/gorm"
	"log"
	"math"
)

// NovelService 定义了与小说相关的业务逻辑接口
type NovelService interface {
	GetRankedNovels(query *dto.ListQuery) (*dto.PaginatedResponse, error)
	GetNovelWithCalculatedScores(id uint) (*NovelScoreDetails, error)
	CreateRatingForNovel(userID, novelID uint, score int, comment string) (*model.Rating, error)
	VoteForRating(userID, ratingID uint, voteType model.VoteType) error
	CreateNovel(req *dto.CreateNovelRequest) (*model.Novel, error)
}

// NovelScoreDetails 是一个新的 DTO，用于封装小说及其各种计算分数
type NovelScoreDetails struct {
	Novel         *model.Novel `json:"novel"`
	RatingsCount  int          `json:"ratings_count"`
	WeightedScore float64      `json:"weighted_score"`
}

// novelService 结构体实现了 NovelService 接口
type novelService struct {
	repo         repository.NovelRepository
	trustSvc     TrustService
	m            float64 // IMDb 公式参数 m
	c            float64 // IMDb 公式参数 c
	categoryRepo repository.CategoryRepository
	tagRepo      repository.TagRepository
}

// NewNovelService 是 novelService 的构造函数，负责所有依赖的注入
func NewNovelService(repo repository.NovelRepository, trustSvc TrustService, categoryRepo repository.CategoryRepository, tagRepo repository.TagRepository, cfg *config.AlgorithmConfig) NovelService {
	return &novelService{
		repo:         repo,
		trustSvc:     trustSvc,
		m:            cfg.ImdbM,
		c:            cfg.ImdbC,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
	}

}

// GetRankedNovels 实现了获取排序和分页后的小说列表的业务逻辑
func (s *novelService) GetRankedNovels(query *dto.ListQuery) (*dto.PaginatedResponse, error) {
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	// 直接将 query 传递给 Repository
	novels, total, err := s.repo.FindAll(query)
	if err != nil {
		return nil, err
	}
	return &dto.PaginatedResponse{
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
		Data:     novels,
	}, nil
}

// GetNovelWithCalculatedScores 直接从数据库读取预计算的分数，实现高性能读取
func (s *novelService) GetNovelWithCalculatedScores(id uint) (*NovelScoreDetails, error) {
	novel, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return &NovelScoreDetails{
		Novel:         novel,
		RatingsCount:  novel.RatingsCount,
		WeightedScore: novel.WeightedScore,
	}, nil
}

// CreateRatingForNovel 封装了创建评分的业务逻辑
func (s *novelService) CreateRatingForNovel(userID, novelID uint, score int, comment string) (*model.Rating, error) {
	_, err := s.repo.FindByID(novelID)
	if err != nil {
		return nil, err
	}
	// [TODO] 在这里可以增加业务规则，例如：一个用户对一本书只能评分一次
	rating := &model.Rating{
		UserID:  userID,
		NovelID: novelID,
		Score:   score,
		Comment: comment,
	}
	if err := s.repo.CreateRating(rating); err != nil {
		return nil, errors.New("failed to create rating in repository")
	}
	go s.triggerCalculationsOnNewRating(rating)
	return rating, nil
}

// VoteForRating 实现了完整的投票业务逻辑
func (s *novelService) VoteForRating(userID, ratingID uint, voteType model.VoteType) error {
	rating, err := s.repo.FindRatingByID(ratingID)
	if err != nil {
		return err
	}
	oldVote, err := s.repo.FindUserVote(userID, ratingID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("failed to check existing vote")
	}

	var newVote *model.RatingVote
	isCancelVote := false
	var voteChange int

	if errors.Is(err, gorm.ErrRecordNotFound) { // 首次投票
		newVote = &model.RatingVote{UserID: userID, RatingID: ratingID, Vote: voteType}
		voteChange = int(voteType) // 赞同为+1，反对为-1
		if voteType == model.VoteTypeUp {
			rating.UpvotesCount++
		} else {
			rating.DownvotesCount++
		}
	} else { // 已投过票
		if oldVote.Vote == voteType { // 取消投票
			isCancelVote = true
			voteChange = -int(voteType) // 取消赞同为-1，取消反对为+1
			if oldVote.Vote == model.VoteTypeUp {
				rating.UpvotesCount--
			} else {
				rating.DownvotesCount--
			}
		} else { // 改票
			newVote = &model.RatingVote{UserID: userID, RatingID: ratingID, Vote: voteType}
			voteChange = 2 * int(voteType) // 赞->踩为-2, 踩->赞为+2
			if voteType == model.VoteTypeUp {
				rating.UpvotesCount++
				rating.DownvotesCount--
			} else {
				rating.DownvotesCount++
				rating.UpvotesCount--
			}
		}
	}
	if isCancelVote {
		err = s.repo.UpdateRatingVote(rating, oldVote, nil)
	} else {
		err = s.repo.UpdateRatingVote(rating, oldVote, newVote)
	}
	if err != nil {
		return errors.New("failed to update vote")
	}
	go s.triggerCalculationsOnVote(userID, rating, voteChange)
	return nil
}

// --- 后台异步计算任务 ---

func (s *novelService) triggerCalculationsOnNewRating(rating *model.Rating) {
	initialWeight := s.calculateAndSaveSingleRatingWeight(rating)
	s.trustSvc.UpdateTrustScoreOnNewRating(rating.UserID, initialWeight)
	s.recalculateAndUpdateNovelScores(rating.NovelID)
}

func (s *novelService) triggerCalculationsOnVote(voterID uint, rating *model.Rating, voteChange int) {
	s.calculateAndSaveSingleRatingWeight(rating)
	s.trustSvc.UpdateTrustScoreOnVote(voterID, rating.UserID, voteChange)
	s.recalculateAndUpdateNovelScores(rating.NovelID)
}

// --- 核心算法与辅助函数 ---

func (s *novelService) calculateAndSaveSingleRatingWeight(rating *model.Rating) float64 {
	wAction := s.calculateActionWeight(rating)
	wQuality := s.calculateQualityWeight(rating)
	wUser, _ := s.trustSvc.GetUserTrustScore(rating.UserID) // 在后台任务中，我们可以忽略错误，使用默认值
	wCommunity := s.calculateCommunityWeight(rating)

	finalWeight := wAction * wQuality * wUser * wCommunity
	rating.Weight = finalWeight
	if err := s.repo.UpdateRating(rating); err != nil {
		log.Printf("Failed to update rating weight for rating %d: %v", rating.ID, err)
	}
	return finalWeight
}

func (s *novelService) recalculateAndUpdateNovelScores(novelID uint) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("FATAL: Recovered in recalculateAndUpdateNovelScores for novelID %d. Panic: %v", novelID, r)
		}
	}()
	novel, err := s.repo.FindByIDWithRatings(novelID)
	if err != nil {
		log.Printf("ERROR: Failed to find novel %d for score recalculation: %v", novelID, err)
		return
	}
	var totalWeightedScore, totalWeight float64
	for _, rating := range novel.Ratings {
		totalWeightedScore += float64(rating.Score) * rating.Weight
		totalWeight += rating.Weight
	}
	var weightedAvgScore float64
	if totalWeight > 0 {
		weightedAvgScore = totalWeightedScore / totalWeight
	}
	v_w, R_w, m, c := totalWeight, weightedAvgScore, s.m, s.c
	finalWeightedScore := (v_w/(v_w+m))*R_w + (m/(v_w+m))*c
	novel.WeightedScore = finalWeightedScore
	novel.RatingsCount = len(novel.Ratings)
	if err := s.repo.Update(novel); err != nil {
		log.Printf("ERROR: Failed to update scores for novel %d: %v", novelID, err)
	}
	log.Printf("INFO: Successfully recalculated scores for novel %d. New WeightedScore: %.2f, RatingsCount: %d", novelID, finalWeightedScore, novel.RatingsCount)
}

func (s *novelService) calculateActionWeight(rating *model.Rating) float64 {
	if rating.Comment != "" {
		return 1.0
	}
	return 0.5
}

func (s *novelService) calculateQualityWeight(rating *model.Rating) float64 {
	// TODO 未来考虑集成AI进行评论质量的权重计算
	return 1.0
}

func (s *novelService) calculateCommunityWeight(rating *model.Rating) float64 {
	netUpvotes := rating.UpvotesCount - rating.DownvotesCount
	if netUpvotes < 0 {
		netUpvotes = 0
	}
	return 1 + 0.5*math.Log10(float64(netUpvotes)+1)
}

func (s *novelService) CreateNovel(req *dto.CreateNovelRequest) (*model.Novel, error) {
	category, err := s.categoryRepo.FindOrCreate(req.CategoryName)
	if err != nil {
		return nil, errors.New("failed to process category")
	}

	tags, err := s.tagRepo.FindOrCreateByNames(req.TagNames)
	if err != nil {
		return nil, errors.New("failed to process tags")
	}

	// --- 核心修复点：安全地处理指针类型的 DTO 字段 ---
	novel := &model.Novel{
		Title:               req.Title,
		Author:              req.Author,
		Description:         req.Description,
		CoverImageURL:       req.CoverImageURL,
		PublicationType:     model.PublicationType(req.PublicationType),
		SerializationStatus: model.SerializationStatus(req.SerializationStatus),
		CategoryID:          category.ID,
		Tags:                tags,
	}

	// 安全地解引用指针，如果指针不为 nil，则赋值
	if req.WordCount != nil {
		novel.WordCount = *req.WordCount
	}
	if req.Publisher != nil {
		novel.Publisher = req.Publisher
	}
	if req.Isbn != nil {
		novel.Isbn = req.Isbn
	}
	if req.PublicationSite != nil {
		novel.PublicationSite = *req.PublicationSite
	}
	// --- 修复结束 ---

	if err := s.repo.CreateInTx(novel); err != nil {
		return nil, errors.New("failed to create novel in transaction")
	}

	return novel, nil
}
