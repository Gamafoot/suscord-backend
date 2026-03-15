package entity

type User struct {
	ID         uint
	Username   string
	AvatarPath string
}

type UnsafeUser struct {
	ID         uint
	Username   string
	Password   string
	AvatarPath string
}

type UpdateUserInput struct {
	Username *string
	File     *File
}
