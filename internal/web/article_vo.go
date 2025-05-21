package web

type ArticleVO struct {
	Id       int64  `json:"id"`
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
	Content  string `json:"content"`
	Status   uint8  `json:"status"`
	Author   string `json:"author"`
	Ctime    string `json:"ctime"`
	Utime    string `json:"utime"`

	LikeCnt    int64 `json:"likeCnt"`    // 点赞数
	CollectCnt int64 `json:"collectCnt"` // 收藏数
	ReadCnt    int64 `json:"readCnt"`    // 阅读数

	Liked     bool `json:"liked"`     // 个人是否点赞
	Collected bool `json:"collected"` // 个人是否收藏
}
