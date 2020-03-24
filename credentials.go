package dodod

type DbCredentials interface {
	ReadPath() (dbPath string, err error)
	ReadPassword() (password string, err error)
}
