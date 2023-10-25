package builder

import (
	"bytes"
	"io"
	"sync"
	"syscall"
	"time"

	"github.com/Cloud-Foundations/Dominator/imagebuilder/logarchiver"
	"github.com/Cloud-Foundations/Dominator/lib/filesystem/util"
	"github.com/Cloud-Foundations/Dominator/lib/filter"
	"github.com/Cloud-Foundations/Dominator/lib/hash"
	"github.com/Cloud-Foundations/Dominator/lib/image"
	"github.com/Cloud-Foundations/Dominator/lib/log"
	"github.com/Cloud-Foundations/Dominator/lib/slavedriver"
	"github.com/Cloud-Foundations/Dominator/lib/srpc"
	"github.com/Cloud-Foundations/Dominator/lib/triggers"
	proto "github.com/Cloud-Foundations/Dominator/proto/imaginator"
)

// Private interface types.

type buildLogger interface {
	Bytes() []byte
	io.Writer
}

type environmentGetter interface {
	getenv() map[string]string
}

type imageBuilder interface {
	build(b *Builder, client *srpc.Client, request proto.BuildImageRequest,
		buildLog buildLogger) (*image.Image, error)
}

// Other private types.

type argList []string

type bootstrapStream struct {
	builder          *Builder
	name             string
	BootstrapCommand []string
	*filter.Filter
	imageFilter      *filter.Filter
	ImageFilterUrl   string
	imageTriggers    *triggers.Triggers
	ImageTriggersUrl string
	PackagerType     string
}

type buildResultType struct {
	imageName  string
	startTime  time.Time
	finishTime time.Time
	buildLog   []byte
	error      error
}

type currentBuildInfo struct {
	buffer       *bytes.Buffer
	slaveAddress string
	startedAt    time.Time
}

type dependencyDataType struct {
	fetchLog           []byte
	generatedAt        time.Time
	streamToSource     map[string]string // K: stream name, V: source stream.
	unbuildableSources map[string]struct{}
}

type imageStreamsConfigurationType struct {
	Streams map[string]*imageStreamType `json:",omitempty"`
}

type imageStreamType struct {
	builder           *Builder
	builderUsers      map[string]struct{}
	name              string
	BuilderGroups     []string
	BuilderUsers      []string
	ManifestUrl       string
	ManifestDirectory string
	Variables         map[string]string
}

type inodeData struct {
	ctime syscall.Timespec
	hash  hash.Hash
	size  uint64
}

type listCommandType struct {
	ArgList        argList
	SizeMultiplier uint64
}

type manifestConfigType struct {
	SourceImage string
	*filter.Filter
}

type masterConfigurationType struct {
	BindMounts                []string                    `json:",omitempty"`
	BootstrapStreams          map[string]*bootstrapStream `json:",omitempty"`
	ImageStreamsCheckInterval uint                        `json:",omitempty"`
	ImageStreamsToAutoRebuild []string                    `json:",omitempty"`
	ImageStreamsUrl           string                      `json:",omitempty"`
	PackagerTypes             map[string]packagerType     `json:",omitempty"`
	RelationshipsQuickLinks   []WebLink                   `json:",omitempty"`
}

// manifestLocationType contains the expanded location of a manifest. These
// data may include secrets (i.e. username and password).
type manifestLocationType struct {
	directory string
	url       string
}

type manifestType struct {
	filter          *filter.Filter
	sourceImageInfo *sourceImageInfoType
}

type packagerType struct {
	CleanCommand   argList
	InstallCommand argList
	ListCommand    listCommandType
	RemoveCommand  argList
	UpdateCommand  argList
	UpgradeCommand argList
	Verbatim       []string
}

type sourceImageInfoType struct {
	computedFiles []util.ComputedFile
	filter        *filter.Filter
	imageName     string
	treeCache     *treeCache
	triggers      *triggers.Triggers
}

type testResultType struct {
	buffer   chan byte
	duration time.Duration
	err      error
	prog     string
}

type treeCache struct {
	hitBytes    uint64
	inodeTable  map[uint64]inodeData
	numHits     uint64
	pathToInode map[string]uint64
}

type WebLink struct {
	Name string
	URL  string
}

type BuildLocalOptions struct {
	BindMounts        []string
	ManifestDirectory string
	Variables         map[string]string
}

type Builder struct {
	buildLogArchiver          logarchiver.BuildLogArchiver
	bindMounts                []string
	generateDependencyTrigger chan<- chan<- struct{}
	stateDir                  string
	imageServerAddress        string
	logger                    log.DebugLogger
	imageStreamsUrl           string
	initialNamespace          string // For catching golang bugs.
	minimumExpirationDuration time.Duration
	streamsLoadedChannel      <-chan struct{} // Closed when streams loaded.
	streamsLock               sync.RWMutex
	bootstrapStreams          map[string]*bootstrapStream
	imageStreams              map[string]*imageStreamType
	imageStreamsToAutoRebuild []string
	relationshipsQuickLinks   []WebLink
	slaveDriver               *slavedriver.SlaveDriver
	buildResultsLock          sync.RWMutex
	currentBuildInfos         map[string]*currentBuildInfo // Key: stream name.
	lastBuildResults          map[string]buildResultType   // Key: stream name.
	packagerTypes             map[string]packagerType
	variables                 map[string]string
	dependencyDataLock        sync.RWMutex
	dependencyData            *dependencyDataType
	dependencyDataAttempt     time.Time
	dependencyDataError       error
}

type BuilderOptions struct {
	ConfigurationURL          string
	ImageRebuildInterval      time.Duration
	ImageServerAddress        string
	MinimumExpirationDuration time.Duration // Default: 15 minutes. Min: 5 min.
	StateDirectory            string
	VariablesFile             string
}

type BuilderParams struct {
	BuildLogArchiver logarchiver.BuildLogArchiver
	Logger           log.DebugLogger
	SlaveDriver      *slavedriver.SlaveDriver
}

func Load(confUrl, variablesFile, stateDir, imageServerAddress string,
	imageRebuildInterval time.Duration, slaveDriver *slavedriver.SlaveDriver,
	logger log.DebugLogger) (*Builder, error) {
	return LoadWithOptionsAndParams(
		BuilderOptions{
			ConfigurationURL:     confUrl,
			ImageRebuildInterval: imageRebuildInterval,
			ImageServerAddress:   imageServerAddress,
			StateDirectory:       stateDir,
			VariablesFile:        variablesFile,
		},
		BuilderParams{
			Logger:      logger,
			SlaveDriver: slaveDriver,
		})
}

func LoadWithOptionsAndParams(options BuilderOptions,
	params BuilderParams) (*Builder, error) {
	return load(options, params)
}

func (b *Builder) BuildImage(request proto.BuildImageRequest,
	authInfo *srpc.AuthInformation,
	logWriter io.Writer) (*image.Image, string, error) {
	return b.buildImage(request, authInfo, logWriter)
}

func (b *Builder) GetCurrentBuildLog(streamName string) ([]byte, error) {
	return b.getCurrentBuildLog(streamName)
}

func (b *Builder) GetDirectedGraph(request proto.GetDirectedGraphRequest) (
	proto.GetDirectedGraphResult, error) {
	return b.getDirectedGraph(request)
}

func (b *Builder) GetDependencies(request proto.GetDependenciesRequest) (
	proto.GetDependenciesResult, error) {
	return b.getDependencies(request)
}

func (b *Builder) GetLatestBuildLog(streamName string) ([]byte, error) {
	return b.getLatestBuildLog(streamName)
}

func (b *Builder) GetRelationshipsQuickLinks() ([]WebLink, error) {
	return b.relationshipsQuickLinks, nil
}

func (b *Builder) ReplaceIdleSlaves(immediateGetNew bool) error {
	return b.replaceIdleSlaves(immediateGetNew)
}

func (b *Builder) ShowImageStream(writer io.Writer, streamName string) {
	b.showImageStream(writer, streamName)
}

func (b *Builder) ShowImageStreams(writer io.Writer) {
	b.showImageStreams(writer)
}

func (b *Builder) WaitForStreamsLoaded(timeout time.Duration) error {
	return b.waitForStreamsLoaded(timeout)
}

func (b *Builder) WriteHtml(writer io.Writer) {
	b.writeHtml(writer)
}

func BuildImageFromManifest(client *srpc.Client, manifestDir, streamName string,
	expiresIn time.Duration, bindMounts []string, buildLog buildLogger,
	logger log.Logger) (
	string, error) {
	return BuildImageFromManifestWithOptions(
		client,
		BuildLocalOptions{
			BindMounts:        bindMounts,
			ManifestDirectory: manifestDir,
		},
		streamName,
		expiresIn,
		buildLog)
}

func BuildImageFromManifestWithOptions(client *srpc.Client,
	options BuildLocalOptions, streamName string, expiresIn time.Duration,
	buildLog buildLogger) (string, error) {
	_, name, err := buildImageFromManifestAndUpload(client, options, streamName,
		expiresIn, buildLog)
	return name, err
}

func BuildTreeFromManifest(client *srpc.Client, manifestDir string,
	bindMounts []string, buildLog io.Writer,
	logger log.Logger) (string, error) {
	return BuildTreeFromManifestWithOptions(
		client,
		BuildLocalOptions{
			BindMounts:        bindMounts,
			ManifestDirectory: manifestDir,
		},
		buildLog)
}

func BuildTreeFromManifestWithOptions(client *srpc.Client,
	options BuildLocalOptions, buildLog io.Writer) (string, error) {
	return buildTreeFromManifest(client, options, buildLog)
}

func ProcessManifest(manifestDir, rootDir string, bindMounts []string,
	buildLog io.Writer) error {
	return processManifest(manifestDir, rootDir, bindMounts, nil, buildLog)
}

func ProcessManifestWithOptions(options BuildLocalOptions,
	rootDir string, buildLog io.Writer) error {
	return processManifest(options.ManifestDirectory, rootDir,
		options.BindMounts, variablesGetter(options.Variables), buildLog)
}

func UnpackImageAndProcessManifest(client *srpc.Client, manifestDir string,
	rootDir string, bindMounts []string, buildLog io.Writer) error {
	_, err := unpackImageAndProcessManifest(client, manifestDir, 0, rootDir,
		bindMounts, true, nil, buildLog)
	return err
}

func UnpackImageAndProcessManifestWithOptions(client *srpc.Client,
	options BuildLocalOptions, rootDir string, buildLog io.Writer) error {
	_, err := unpackImageAndProcessManifest(client,
		options.ManifestDirectory, 0, rootDir, options.BindMounts, true,
		variablesGetter(options.Variables), buildLog)
	return err
}
