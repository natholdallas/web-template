# 后端常用代码块

## 关联表-多对一

```go
type User struct {
	User     User           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user"`        // User
	UserID   uint           `gorm:"column:user_id;comment:User" json:"user_id"`                      // User ID
}
```

## orms.Dict

使用 `orms.Dict` 的时候, type 设置为 `json`
