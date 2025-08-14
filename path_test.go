package pathio_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/conceitho/pathio"
	"github.com/stretchr/testify/suite"
)

type SuiteTest struct {
	suite.Suite
	rootDir  string       // Diretório raíz do teste
	subDir   string       // Subdiretório criado por alguns testes
	rootPath pathio.IPath // Ponteiro para o path raíz
}

func (s *SuiteTest) SetupSuite() {
	s.rootDir = filepath.Join(os.TempDir(), "fluent_tests")
	s.subDir = "temp"
	// Cria diretório raíz de testes
	if err := os.MkdirAll(s.rootDir, 0755); err != nil {
		panic(err)
	}
	s.rootPath, _ = pathio.New(s.rootDir)
}

// this function executes after all tests executed
func (s *SuiteTest) TearDownSuite() {
	_ = os.RemoveAll(s.rootDir)
}

// this function executes before each test case
func (s *SuiteTest) SetupTest() {
	if err := s.rootPath.Reset(); err != nil {
		panic(err)
	}
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(SuiteTest))
}

func (s *SuiteTest) TestNew() {
	expected := s.rootDir
	p, e := pathio.New(expected)
	s.Nil(e)
	s.Equal(expected, p.Here())
}

func (s *SuiteTest) TestPathIO_AttachChild_WhenValidPath_ShouldReturnPath() {

}

func (s *SuiteTest) TestPathIO_DirExists_WhenValidPath_ShouldReturnTrue() {
	s.True(s.rootPath.DirExists())
	s.Equal(s.rootDir, s.rootPath.Here())
}

func (s *SuiteTest) TestPathIO_DirExists_WhenInvalidPath_ShouldReturnFalse() {
	expected := ""
	p, e := pathio.New(expected)
	s.Nil(p)
	s.ErrorIs(errors.Unwrap(e), pathio.ErrPathNotFound)
}

func (s *SuiteTest) TestPathIO_Relative_WhenValidPath_ShouldReturnLastPartOfPath() {
	expected := filepath.Base(s.rootDir)
	s.Equal(expected, s.rootPath.Relative())
}

func (s *SuiteTest) TestPathIO_CreateChild_WhenValidPath_ShouldReturnTrue() {
	expected := filepath.Join(s.rootDir, s.subDir)
	p, e := s.rootPath.CreateChild(s.subDir)
	s.Nil(e)
	s.True(p.DirExists())
	s.Equal(expected, p.Here())
}

func (s *SuiteTest) TestPathIO_HasChild_WhenEmptyCurrentDir_ShouldReturningFalse() {
	s.False(s.rootPath.HasChilds())
}

func (s *SuiteTest) TestPathIO_HasChild_WhenNotEmptyCurrentDir_ShouldReturningTrue() {
	_, _ = s.rootPath.CreateChild(s.subDir)
	s.True(s.rootPath.HasChilds())
}

func (s *SuiteTest) TestPathIO_FindChild_WhenInvalidPath_ShouldReturnNil() {
	path, ok := s.rootPath.FindChild(s.subDir)
	s.Nil(path)
	s.False(ok)
}

func (s *SuiteTest) TestPathIO_FindChild_WhenValidPath_ShouldReturnPath() {
	path, err := s.rootPath.CreateChild(s.subDir)
	s.Nil(err)
	s.NotNil(path)
	path, ok := s.rootPath.FindChild(s.subDir)
	s.True(ok)
	s.NotNil(path)
	s.Equal(s.subDir, path.Relative())
}

func (s *SuiteTest) TestPathIO_FileName_WhenInvalidFileName_ShouldReturnError() {
	path, err := s.rootPath.FileName("")
	s.Empty(path)
	s.NotNil(err)
	s.ErrorIs(err, pathio.ErrFileNameIsEmpty)
}

func (s *SuiteTest) TestPathIO_FileName_WhenValidFileName_ShouldReturnPathWithFileName() {
	fileName := "filename.txt"
	expected := filepath.Join(s.rootDir, fileName)
	path, err := s.rootPath.FileName(fileName)
	s.Nil(err)
	s.Equal(path, expected)
}

func (s *SuiteTest) TestPathIO_Files_WhenDirIsEmpty_ShouldReturnEmptyList() {
	fileList, err := s.rootPath.Files()
	s.ErrorIs(err, pathio.ErrDirIsEmpty)
	s.NotNil(fileList)
	s.Empty(fileList)
}

func (s *SuiteTest) TestPathIO_Files_WhenDirIsNotEmpty_ShouldReturnFileList() {
	defer func() {
		_ = s.resetRootDir
	}()
	for _, testCase := range []struct {
		files    []string
		expected int
	}{
		{files: []string{"file1.txt", "file2.txt", "file3.txt"}, expected: 3},
		{files: []string{"file1.txt"}, expected: 1},
		{files: []string{"file1.txt", "file2.txt"}, expected: 2},
	} {
		for _, fileName := range testCase.files {
			err := s.touchFile(fileName)
			s.Nil(err)
		}
		fileList, err := s.rootPath.Files()
		s.Nil(err)
		s.NotEmpty(fileList)
		s.Equal(testCase.expected, len(fileList))
		s.resetRootDir()
	}
}

func (s *SuiteTest) TestPathIO_FileByExt_WhenDirIsEmpty_ShouldReturnEmptyList() {
	fileList, err := s.rootPath.FilesByExt(".txt")
	s.ErrorIs(err, pathio.ErrFilesNotFound)
	s.NotNil(fileList)
	s.Empty(fileList)
}

func (s *SuiteTest) TestPathIO_FilesByExt_WhenDirIsNotEmpty_ShouldReturnFileList() {
	defer func() {
		s.resetRootDir()
	}()
	for _, testCase := range []struct {
		files    []string
		expected int
	}{
		{files: []string{"file1.txt", "file2.txt", "file3.dat"}, expected: 2},
		{files: []string{"file1.txt", "file2.csv"}, expected: 1},
	} {
		for _, fileName := range testCase.files {
			err := s.touchFile(fileName)
			s.Nil(err)
		}

		fileList, err := s.rootPath.FilesByExt(".txt")

		s.Nil(err)
		s.NotEmpty(fileList)
		s.Equal(testCase.expected, len(fileList))

		s.resetRootDir()
	}
}

func (s *SuiteTest) TestPathIO_Childs_WhenHasNoSubdir_ShouldReturnEmptyList() {
	l := s.rootPath.Childs()
	s.Empty(l)
}

func (s *SuiteTest) TestPathIO_Childs_WhenHasSubdir_ShouldReturnFillList() {
	for i := range 10 {
		_, _ = s.rootPath.CreateChild(fmt.Sprintf("_%d", i))
	}
	l := s.rootPath.Childs()
	s.Equal(10, len(l))
	s.resetRootDir()
}

func (s *SuiteTest) TestPathIO_Parent_WhenHasNoParent_ShouldReturnNil() {
	p := s.rootPath.Parent()
	s.Nil(p)
}
func (s *SuiteTest) TestPathIO_Parent_WhenHasParent_ShouldReturnPath() {
	p, err := s.rootPath.CreateChild(s.subDir)
	s.Nil(err)
	s.NotNil(p)
	s.Equal(s.subDir, p.Relative())
	s.resetRootDir()
}

func (s *SuiteTest) TestPathIO_AttachChild_WhenHasChilds_ShouldReturnFillList() {
	// Cria estrutura de diretórios
	_ = os.Mkdir(filepath.Join(s.rootDir, s.subDir), 0755)
	for i := range 10 {
		dirName := filepath.Join(s.rootDir, s.subDir, fmt.Sprintf("_%d", i))
		_ = os.Mkdir(dirName, 0755)
	}
	defer func() {
		s.resetRootDir()
	}()
	p, err := s.rootPath.AttachChild(s.subDir)
	s.Nil(err)
	c := p.Childs()
	s.Equal(s.subDir, p.Relative())
	s.NotEmpty(c)
	s.Equal(10, len(c))
}

func (s *SuiteTest) touchFile(fileName string) error {
	file, err := s.rootPath.FileName(fileName)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(file, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}

func (s *SuiteTest) resetRootDir() {
	_ = os.RemoveAll(s.rootDir)
	_ = os.Mkdir(s.rootDir, 0755)
	_ = s.rootPath.Reset()
}
