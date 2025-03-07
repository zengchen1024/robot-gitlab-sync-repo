package obsimpl

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"

	dobs "github.com/opensourceways/robot-gitlab-sync-repo/domain/obs"
	"github.com/opensourceways/robot-gitlab-sync-repo/utils"
)

func NewOBS(cfg *Config) (dobs.OBS, error) {
	cli, err := obs.New(cfg.AccessKey, cfg.SecretKey, cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("new obs client failed, err:%s", err.Error())
	}

	_, err, _ = utils.RunCmd(
		cfg.OBSUtilPath, "config",
		"-i="+cfg.AccessKey, "-k="+cfg.SecretKey, "-e="+cfg.Endpoint,
	)
	if err != nil {
		return nil, fmt.Errorf("obsutil config failed, err:%s", err.Error())
	}

	return &obsImpl{
		obsClient: cli,
		bucket:    cfg.Bucket,
		obsutil:   cfg.OBSUtilPath,
	}, nil
}

type obsImpl struct {
	obsClient *obs.ObsClient
	bucket    string
	obsutil   string
}

func (s *obsImpl) SaveObject(path, content string) error {
	input := &obs.PutObjectInput{}
	input.Bucket = s.bucket
	input.Key = path
	input.Body = strings.NewReader(content)
	input.ContentMD5 = utils.GenMD5([]byte(content))

	_, err := s.obsClient.PutObject(input)

	return err
}

func (s *obsImpl) CopyObject(dst, src string) error {
	input := &obs.CopyObjectInput{}
	input.Bucket = s.bucket
	input.Key = dst
	input.CopySourceBucket = s.bucket
	input.CopySourceKey = src

	_, err := s.obsClient.CopyObject(input)

	return err
}

func (s *obsImpl) GetObject(path string) ([]byte, error) {
	input := &obs.GetObjectInput{}
	input.Bucket = s.bucket
	input.Key = path

	output, err := s.obsClient.GetObject(input)
	if err != nil {
		v, ok := err.(obs.ObsError)
		if ok && v.BaseModel.StatusCode == 404 {
			return nil, nil
		}

		return nil, err
	}

	v, err := ioutil.ReadAll(output.Body)

	output.Body.Close()

	return v, err
}

func (s *obsImpl) OBSUtilPath() string {
	return s.obsutil
}
