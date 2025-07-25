package integration

import (
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"red-feed/interactive/domain"
	"red-feed/interactive/integration/startup"
	"red-feed/interactive/repository/dao"
	"testing"
	"time"
)

type InteractiveTestSuite struct {
	suite.Suite
	db  *gorm.DB
	rdb redis.Cmdable
}

func (s *InteractiveTestSuite) SetupSuite() {
	s.db = startup.InitTestDB()
	s.rdb = startup.InitRedis()
}

func (s *InteractiveTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err := s.db.Exec("TRUNCATE TABLE `interactives`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `user_like_bizs`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `user_collection_bizs`").Error
	assert.NoError(s.T(), err)
	// 清空 Redis
	err = s.rdb.FlushDB(ctx).Err()
	assert.NoError(s.T(), err)
}

func (s *InteractiveTestSuite) TestIncrReadCnt() {
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64

		wantErr error
	}{
		{
			// DB 和缓存都有数据
			name: "增加成功,db和redis",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				err := s.db.Create(dao.Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      2,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					Ctime:      6,
					Utime:      7,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:2",
					"read_cnt", 3).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 1).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.Utime > 7)
				data.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:    1,
					Biz:   "test",
					BizId: 2,
					// +1 之后
					ReadCnt:    4,
					CollectCnt: 4,
					LikeCnt:    5,
					Ctime:      6,
				}, data)
				cnt, err := s.rdb.HGet(ctx, "interactive:test:2", "read_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 4, cnt)
				err = s.rdb.Del(ctx, "interactive:test:2").Err()
				assert.NoError(t, err)
			},
			biz:   "test",
			bizId: 2,
		},
		{
			// DB 有数据，缓存没有数据
			name: "增加成功,db有",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         3,
					Biz:        "test",
					BizId:      3,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					Ctime:      6,
					Utime:      7,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 3).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.Utime > 7)
				data.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:    3,
					Biz:   "test",
					BizId: 3,
					// +1 之后
					ReadCnt:    4,
					CollectCnt: 4,
					LikeCnt:    5,
					Ctime:      6,
				}, data)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:3").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},
			biz:   "test",
			bizId: 3,
		},
		{
			name:   "增加成功-都没有",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("biz = ? AND biz_id = ?", "test", 4).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.Utime > 0)
				assert.True(t, data.Ctime > 0)
				assert.True(t, data.Id > 0)
				data.Id = 0
				data.Utime = 0
				data.Ctime = 0
				assert.Equal(t, dao.Interactive{
					Biz:     "test",
					BizId:   4,
					ReadCnt: 1,
				}, data)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:4").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},
			biz:   "test",
			bizId: 4,
		},
	}

	// 不同于 AsyncSms 服务，我们不需要 mock，所以创建一个就可以
	// 不需要每个测试都创建
	svc := startup.InitInteractiveService()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := svc.IncrReadCnt(context.Background(), tc.biz, tc.bizId)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func (s *InteractiveTestSuite) TestLike() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64
		uid   int64

		wantErr error
	}{
		{
			name: "点赞-DB和cache都有",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				err := s.db.Create(dao.Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      2,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					Ctime:      6,
					Utime:      7,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:2",
					"like_cnt", 3).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 1).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.Utime > 7)
				data.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      2,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    6,
					Ctime:      6,
				}, data)

				var likeBiz dao.UserLikeBiz
				err = s.db.Where("biz = ? AND biz_id = ? AND uid = ?",
					"test", 2, 123).First(&likeBiz).Error
				assert.NoError(t, err)
				assert.True(t, likeBiz.Id > 0)
				assert.True(t, likeBiz.Ctime > 0)
				assert.True(t, likeBiz.Utime > 0)
				likeBiz.Id = 0
				likeBiz.Ctime = 0
				likeBiz.Utime = 0
				assert.Equal(t, dao.UserLikeBiz{
					Biz:    "test",
					BizId:  2,
					Uid:    123,
					Status: 1,
				}, likeBiz)

				cnt, err := s.rdb.HGet(ctx, "interactive:test:2", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 4, cnt)
				err = s.rdb.Del(ctx, "interactive:test:2").Err()
				assert.NoError(t, err)
			},
			biz:   "test",
			bizId: 2,
			uid:   123,
		},
		{
			name:   "点赞-都没有",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("biz = ? AND biz_id = ?", "test", 3).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.Utime > 0)
				assert.True(t, data.Ctime > 0)
				assert.True(t, data.Id > 0)
				data.Utime = 0
				data.Ctime = 0
				data.Id = 0
				assert.Equal(t, dao.Interactive{
					Biz:     "test",
					BizId:   3,
					LikeCnt: 1,
				}, data)

				var likeBiz dao.UserLikeBiz
				err = s.db.Where("biz = ? AND biz_id = ? AND uid = ?",
					"test", 3, 123).First(&likeBiz).Error
				assert.NoError(t, err)
				assert.True(t, likeBiz.Id > 0)
				assert.True(t, likeBiz.Ctime > 0)
				assert.True(t, likeBiz.Utime > 0)
				likeBiz.Id = 0
				likeBiz.Ctime = 0
				likeBiz.Utime = 0
				assert.Equal(t, dao.UserLikeBiz{
					Biz:    "test",
					BizId:  3,
					Uid:    123,
					Status: 1,
				}, likeBiz)

				cnt, err := s.rdb.Exists(ctx, "interactive:test:2").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},
			biz:   "test",
			bizId: 3,
			uid:   123,
		},
	}

	svc := startup.InitInteractiveService()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := svc.Like(context.Background(), tc.biz, tc.bizId, tc.uid)
			assert.NoError(t, err)
			tc.after(t)
		})
	}
}

func (s *InteractiveTestSuite) TestDislike() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64
		uid   int64

		wantErr error
	}{
		{
			name: "取消点赞-DB和cache都有",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				err := s.db.Create(dao.Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      2,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    5,
					Ctime:      6,
					Utime:      7,
				}).Error
				assert.NoError(t, err)
				err = s.db.Create(dao.UserLikeBiz{
					Id:     1,
					Biz:    "test",
					BizId:  2,
					Uid:    123,
					Ctime:  6,
					Utime:  7,
					Status: 1,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:2",
					"like_cnt", 3).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var data dao.Interactive
				err := s.db.Where("id = ?", 1).First(&data).Error
				assert.NoError(t, err)
				assert.True(t, data.Utime > 7)
				data.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      2,
					ReadCnt:    3,
					CollectCnt: 4,
					LikeCnt:    4,
					Ctime:      6,
				}, data)

				var likeBiz dao.UserLikeBiz
				err = s.db.Where("id = ?", 1).First(&likeBiz).Error
				assert.NoError(t, err)
				assert.True(t, likeBiz.Utime > 7)
				likeBiz.Utime = 0
				assert.Equal(t, dao.UserLikeBiz{
					Id:     1,
					Biz:    "test",
					BizId:  2,
					Uid:    123,
					Ctime:  6,
					Status: 0,
				}, likeBiz)

				cnt, err := s.rdb.HGet(ctx, "interactive:test:2", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 2, cnt)
				err = s.rdb.Del(ctx, "interactive:test:2").Err()
				assert.NoError(t, err)
			},
			biz:   "test",
			bizId: 2,
			uid:   123,
		},
	}

	svc := startup.InitInteractiveService()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := svc.CancelLike(context.Background(), tc.biz, tc.bizId, tc.uid)
			assert.NoError(t, err)
			tc.after(t)
		})
	}
}

func (s *InteractiveTestSuite) TestCollect() {
	testCases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)

		bizId int64
		biz   string
		cid   int64
		uid   int64

		wantErr error
	}{
		{
			name:   "收藏成功,db和缓存都没有",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var intr dao.Interactive
				err := s.db.Where("biz = ? AND biz_id = ?", "test", 1).First(&intr).Error
				assert.NoError(t, err)
				assert.True(t, intr.Ctime > 0)
				intr.Ctime = 0
				assert.True(t, intr.Utime > 0)
				intr.Utime = 0
				assert.True(t, intr.Id > 0)
				intr.Id = 0
				assert.Equal(t, dao.Interactive{
					Biz:        "test",
					BizId:      1,
					CollectCnt: 1,
				}, intr)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:1").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
				// 收藏记录
				var cbiz dao.UserCollectionBiz
				err = s.db.WithContext(ctx).
					Where("uid = ? AND biz = ? AND biz_id = ?", 1, "test", 1).
					First(&cbiz).Error
				assert.NoError(t, err)
				assert.True(t, cbiz.Ctime > 0)
				cbiz.Ctime = 0
				assert.True(t, cbiz.Utime > 0)
				cbiz.Utime = 0
				assert.True(t, cbiz.Id > 0)
				cbiz.Id = 0
				assert.Equal(t, dao.UserCollectionBiz{
					Biz:   "test",
					BizId: 1,
					Cid:   1,
					Uid:   1,
				}, cbiz)
			},
			bizId: 1,
			biz:   "test",
			cid:   1,
			uid:   1,
		},
		{
			name: "收藏成功,db有缓存没有",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := s.db.WithContext(ctx).Create(&dao.Interactive{
					Biz:        "test",
					BizId:      2,
					CollectCnt: 10,
					Ctime:      123,
					Utime:      234,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var intr dao.Interactive
				err := s.db.WithContext(ctx).
					Where("biz = ? AND biz_id = ?", "test", 2).First(&intr).Error
				assert.NoError(t, err)
				assert.True(t, intr.Ctime > 0)
				intr.Ctime = 0
				assert.True(t, intr.Utime > 0)
				intr.Utime = 0
				assert.True(t, intr.Id > 0)
				intr.Id = 0
				assert.Equal(t, dao.Interactive{
					Biz:        "test",
					BizId:      2,
					CollectCnt: 11,
				}, intr)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:2").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)

				var cbiz dao.UserCollectionBiz
				err = s.db.WithContext(ctx).
					Where("uid = ? AND biz = ? AND biz_id = ?", 1, "test", 2).
					First(&cbiz).Error
				assert.NoError(t, err)
				assert.True(t, cbiz.Ctime > 0)
				cbiz.Ctime = 0
				assert.True(t, cbiz.Utime > 0)
				cbiz.Utime = 0
				assert.True(t, cbiz.Id > 0)
				cbiz.Id = 0
				assert.Equal(t, dao.UserCollectionBiz{
					Biz:   "test",
					BizId: 2,
					Cid:   1,
					Uid:   1,
				}, cbiz)
			},
			bizId: 2,
			biz:   "test",
			cid:   1,
			uid:   1,
		},
		{
			name: "收藏成功,db和缓存都有",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := s.db.WithContext(ctx).Create(&dao.Interactive{
					Biz:        "test",
					BizId:      3,
					CollectCnt: 10,
					Ctime:      123,
					Utime:      234,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:3", "collect_cnt", 10).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var intr dao.Interactive
				err := s.db.WithContext(ctx).
					Where("biz = ? AND biz_id = ?", "test", 3).First(&intr).Error
				assert.NoError(t, err)
				assert.True(t, intr.Ctime > 0)
				intr.Ctime = 0
				assert.True(t, intr.Utime > 0)
				intr.Utime = 0
				assert.True(t, intr.Id > 0)
				intr.Id = 0
				assert.Equal(t, dao.Interactive{
					Biz:        "test",
					BizId:      3,
					CollectCnt: 11,
				}, intr)
				cnt, err := s.rdb.HGet(ctx, "interactive:test:3", "collect_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 11, cnt)

				var cbiz dao.UserCollectionBiz
				err = s.db.WithContext(ctx).
					Where("uid = ? AND biz = ? AND biz_id = ?", 1, "test", 3).
					First(&cbiz).Error
				assert.NoError(t, err)
				assert.True(t, cbiz.Ctime > 0)
				cbiz.Ctime = 0
				assert.True(t, cbiz.Utime > 0)
				cbiz.Utime = 0
				assert.True(t, cbiz.Id > 0)
				cbiz.Id = 0
				assert.Equal(t, dao.UserCollectionBiz{
					Biz:   "test",
					BizId: 3,
					Cid:   1,
					Uid:   1,
				}, cbiz)
			},
			bizId: 3,
			biz:   "test",
			cid:   1,
			uid:   1,
		},
	}

	svc := startup.InitInteractiveService()

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := svc.Collect(context.Background(), tc.biz, tc.bizId, tc.cid, tc.uid)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func (s *InteractiveTestSuite) TestGet() {
	testCases := []struct {
		name string

		before func(t *testing.T)

		bizId int64
		biz   string
		uid   int64

		wantErr error
		wantRes domain.Interactive
	}{
		{
			name:  "全部取出来了-无缓存",
			biz:   "test",
			bizId: 12,
			uid:   123,
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				err := s.db.WithContext(ctx).Create(&dao.Interactive{
					Biz:        "test",
					BizId:      12,
					ReadCnt:    100,
					CollectCnt: 200,
					LikeCnt:    300,
					Ctime:      123,
					Utime:      234,
				}).Error
				assert.NoError(t, err)
			},
			wantRes: domain.Interactive{
				BizId:      12,
				ReadCnt:    100,
				CollectCnt: 200,
				LikeCnt:    300,
			},
		},
		{
			name:  "全部取出来了-命中缓存-用户已点赞收藏",
			biz:   "test",
			bizId: 3,
			uid:   123,
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				err := s.db.WithContext(ctx).
					Create(&dao.UserCollectionBiz{
						Cid:   1,
						Biz:   "test",
						BizId: 3,
						Uid:   123,
						Ctime: 123,
						Utime: 124,
					}).Error
				assert.NoError(t, err)
				err = s.db.WithContext(ctx).
					Create(&dao.UserLikeBiz{
						Biz:    "test",
						BizId:  3,
						Uid:    123,
						Ctime:  123,
						Utime:  124,
						Status: 1,
					}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:3",
					"read_cnt", 0, "collect_cnt", 1).Err()
				assert.NoError(t, err)
			},
			wantRes: domain.Interactive{
				BizId:      3,
				CollectCnt: 1,
				Collected:  true,
				Liked:      true,
			},
		},
	}

	svc := startup.InitInteractiveService()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			res, err := svc.Get(context.Background(), tc.biz, tc.bizId, tc.uid)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func (s *InteractiveTestSuite) TestGetByIds() {
	preCtx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// 准备数据
	for i := 1; i < 5; i++ {
		i := int64(i)
		err := s.db.WithContext(preCtx).
			Create(&dao.Interactive{
				Id:         i,
				Biz:        "test",
				BizId:      i,
				ReadCnt:    i,
				CollectCnt: i + 1,
				LikeCnt:    i + 2,
			}).Error
		assert.NoError(s.T(), err)
	}

	testCases := []struct {
		name string

		before func(t *testing.T)
		biz    string
		ids    []int64

		wantErr error
		wantRes map[int64]domain.Interactive
	}{
		{
			name: "查找成功",
			biz:  "test",
			ids:  []int64{1, 2},
			wantRes: map[int64]domain.Interactive{
				1: {
					BizId:      1,
					ReadCnt:    1,
					CollectCnt: 2,
					LikeCnt:    3,
				},
				2: {
					BizId:      2,
					ReadCnt:    2,
					CollectCnt: 3,
					LikeCnt:    4,
				},
			},
		},
		{
			name:    "没有对应的数据",
			biz:     "test",
			ids:     []int64{100, 200},
			wantRes: map[int64]domain.Interactive{},
		},
	}

	svc := startup.InitInteractiveService()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			res, err := svc.GetByIds(context.Background(), tc.biz, tc.ids)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func TestInteractiveService(t *testing.T) {
	suite.Run(t, &InteractiveTestSuite{})
}
