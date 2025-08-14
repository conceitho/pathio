package pathio

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrPathNotFound    = errors.New("diretório não localizado")
	ErrDirIsEmpty      = errors.New("diretório está vazio")
	ErrFilesNotFound   = errors.New("não foram localizados arquivos")
	ErrFileNameIsEmpty = errors.New("nome do arquivo está vazio")
)

type IPath interface {
	CreateChild(childPath string) (IPath, error)
	CreateChilds(childPath ...string) (IPath, error)
	AttachChild(childPath string) (IPath, error)
	Here() string
	Relative() string
	Parent() IPath
	Childs() []IPath
	HasChilds() bool
	FindChild(directory string) (IPath, bool)
	FileName(name string) (string, error)
	Files() ([]string, error)
	FilesByExt(mask string) ([]string, error)
	DirExists() bool
	Reset() error
}

type pathIO struct {
	here   string
	parent IPath
	childs map[string]IPath
}

func New(absolutePath string) (IPath, error) {
	r, err := newPathIO(nil, absolutePath, true)
	if err != nil {
		return nil, fmt.Errorf("error creating new pathIO: %w", err)
	}
	return r, nil
}

func (p *pathIO) CreateChild(childPath string) (IPath, error) {
	result, err := p.AttachChild(childPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao anexar diretório %s: %v", childPath, err)
	}
	if !result.DirExists() {
		if err := os.Mkdir(result.Here(), 0755); err != nil {
			return nil, fmt.Errorf("falha ao criar diretório %s: %v", result.Here(), err)
		}
	}
	return result, nil
}

func (p *pathIO) CreateChilds(childPath ...string) (IPath, error) {
	var (
		parent IPath = p
		result IPath
		err    error
	)
	for _, subdir := range childPath {
		result, err = parent.CreateChild(subdir)
		if err != nil {
			return parent, fmt.Errorf("criando subdiretório %s: %v", filepath.Join(parent.Here(), subdir), err)
		}
		parent = result
	}
	return parent, nil
}

func (p *pathIO) AttachChild(childPath string) (IPath, error) {
	result, ok := p.FindChild(childPath)
	if !ok {
		var err error
		result, err = newPathIO(p, childPath, false)
		if err != nil {
			return nil, fmt.Errorf("erro criando path %s: %v", childPath, err)
		}
	}
	return result, nil
}

func (p *pathIO) Here() string {
	return p.here
}

func (p *pathIO) Relative() string {
	return filepath.Base(p.here)
}

func (p *pathIO) Parent() IPath {
	return p.parent
}

func (p *pathIO) Childs() []IPath {
	var result = make([]IPath, len(p.childs))
	i := 0
	for _, v := range p.childs {
		result[i] = v
		i++
	}
	return result
}

func (p *pathIO) HasChilds() bool {
	return len(p.childs) > 0
}

func (p *pathIO) FindChild(directory string) (IPath, bool) {
	result, ok := p.childs[directory]
	return result, ok
}

func (p *pathIO) FileName(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", ErrFileNameIsEmpty
	}
	return filepath.Join(p.here, name), nil
}

func (p *pathIO) Files() ([]string, error) {
	entries, err := os.ReadDir(p.here)
	if err != nil {
		return nil, fmt.Errorf("erro lendo arquivos: %w", err)
	}
	result := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		result = append(result, e.Name())
	}
	if len(result) == 0 {
		return result, ErrDirIsEmpty
	}
	return result, nil
}

func (p *pathIO) FilesByExt(mask string) ([]string, error) {
	entries, err := os.ReadDir(p.here)
	if err != nil {
		return nil, fmt.Errorf("erro lendo arquivo do tipo %s: %v", mask, err)
	}
	result := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) == mask {
			result = append(result, e.Name())
		}
	}
	if len(result) == 0 {
		return result, errors.Join(ErrFilesNotFound, fmt.Errorf(". extensão \"%s\"", mask))
	}
	return result, nil
}

func (p *pathIO) DirExists() bool {
	return exists(p.here)
}

func (p *pathIO) Reset() error {
	return p.reset(true)
}

func newPathIO(parent IPath, path string, requiredValidPath bool) (IPath, error) {
	result := &pathIO{
		here:   path,
		parent: parent,
		childs: make(map[string]IPath),
	}
	if parent != nil {
		result.here = filepath.Join(parent.Here(), path, "/")
		x, _ := parent.(*pathIO)
		x.addChild(result)
	}
	if result.DirExists() {
		if err := result.attachDirs(); err != nil {
			return nil, fmt.Errorf("erro lendo subdiretórios: %v", err)
		}
	} else if requiredValidPath {
		return nil, fmt.Errorf("path válido é requerido: %w", ErrPathNotFound)
	}
	return result, nil
}

func (p *pathIO) attachDirs() error {
	entries, err := os.ReadDir(p.here)
	if err != nil {
		return fmt.Errorf("falha ao ler diretório %s: %v", p.here, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			_, err := p.AttachChild(e.Name())
			if err != nil {
				return fmt.Errorf("falha ao anexar diretório %s: %v", e.Name(), err)
			}
		}
	}
	return nil
}

func exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (p *pathIO) addChild(path IPath) {
	p.childs[path.Relative()] = path
}

func (p *pathIO) reset(attachDirs bool) error {
	var sdir *pathIO
	for _, path := range p.childs {
		sdir = path.(*pathIO)
		if sdir == nil {
			continue
		}
		if err := sdir.reset(false); err != nil {
			return fmt.Errorf("fail reset dir: %v", err)
		}
	}
	clear(p.childs)
	if attachDirs && sdir != nil {
		return sdir.attachDirs()
	}
	return nil
}
