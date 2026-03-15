package entity

type Chat struct {
	ID         uint
	Name       string
	AvatarPath string
	Type       string
}

type CreateGroupChatInput struct {
	Name string
	File *File
}

type CreateChatInput struct {
	Type       string
	Name       string
	AvatarPath string
}

type UpdateChatInput struct {
	Name *string
	File *File
}
