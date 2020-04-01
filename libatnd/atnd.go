package libatnd

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// configFileName はメンバーのアドレスを保管しておくファイルの名前です。
const configFileName = "atnd_config.json"

// configPerm は config ファイルのパーミッションです。
const configPerm = 0600

// keyFileName は Bluetooth アドレスを暗号化するためのキーのファイルです。
const encKeyFileName = ".atnd_key"

// encKeyPerm は暗号化キーファイルのパーミッションです。
const encKeyPerm = 0600

// atnd は Atnd のシングルトンです。
var atnd *Atnd

func init() {
	var err error
	atnd, err = newAtnd()
	if err != nil {
		log.Fatal("create Atnd error:", err)
	}
}

// InvalidNameError は name が使えないときのエラーです。
type InvalidNameError struct {
	Name string
}

func (e InvalidNameError) Error() string {
	return fmt.Sprintf("invalid name: %q", e.Name)
}

// InvalidMACAddressError は不正な MAC アドレスを示します。
type InvalidMACAddressError struct {
	Address string
}

// Error です。
func (e InvalidMACAddressError) Error() string {
	return fmt.Sprintf("invalid MAC address: %s", e.Address)
}

// MemberNotExistError はメンバーが存在しないことを表すエラーです。
type MemberNotExistError struct {
	Name string
}

// Error です。
func (e MemberNotExistError) Error() string {
	return fmt.Sprintf("member not exist: %q", e.Name)
}

// ErrBluetoothNotAvailable は Bluetooth が落ちてることを示すエラーです。
var ErrBluetoothNotAvailable = errors.New("bluetooth not available")

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

// ErrL2pingNotFound は l2ping が $PATH にないときのエラーです。
var ErrL2pingNotFound = errors.New("l2ping not found in $PATH")

// Atnd は在室判定をする構造体です。
type Atnd struct {
	// 設定ファイルのパスです。
	confPath string

	// 設定ファイルの中身です。
	muConfig *sync.RWMutex
	config   *config

	// Bluetooth アドレスを暗号化するキーです。
	encKey []byte

	// メンバーの名前と最後に観測した時間との対応。
	// Bot の起動から観測してない場合は nil が入る。
	muStatus *sync.RWMutex
	status   map[string]*time.Time

	// Search は同時に実行できないのでセマフォを使います。
	semaSearch       chan struct{}
	semaSearchMember chan struct{}
}

// Instance はシングルトンの Atnd を返します。
func Instance() *Atnd {
	return atnd
}

// newAttend は Atnd を作って返します。
func newAtnd() (*Atnd, error) {
	a := new(Atnd)

	var err error
	a.confPath, err = a.configPath()
	if err != nil {
		return nil, fmt.Errorf("create new Atnd failed: %w", err)
	}
	if err := a.initConfig(); err != nil {
		return nil, fmt.Errorf("create new Atnd failed: %w", err)
	}

	if err := a.initEncKey(); err != nil {
		return nil, fmt.Errorf("create new Atnd failed: %w", err)
	}

	a.initStatus()

	a.semaSearch = make(chan struct{}, 1)
	a.semaSearchMember = make(chan struct{}, 1)

	return a, nil
}

// initConfig は必要であれば config ファイルを生成して a.config を初期化します。
func (a *Atnd) initConfig() error {
	// config ファイルが無ければ生成
	_, err := os.Stat(a.confPath)
	if os.IsNotExist(err) {
		if err := a.createConfigFile(); err != nil {
			return fmt.Errorf("cannot init config: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("cannot init config: %w", err)
	}

	conf, err := a.loadConfigFile()
	if err != nil {
		return fmt.Errorf("cannot init config: %w", err)
	}

	a.muConfig = new(sync.RWMutex)
	a.config = conf

	return nil
}

// createConfigFile は config ファイルを生成します。
func (a *Atnd) createConfigFile() error {
	conf := newConfig()

	bytes, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot create config file: %w", err)
	}

	if err := ioutil.WriteFile(a.confPath, bytes, configPerm); err != nil {
		return fmt.Errorf("cannot create config file: %w", err)
	}

	return nil
}

// loadConfigFile は設定ファイルをファイルから読みます。
func (a *Atnd) loadConfigFile() (conf *config, err error) {
	bytes, err := ioutil.ReadFile(a.confPath)
	if err != nil {
		err = fmt.Errorf("cannot load config file: %w", err)
		return
	}

	if err = json.Unmarshal(bytes, &conf); err != nil {
		err = fmt.Errorf("cannot load config file: %w", err)
		return
	}

	return
}

// dumpConfig は config を書き出します。
func (a *Atnd) dumpConfig() error {
	bytes, err := json.MarshalIndent(a.config, "", "  ")
	if err != nil {
		return fmt.Errorf("dump config error: %w", err)
	}

	if err := ioutil.WriteFile(a.confPath, bytes, configPerm); err != nil {
		return fmt.Errorf("dump config error: %w", err)
	}

	return nil
}

// configPath は設定ファイルの場所を返します。
func (a *Atnd) configPath() (string, error) {
	realExec, err := a.realExecPath()
	if err != nil {
		return "", fmt.Errorf("cannot get config path: %w", err)
	}
	return filepath.Join(filepath.Dir(realExec), configFileName), nil
}

// realExecDir は実行ファイルの実体のパスを返します。
func (*Atnd) realExecPath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot get real executable path: %w", err)
	}

	realExec, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", fmt.Errorf("cannot get real executable path: %w", err)
	}

	return realExec, nil
}

// generateNewKey は新しい暗号化キーを生成します。
func (*Atnd) generateNewKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate new key failed: %w", err)
	}
	return key, nil
}

// encKeyPath は暗号化のキーファイルのパスを変えします。
func (a *Atnd) encKeyPath() (string, error) {
	realExec, err := a.realExecPath()
	if err != nil {
		return "", fmt.Errorf("cannot get enc key path: %w", err)
	}
	return filepath.Join(filepath.Dir(realExec), encKeyFileName), nil
}

// initEncEey は必要に応じてキーファイルを生成して encKey を初期化します。
func (a *Atnd) initEncKey() error {
	// キーファイルがなければ生成
	encPath, err := a.encKeyPath()
	if err != nil {
		return fmt.Errorf("init enc key failed: %w", err)
	}

	_, err = os.Stat(encPath)
	if os.IsNotExist(err) {
		if err := a.createEncKeyFile(encPath); err != nil {
			return fmt.Errorf("init enc key failed: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("init enc key failed: %w", err)
	}

	key, err := a.loadEncKeyFile(encPath)
	if err != nil {
		return fmt.Errorf("init enc key failed: %w", err)
	}

	a.encKey = key
	return nil
}

// createEncKeyFile は暗号化キーファイルを生成します。
func (a *Atnd) createEncKeyFile(encPath string) error {
	key, err := a.generateNewKey()
	if err != nil {
		return fmt.Errorf("create enc key file failed: %w", err)
	}

	if err := ioutil.WriteFile(encPath, key, encKeyPerm); err != nil {
		return fmt.Errorf("create enc key failed: %w", err)
	}

	return nil
}

// loadEncKeyFile はファイルから暗号化キーをロードします。
func (a *Atnd) loadEncKeyFile(encPath string) ([]byte, error) {
	key, err := ioutil.ReadFile(encPath)
	if err != nil {
		return nil, fmt.Errorf("load enc key failed: %w", err)
	}
	return key, nil
}

// encode は bytes をエンコードします。
func (a *Atnd) encode(plain string) ([]byte, error) {
	block, err := aes.NewCipher(a.encKey)
	if err != nil {
		return nil, fmt.Errorf("encode error: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("encode error: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("encode error: %w", err)
	}

	encoded := gcm.Seal(nil, nonce, []byte(plain), nil)
	encoded = append(nonce, encoded...)
	return encoded, nil
}

// decode は encoded をデコードします。
func (a *Atnd) decode(encoded []byte) (string, error) {
	block, err := aes.NewCipher(a.encKey)
	if err != nil {
		return "", fmt.Errorf("decode error: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("decode error: %w", err)
	}

	nonce := encoded[:gcm.NonceSize()]
	plain, err := gcm.Open(nil, nonce, encoded[gcm.NonceSize():], nil)
	if err != nil {
		return "", fmt.Errorf("decode error: %w", err)
	}

	return string(plain), nil
}

// initStatus は a.config から a.status を初期化します。
func (a *Atnd) initStatus() {
	a.muStatus = new(sync.RWMutex)

	a.status = map[string]*time.Time{}
	for _, member := range a.config.Members {
		a.status[member.Name] = nil
	}
}

// Status は出席状況を返します。最後に在室した時間の近い順にソートされています。
func (a *Atnd) Status() []*Attendance {
	res := []*Attendance{}

	a.muStatus.RLock()
	defer a.muStatus.RUnlock()

	for name, atndTime := range a.status {
		if atndTime == nil {
			continue
		}
		res = append(res, &Attendance{Name: name, Time: *atndTime})
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Time.Before(res[j].Time) })

	return res
}

// SetMember は name を addr の状態にセットします。
func (a *Atnd) SetMember(name, addr string) error {
	if name == "" {
		return InvalidNameError{Name: name}
	}
	if !IsValidMACAddress(addr) {
		return InvalidMACAddressError{Address: addr}
	}

	encodedAddr, err := a.encode(addr)
	if err != nil {
		return fmt.Errorf("set member error: %w", err)
	}

	a.muConfig.Lock()
	defer a.muConfig.Unlock()

	found := false
	for _, mem := range a.config.Members {
		if mem.Name == name {
			found = true
			break
		}
	}

	if found {
		a.updateMember(name, encodedAddr)
	} else {
		a.addMember(name, encodedAddr)
	}

	if err := a.dumpConfig(); err != nil {
		return fmt.Errorf("set member error: %w", err)
	}

	return nil
}

// addMember はメンバー情報を追加します。
func (a *Atnd) addMember(name string, encodedAddr []byte) {
	newMember := member{Name: name, EncodedAddress: encodedAddr}
	a.config.Members = append(a.config.Members, &newMember)
}

// updateMember は name の addr を変更します。
func (a *Atnd) updateMember(name string, encodedAddr []byte) {
	for i := range a.config.Members {
		if a.config.Members[i].Name == name {
			a.config.Members[i].EncodedAddress = encodedAddr
			return
		}
	}
}

// DeleteMember は name のメンバーを消し去ります。
func (a *Atnd) DeleteMember(name string) error {
	a.muConfig.Lock()
	defer a.muConfig.Unlock()

	for i := 0; i < len(a.config.Members); i++ {
		if a.config.Members[i].Name == name {
			a.config.Members = append(a.config.Members[:i], a.config.Members[i+1:]...)
			if err := a.dumpConfig(); err != nil {
				return fmt.Errorf("delete member error: %w", err)
			}
			return nil
		}
	}

	return MemberNotExistError{Name: name}
}

// SearchContext はメンバーをサーチして出席している人のリストを返します。
func (a *Atnd) SearchContext(ctx context.Context) ([]*Attendance, error) {
	res := []*Attendance{}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case a.semaSearch <- struct{}{}:
		defer func() { <-a.semaSearch }()

		for _, mem := range a.config.Members {
			attendance, err := a.SearchMemberContext(ctx, mem.Name)
			if err != nil {
				return nil, fmt.Errorf("search failed: %w", err)
			}
			if attendance != nil {
				res = append(res, attendance)
			}
		}
	}

	return res, nil
}

// Search はメンバーをサーチして出席している人のリストを返します。
func (a *Atnd) Search() ([]*Attendance, error) {
	return a.SearchContext(context.Background())
}

// SearchMemberContext はひとりのメンバーをサーチします。いなかったら nil です。
func (a *Atnd) SearchMemberContext(ctx context.Context, name string) (*Attendance, error) {
	encodedAddr, err := a.findAddr(name)
	if err != nil {
		return nil, fmt.Errorf("search member failed: %w", err)
	}
	addr, err := a.decode(encodedAddr)
	if err != nil {
		return nil, fmt.Errorf("search member failed: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case a.semaSearchMember <- struct{}{}:
		defer func() { <-a.semaSearchMember }()

		exist, err := a.sendPing(ctx, addr)
		if err != nil {
			return nil, fmt.Errorf("search member failed: %w", err)
		}

		if exist {
			now := time.Now()
			a.updateStatus(name, &now)
			return &Attendance{Name: name, Time: now}, nil
		}
	}

	return nil, nil
}

// findAddr は name に対応する暗号化された address を返します。
func (a *Atnd) findAddr(name string) ([]byte, error) {
	a.muConfig.RLock()
	defer a.muConfig.RUnlock()

	for _, mem := range a.config.Members {
		if mem.Name == name {
			return mem.EncodedAddress, nil
		}
	}

	return nil, MemberNotExistError{Name: name}
}

// sendPing は メンバーがいる場合に true になります。
func (*Atnd) sendPing(ctx context.Context, addr string) (bool, error) {
	stdout := new(bytes.Buffer)
	cmd := exec.CommandContext(ctx, "l2ping", "-c", "1", addr)
	cmd.Stdout = stdout

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return false, ErrL2pingNotFound
		} else if strings.Contains(stdout.String(), "No route to host") {
			return false, ErrBluetoothNotAvailable
		}

		return false, nil
	}

	return true, nil
}

// updateStatus は name の在室時間を ts に更新します。
func (a *Atnd) updateStatus(name string, ts *time.Time) {
	a.muStatus.Lock()
	defer a.muStatus.Unlock()

	a.status[name] = ts
}

// Members は登録されているメンバーの名前のリストを返します。
func (a *Atnd) Members() []string {
	res := []string{}

	a.muConfig.RLock()
	defer a.muConfig.RUnlock()

	for _, mem := range a.config.Members {
		res = append(res, mem.Name)
	}

	return res
}

// config は設定ファイルの構造体です。
type config struct {
	Members []*member `json:"members"`
}

// newConfig は初期状態の config を返します。
func newConfig() *config {
	return &config{
		Members: []*member{},
	}
}

// member は設定ファイルのメンバーを表します。
type member struct {
	Name           string `json:"name"`            // 表示名です。
	EncodedAddress []byte `json:"encoded_address"` // 暗号化された Bluetooth アドレスです。
}

// Attendance はそのメンバーの最後に出席した時間を表します。
type Attendance struct {
	Name string    // 表示名です。
	Time time.Time // 最後に在室確認した時間です。
}
