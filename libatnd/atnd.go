package libatnd

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const envDBNAME = "MILBOT_ATND_DB_NAME"

var atnd = new(attend)

// InvalidMACAddressError は不正な MAC アドレスを示します。
type InvalidMACAddressError struct {
	addr string
}

// Error です。
func (e InvalidMACAddressError) Error() string {
	return fmt.Sprintf("invalid MAC address: %s", e.addr)
}

// IsValidMACAddress は addr が有効な MAC アドレスかどうかを返します。
func IsValidMACAddress(addr string) bool {
	runes := []rune(strings.ToLower(addr))

	if len(runes) != 17 {
		return false
	}

	for i, r := range runes {
		if i%3 == 2 {
			if r != ':' {
				return false
			}
		} else {
			if !strings.ContainsRune("0123456789abcdef", r) {
				return false
			}
		}
	}

	return true
}

type attend struct {
	started bool // 起動済みかどうか
	db      *sql.DB

	// Slack ID と最後に観測した時間との対応。
	// Bot の起動から観測してない場合は nil が入る。
	stat map[string]*time.Time
}

// init は attend を初期化します。
func (a *attend) init() error {
	if a.started {
		return errors.New("attend already started")
	}
	a.started = true

	if err := a.connectDB(); err != nil {
		return fmt.Errorf("atnd connect db error: %w", err)
	}

	if err := a.initDB(); err != nil {
		return fmt.Errorf("atnd initDB error: %w", err)
	}

	if err := a.initStatus(); err != nil {
		return fmt.Errorf("atnd initStatus error: %w", err)
	}

	return nil
}

// connectDB でデータベースに接続します。
func (a *attend) connectDB() error {
	dbpath, err := a.dbPath()
	if err != nil {
		return err
	}
	a.db, err = sql.Open("sqlite3", dbpath)
	if err != nil {
		return err
	}
	return a.db.Ping()
}

// dbPath は database のデータのパスを返します。
func (*attend) dbPath() (string, error) {
	fname, ok := os.LookupEnv(envDBNAME)
	if !ok {
		return "", errors.New(envDBNAME + " not found")
	}
	binpath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(binpath), fname), nil
}

// initDB はデータベースを初期化します。
func (a *attend) initDB() error {
	cmd := "create table if not exist members(name text primary key, address text not null);"
	if _, err := a.db.Exec(cmd); err != nil {
		return err
	}
	return nil
}

// members はデータベースに保存されている全員の名前を返します。
func (a *attend) membersContext(ctx context.Context) ([]string, error) {
	cmd := `select slack_id from members;`
	rows, err := a.db.QueryContext(ctx, cmd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []string{}
	var name string
	for rows.Next() {
		if err := rows.Scan(&name); err != nil {
			return res, err
		}
		res = append(res, name)
	}
	if rows.Err() != nil {
		return res, rows.Err()
	}
	return res, nil
}

// initStatus は status を初期化します。
func (a *attend) initStatus() error {
	members, err := a.membersContext(context.Background())
	if err != nil {
		return err
	}

	a.stat = map[string]*time.Time{}
	for _, member := range members {
		a.stat[member] = nil
	}
	return nil
}

// status は現在の在室状況を返します。
func (a *attend) status() map[string]*time.Time {
	res := map[string]*time.Time{}

	for name, time := range a.stat {
		if time == nil {
			continue
		}
		res[name] = time
	}

	return res
}

// setMemberContext はメンバー登録または修正します。
func (a *attend) setMemberContext(ctx context.Context, name, addr string) error {
	if !IsValidMACAddress(addr) {
		return InvalidMACAddressError{addr: addr}
	}

	hashedAddr := sha256.Sum256([]byte(addr))

	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("set member failed: %w", err)
	}

	var tmpName string
	err = a.db.QueryRowContext(ctx, `select * from members where name = $1`, name).Scan(&tmpName)
	if errors.Is(err, sql.ErrNoRows) {
		cmd := `insert into members value($1, $2)`
		if _, err := a.db.ExecContext(ctx, cmd, name, hashedAddr); err != nil {
			// TODO
		}
	}

	return nil
}

// Start で在室確認を開始します。一度しか実行できません。
func Start() error {
	return nil
}

// Status は在室状況を返します。
func Status() map[string]*time.Time {
	return nil
}

// SetMember は name の addr を登録します
func SetMember(name, addr string) error {
	return nil
}

// DeleteMember は name を削除します。
func DeleteMember(name string) error {
	return nil
}
