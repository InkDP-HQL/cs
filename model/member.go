package model

type member struct {
	ID       string `json:"id,omitempty"`
	Account  string `json:"account,omitempty"`
	Password string `json:"password,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`

	// 存储库信息
	Parcel string `json:"parcel,omitempty"`
	KeyID  string `json:"key,omitempty"`
	Secret string `json:"secret,omitempty"`
}
type Members struct {
	ID       string `json:"id,omitempty"`
	Account  string `json:"account,omitempty"`
	Password string `json:"password,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`

	// 存储库信息
	Parcel string `json:"parcel,omitempty"`
	KeyID  string `json:"key,omitempty"`
	Secret string `json:"secret,omitempty"`
}

func NewMember() *member {
	return &member{}
}

func (m *member) index() string {
	return "member"
}

func (m *member) id() string {
	return m.ID
}

func (m *member) SetID(id string) {
	m.ID = id
}

func (m *member) GetID() (id string) {
	return m.ID
}

// Add 添加用户
func (m *member) Add() error {
	if m.Account == "" {
		return ErrAccountIsEmpty
	}

	if m.Password == "" {
		return ErrPasswordIsEmpty
	}
	if m.Phone == "" {
		return ErrPhoneIsEmpty
	}
	if m.Email == "" {
		return ErrEmailIsEmpty
	}

	// 查看用户是否已注册
	me := new(member)
	err := NewDB().SetIndex(m.index()).Set("account", m.Account).Match(me)
	if err != nil && err != ErrIndexNotFoundException {
		return err
	}
	if me.Password != "" {
		return ErrRepeatingAccountName
	}

	// 添加用户记录
	m.ID = uuid()
	_, err = NewDB().SetIndex(m.index()).Create(m)
	if err != nil {
		return err
	}

	return nil
}

// Login 登录
func (m *member) Login() (me *member, err error) {
	if m.Account == "" {
		return nil, ErrAccountIsEmpty
	}

	if m.Password == "" {
		return nil, ErrPasswordIsEmpty
	}

	// 向存储请求记录
	me = new(member)
	err = NewDB().SetIndex(m.index()).Set("account", m.Account).Match(me)
	if err != nil {
		return
	}

	if me.Password == "" || (me.Password != m.Password) {
		return nil, ErrAccountNotMatchPassword
	}
	me.Password = ""

	return
}

// Update 修改用户信息
func (m *member) Update() (err error) {
	err = NewDB().SetIndex(m.index()).Update(m)
	if err != nil {
		return
	}
	return
}

// Read 读取用户信息
func (m *member) Read() (err error) {
	err = NewDB().SetIndex(m.index()).Set("id", m.ID).Match(m)
	if err != nil {
		return
	}
	return
}
