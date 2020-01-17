package run

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/pelotech/drone-helm3/internal/env"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

type DepUpdateTestSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	mockCmd         *Mockcmd
	originalCommand func(string, ...string) cmd
}

func (suite *DepUpdateTestSuite) BeforeTest(_, _ string) {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockCmd = NewMockcmd(suite.ctrl)

	suite.originalCommand = command
	command = func(path string, args ...string) cmd { return suite.mockCmd }
}

func (suite *DepUpdateTestSuite) AfterTest(_, _ string) {
	command = suite.originalCommand
}

func TestDepUpdateTestSuite(t *testing.T) {
	suite.Run(t, new(DepUpdateTestSuite))
}

func (suite *DepUpdateTestSuite) TestNewDepUpdate() {
	cfg := env.Config{
		Chart: "scatterplot",
	}
	d := NewDepUpdate(cfg)
	suite.Equal("scatterplot", d.chart)
}

func (suite *DepUpdateTestSuite) TestPrepareAndExecute() {
	defer suite.ctrl.Finish()

	stdout := strings.Builder{}
	stderr := strings.Builder{}
	cfg := env.Config{
		Chart:  "your_top_songs_2019",
		Stdout: &stdout,
		Stderr: &stderr,
	}

	command = func(path string, args ...string) cmd {
		suite.Equal(helmBin, path)
		suite.Equal([]string{"dependency", "update", "your_top_songs_2019"}, args)

		return suite.mockCmd
	}
	suite.mockCmd.EXPECT().
		Stdout(&stdout)
	suite.mockCmd.EXPECT().
		Stderr(&stderr)
	suite.mockCmd.EXPECT().
		Run().
		Times(1)

	d := NewDepUpdate(cfg)

	suite.Require().NoError(d.Prepare())
	suite.NoError(d.Execute())
}

func (suite *DepUpdateTestSuite) TestPrepareNamespaceFlag() {
	defer suite.ctrl.Finish()

	cfg := env.Config{
		Namespace: "spotify",
		Chart:     "your_top_songs_2019",
	}

	command = func(path string, args ...string) cmd {
		suite.Equal([]string{"--namespace", "spotify", "dependency", "update", "your_top_songs_2019"}, args)

		return suite.mockCmd
	}
	suite.mockCmd.EXPECT().Stdout(gomock.Any()).AnyTimes()
	suite.mockCmd.EXPECT().Stderr(gomock.Any()).AnyTimes()

	d := NewDepUpdate(cfg)

	suite.Require().NoError(d.Prepare())
}

func (suite *DepUpdateTestSuite) TestPrepareDebugFlag() {
	defer suite.ctrl.Finish()

	stdout := strings.Builder{}
	stderr := strings.Builder{}
	cfg := env.Config{
		Chart:  "your_top_songs_2019",
		Debug:  true,
		Stdout: &stdout,
		Stderr: &stderr,
	}

	command = func(path string, args ...string) cmd {
		suite.mockCmd.EXPECT().
			String().
			Return(fmt.Sprintf("%s %s", path, strings.Join(args, " ")))

		return suite.mockCmd
	}
	suite.mockCmd.EXPECT().Stdout(gomock.Any()).AnyTimes()
	suite.mockCmd.EXPECT().Stderr(gomock.Any()).AnyTimes()

	d := NewDepUpdate(cfg)

	suite.Require().NoError(d.Prepare())

	want := fmt.Sprintf("Generated command: '%s --debug dependency update your_top_songs_2019'\n", helmBin)
	suite.Equal(want, stderr.String())
	suite.Equal("", stdout.String())
}

func (suite *DepUpdateTestSuite) TestPrepareChartRequired() {
	d := NewDepUpdate(env.Config{})

	suite.mockCmd.EXPECT().Stdout(gomock.Any()).AnyTimes()
	suite.mockCmd.EXPECT().Stderr(gomock.Any()).AnyTimes()

	err := d.Prepare()
	suite.EqualError(err, "chart is required")
}
