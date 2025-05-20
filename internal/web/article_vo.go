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
}
