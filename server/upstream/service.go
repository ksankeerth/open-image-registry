package upstream

import (
	"github.com/ksankeerth/open-image-registry/db"
	db_errors "github.com/ksankeerth/open-image-registry/errors/db"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/types/models"
)

type upstreamService struct {
	upstreamDao db.UpstreamDAO
	adapter     *UpstreamAdapter
}

func (svc *upstreamService) createUpstreamRegistry(req *mgmt.CreateUpstreamRegistryRequest) (registryId,
	registryName string, err error) {
	upstreamRegistry := svc.adapter.ToUpstreamEntity(req)
	authConfig := svc.adapter.ToUpstreamAuthConfig(req)
	accessConfig := svc.adapter.ToUpstreamAccessConfig(req)
	storageConfig := svc.adapter.ToUpstreamStorageConfig(req)
	cacheConfig := svc.adapter.ToUpstreamCacheConfig(req)
	registryId, registryName, err = svc.upstreamDao.CreateUpstreamRegistry(upstreamRegistry, authConfig, accessConfig,
		storageConfig, cacheConfig, "")
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when creating upstream registry: %s", upstreamRegistry.Name)
	}



	return
}

func (svc *upstreamService) updateUpstreamRegistry(req *mgmt.UpdateUpstreamRegistryRequest) error {
	upstreamRegistry := svc.adapter.ToUpstreamEntity(&req.CreateUpstreamRegistryRequest)
	authConfig := svc.adapter.ToUpstreamAuthConfig(&req.CreateUpstreamRegistryRequest)
	accessConfig := svc.adapter.ToUpstreamAccessConfig(&req.CreateUpstreamRegistryRequest)
	storageConfig := svc.adapter.ToUpstreamStorageConfig(&req.CreateUpstreamRegistryRequest)
	cacheConfig := svc.adapter.ToUpstreamCacheConfig(&req.CreateUpstreamRegistryRequest)

	err := svc.upstreamDao.UpdateUpstreamRegistry(req.RegId, upstreamRegistry, authConfig, accessConfig,
		storageConfig, cacheConfig)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when updating registry: (%s/%s)", req.RegId, req.Name)
	}
	return err
}

func (svc *upstreamService) deleteUpstreamRegistry(registryId string) (notFound bool, err error) {
	err = svc.upstreamDao.DeleteUpstreamRegistry(registryId)
	if db_errors.IsNotFound(err) {
		return true, nil
	}
	return
}

func (svc *upstreamService) getUpstreamRegistryWithConfig(registryId string) (*models.UpstreamRegistryEntity,
	*models.UpstreamRegistryAccessConfig,
	*models.UpstreamRegistryAuthConfig,
	*models.UpstreamRegistryCacheConfig,
	*models.UpstreamRegistryStorageConfig,
	error) {
	return svc.upstreamDao.GetUpstreamRegistryWithConfig(registryId)
}

// TODO: Currently, it lists all registeries but in future we may have to introduce filtering and pagination.
func (svc *upstreamService) listUpstreamRegisteries() (total, limit, page int,
	registeries []*mgmt.UpstreamRegistrySummaryDTO, err error) {
	regModels, err := svc.upstreamDao.ListUpstreamRegistries()
	if err != nil {
		return
	}

	registeries = []*mgmt.UpstreamRegistrySummaryDTO{}
	for _, model := range regModels {
		registeries = append(registeries, svc.adapter.ToUpstreamRegistrySummaryDTO(model))
	}

	return len(registeries), len(registeries), 1, registeries, nil
}