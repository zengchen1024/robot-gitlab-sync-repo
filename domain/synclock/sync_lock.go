package synclock

import (
	"github.com/opensourceways/robot-gitlab-sync-repo/domain"
)

type errorRepoNotExists struct {
	error
}

func NewErrorRepoNotExists(err error) errorRepoNotExists {
	return errorRepoNotExists{err}
}

func IsRepoSyncLockNotExist(err error) bool {
	_, ok := err.(errorRepoNotExists)

	return ok
}

type RepoSyncLock interface {
	Find(
		owner domain.Account,
		repoType domain.ResourceType, repoId string,
	) (domain.RepoSyncLock, error)

	Save(*domain.RepoSyncLock) (domain.RepoSyncLock, error)
}
