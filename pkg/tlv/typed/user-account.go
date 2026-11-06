package typed

type UserName string

func (UserName) Type() Type { return TypeUserName }

func (m UserName) Encode() ([]byte, error)    { return stringToBytes(m, 1, 16) }
func (m *UserName) Decode(value []byte) error { return bytesTostring(value, m) }

type Password string

func (Password) Type() Type { return TypePassword }

func (m Password) Encode() ([]byte, error)    { return stringToBytes(m, 6, 16) }
func (m *Password) Decode(value []byte) error { return bytesTostring(value, m) }

type NewUserName string

func (NewUserName) Type() Type { return TypeNewUserName }

func (m NewUserName) Encode() ([]byte, error)    { return stringToBytes(m, 1, 16) }
func (m *NewUserName) Decode(value []byte) error { return bytesTostring(value, m) }

type NewPassword string

func (NewPassword) Type() Type { return TypeNewPassword }

func (m NewPassword) Encode() ([]byte, error)    { return stringToBytes(m, 6, 16) }
func (m *NewPassword) Decode(value []byte) error { return bytesTostring(value, m) }
