package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/ksankeerth/open-image-registry/errors/db"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/types"
)

type imageRegistryDaoImpl struct {
	db *sql.DB
	*TransactionManager
}

func (i *imageRegistryDaoImpl) CheckImageManifestExists(manifestUniqDigest, namespace,
	repository, regId string, txKey string) (bool, error) {
	var row *sql.Row
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return false, db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(CheckImageManifestExistense, manifestUniqDigest, regId, namespace, repository)
	} else {
		row = i.db.QueryRow(CheckImageManifestExistense, manifestUniqDigest, regId, namespace, repository)
	}
	var res any
	err := row.Scan(&res)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when checking manifest existense in database")
		return false, db.ClassifyError(err, CheckImageManifestExistense)
	}
	return true, nil
}

// This function is not use-ful since we support multi-platform images.
// func (i *imageRegistryDaoImpl) CheckImageManifestExistsByTag(tag, namespace, repository,
// 	regId string, txKey string) (bool, error) {
// 	var row *sql.Row
// 	if txKey != "" {
// 		tx := i.getTx(txKey)
// 		if tx == nil {
// 			return false, db.ErrTxAlreadyClosed
// 		}
// 		row = tx.QueryRow(CheckImageManifestExistenseByTag, tag, regId, namespace, repository)
// 	} else {
// 		row = i.db.QueryRow(CheckImageManifestExistenseByTag, tag, regId, namespace, repository)
// 	}
// 	var res any
// 	err := row.Scan(&res)
// 	if errors.Is(err, sql.ErrNoRows) {
// 		return false, nil
// 	}
// 	if err != nil {
// 		log.Logger().Error().Err(err).Msgf("Error occured when checking manifest existense by tag in database")
// 		return false, db.ClassifyError(err, CheckImageManifestExistense)
// 	}
// 	return false, nil
// }

func (i *imageRegistryDaoImpl) CheckImageBlobExists(digest, namespace, repository,
	regId string, txKey string) (bool, error) {
	var row *sql.Row
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return false, db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(CheckImageBlobExistense, regId, namespace, repository, digest)
	} else {
		row = i.db.QueryRow(CheckImageBlobExistense, regId, namespace, repository, digest)
	}
	var res any
	err := row.Scan(&res)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when checking image blob existence")
		return false, db.ClassifyError(err, CheckImageBlobExistense)
	}

	return true, nil
}

func (i *imageRegistryDaoImpl) CheckImageBlobExistsByIds(digest, namespaceId, repositoryId,
	regId string, txKey string) (bool, error) {
	var row *sql.Row
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return false, db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(CheckImageBlobExistsByIds, regId, namespaceId, repositoryId, digest)
	} else {
		row = i.db.QueryRow(CheckImageBlobExistsByIds, regId, namespaceId, repositoryId, digest)
	}
	var res any
	err := row.Scan(&res)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when checking image blob existence by ids")
		return false, db.ClassifyError(err, CheckImageBlobExistsByIds)
	}
	return true, nil
}

func (i *imageRegistryDaoImpl) GetImageBlobStorageLocationAndSize(digest, namespace, repository,
	regId string, txKey string) (storageLocation string, size int, err error) {
	var row *sql.Row
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return "", -1, db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(GetImageBlobStorageLocationAndSize, regId, namespace, repository, digest)
	} else {
		row = i.db.QueryRow(GetImageBlobStorageLocationAndSize, regId, namespace, repository, digest)
	}
	err = row.Scan(&storageLocation, &size)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Logger().Error().Err(err).Msgf("Error occured when retriving storage location and size of image blob from database: %s:%s:%s:%s", regId, namespace, repository, digest)
	}
	if err != nil {
		return "", -1, db.ClassifyError(err, GetImageBlobStorageLocationAndSize)
	}
	return
}

func (i *imageRegistryDaoImpl) GetImageManifestsByTag(tag, namespace, repository, regId string,
	txKey string) ([]*types.ImageManifestWithPlatform, error) {
	var rows *sql.Rows
	var err error
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return []*types.ImageManifestWithPlatform{}, db.ErrTxAlreadyClosed
		}
		rows, err = tx.Query(GetImageManifestsExistByTagWithPlatformAttributes, regId, namespace, repository, tag)
	} else {
		rows, err = i.db.Query(GetImageManifestsExistByTagWithPlatformAttributes, regId, namespace, repository, tag)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return []*types.ImageManifestWithPlatform{}, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when retriving image manifests by tag")
		return nil, db.ClassifyError(err, GetImageManifestsExistByTagWithPlatformAttributes)
	}
	var res []*types.ImageManifestWithPlatform

	for rows.Next() {
		var entry types.ImageManifestWithPlatform
		err = rows.Scan(&entry.OS, &entry.Arch, &entry.Digest, &entry.Size, &entry.MediaType)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occured when reading results from database")
			return nil, db.ClassifyError(err, GetImageManifestsExistByTagWithPlatformAttributes)
		}
		res = append(res, &entry)
	}

	return res, nil
}

func (i *imageRegistryDaoImpl) GetImageManifestByDigest(digest, namespace, repository,
	regId string, txKey string) (content []byte, mediaType string, size int, err error) {
	var row *sql.Row
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return []byte{}, "", -1, db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(GetImageManifestByDigest, regId, namespace, repository, digest)
	} else {
		row = i.db.QueryRow(GetImageManifestByDigest, regId, namespace, repository, digest)
	}
	err = row.Scan(&content, &mediaType, &size)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to retrieve manifest by digest: %s", digest)
		return []byte{}, "", -1, db.ClassifyError(err, GetImageManifestByDigest)
	}

	return content, mediaType, size, nil
}

func (i *imageRegistryDaoImpl) CreateNamespaceAndRepositoryIfNotExist(regId, namespace,
	repository string, txKey string) (namespaceId, repositoryId string, err error) {
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return "", "", db.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(CheckNamespaceExists, regId, namespace).Scan(&namespaceId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Logger().Error().Err(err).Msgf("Unable to check namespace(%s:%s) existence ", regId, namespace)
			return "", "", db.ClassifyError(err, CheckNamespaceExists)
		}

		if namespaceId == "" {
			err = tx.QueryRow(InsertNamespace, regId, namespace).Scan(&namespaceId)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Unable to create namespace(%s:%s)", regId, namespace)
				return "", "", db.ClassifyError(err, CheckNamespaceExists)
			}
		}

		err = tx.QueryRow(CheckImageRepositoryExists, regId, namespaceId, repository).Scan(&repositoryId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Logger().Error().Err(err).Msgf("Unable to check image repository(%s:%s) existence ", regId, repository)
			return "", "", db.ClassifyError(err, CheckImageRepositoryExists)
		}

		if repositoryId == "" {
			err = tx.QueryRow(InsertImageRepository, regId, namespaceId, repository).Scan(&repositoryId)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Unable to create image repository(%s:%s)", regId, namespace)
				return "", "", db.ClassifyError(err, CheckNamespaceExists)
			}
		}
		return namespaceId, repositoryId, nil
	} else {
		err = i.db.QueryRow(CheckNamespaceExists, regId, namespace).Scan(&namespaceId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Logger().Error().Err(err).Msgf("Unable to check namespace(%s:%s) existence ", regId, namespace)
			return "", "", db.ClassifyError(err, CheckNamespaceExists)
		}

		if namespaceId == "" {
			err = i.db.QueryRow(InsertNamespace, regId, namespace).Scan(&namespaceId)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Unable to create namespace(%s:%s)", regId, namespace)
				return "", "", db.ClassifyError(err, CheckNamespaceExists)
			}
		}

		err = i.db.QueryRow(CheckImageRepositoryExists, regId, namespaceId, repository).Scan(&repositoryId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Logger().Error().Err(err).Msgf("Unable to check image repository(%s:%s) existence ", regId, repository)
			return "", "", db.ClassifyError(err, CheckImageRepositoryExists)
		}

		if repositoryId == "" {
			err = i.db.QueryRow(InsertImageRepository, regId, namespaceId, repository).Scan(&repositoryId)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Unable to create image repository(%s:%s)", regId, namespace)
				return "", "", db.ClassifyError(err, CheckNamespaceExists)
			}
		}
		return namespaceId, repositoryId, nil
	}
}

func (i *imageRegistryDaoImpl) GetBlobUploadSession(sessionId string,
	txKey string) (id string, isChunked bool, lastUpdated time.Time, err error) {
	var row *sql.Row
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return "", false, time.Now(), db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(GetImageBlobUploadSession, sessionId)
	} else {
		row = i.db.QueryRow(GetImageBlobUploadSession, sessionId)
	}

	var chunked int
	err = row.Scan(&chunked, &lastUpdated)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to read result from query: (%s, %s)", GetImageBlobUploadSession, sessionId)
		return "", false, time.Now(), db.ClassifyError(err, GetImageBlobUploadSession)
	}
	if chunked == 1 {
		isChunked = true
	}
	return sessionId, isChunked, lastUpdated, nil
}

func (i *imageRegistryDaoImpl) CreateNewBlobUploadSession(registryId, namespaceId,
	repositoryId, sessionId string, txKey string) error {
	var res sql.Result
	var err error
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}
		// During the creation of blob, We won't have blobdigest. Therefore, sessionId will
		// be stored as blobdigest
		res, err = tx.Exec(PersistImageBlobUploadSession, registryId, namespaceId, repositoryId, sessionId, sessionId, 0)
	} else {
		res, err = i.db.Exec(PersistImageBlobUploadSession, registryId, namespaceId, repositoryId, sessionId, sessionId, 0)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when persisting image blob upload session: %s", sessionId)
		return db.ClassifyError(err, PersistImageBlobUploadSession)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		log.Logger().Error().Err(err).Msgf("Failed to persist image blob upload session: %s", sessionId)
		return db.ClassifyError(err, PersistImageBlobUploadSession)
	}
	return nil
}

func (i *imageRegistryDaoImpl) DeleteBlobUploadSession(sessionId string, txKey string) error {
	var res sql.Result
	var err error

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(DeleteBlobUploadSession, sessionId)
	} else {
		res, err = i.db.Exec(DeleteBlobUploadSession, sessionId)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to remove blob upload session(%s)", sessionId)
		return db.ClassifyError(err, DeleteBlobUploadSession)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		log.Logger().Error().Err(err).Msgf("Deleting IMAGE_BLOB_UPLOAD_SESSION(%s) was not successful", sessionId)
		return db.ClassifyError(err, DeleteBlobUploadSession)
	}

	return nil
}

func (i *imageRegistryDaoImpl) MarkBlobUploadSessionAsChunked(sessionId string, txKey string) error {
	var res sql.Result
	var err error
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(MarkBlobUploadSessionAsChunked, sessionId)
	} else {
		res, err = i.db.Exec(MarkBlobUploadSessionAsChunked, sessionId)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to mark session as chunked.")
		return db.ClassifyError(err, MarkBlobUploadSessionAsChunked)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		log.Logger().Error().Err(err).Msgf("Failed to mark session(%s) as chunked", sessionId)
		return db.ClassifyError(err, MarkBlobUploadSessionAsChunked)
	}
	return nil
}

func (i *imageRegistryDaoImpl) PersistImageBlobMetaInfo(registryId, namespaceId, repositoryId,
	digest, location, mediaType string, size int64, txKey string) error {
	var res sql.Result
	var err error
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(PersistImageBlobMeta, registryId, namespaceId,
			repositoryId, digest, mediaType, size, location)
	} else {
		res, err = i.db.Exec(PersistImageBlobMeta, registryId, namespaceId,
			repositoryId, digest, mediaType, size, location)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when inserting a row into IMAGE_BLOB_META")
		return db.ClassifyError(err, PersistImageBlobMeta)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		log.Logger().Error().Err(err).Msgf("Failed to insert IMAGE_BLOB_META")
		return db.ClassifyError(err, PersistImageBlobMeta)
	}
	return nil
}

func (i *imageRegistryDaoImpl) GetNamespaceId(regId string, namespace string,
	txKey string) (string, error) {
	var row *sql.Row
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return "", db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(GetNamespaceID, regId, namespace)
	} else {
		row = i.db.QueryRow(GetNamespaceID, regId, namespace)
	}
	var nsId string
	err := row.Scan(&nsId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when reading results from database")
		return "", db.ClassifyError(err, GetNamespaceID)
	}
	return nsId, nil
}

func (i *imageRegistryDaoImpl) GetRepositoryId(regId, namespaceId, repository string,
	txKey string) (string, error) {
	var row *sql.Row
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return "", db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(GetRepositoryID, regId, namespaceId, repository)
	} else {
		row = i.db.QueryRow(GetRepositoryID, regId, namespaceId, repository)
	}
	var repoId string
	err := row.Scan(&repoId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when reading results from database")
		return "", db.ClassifyError(err, GetRepositoryID)
	}
	return repoId, nil
}

func (i *imageRegistryDaoImpl) GetImageTagId(regId, namespaceId, repositoryId,
	tag, txKey string) (tagId string, err error) {

	var row *sql.Row

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return "", db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(GetImageTagId, regId, namespaceId, repositoryId, tag)
	} else {
		row = i.db.QueryRow(GetImageTagId, regId, namespaceId, repositoryId, tag)
	}
	err = row.Scan(&tagId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Logger().Info().Msgf("No IMAGE_TAG found for (%s:%s:%s:%s)", regId, namespaceId, repositoryId, tag)
		} else {
			log.Logger().Error().Err(err).Msgf("Unable to retrieve TAG_ID from table IMAGE_TAG")
		}
		return "", db.ClassifyError(err, GetImageTagId)
	}
	return tagId, nil
}

func (i *imageRegistryDaoImpl) CheckImageTagAndConfigDigestMapping(configDigest, tagId,
	txKey string) (exists bool, err error) {
	var row *sql.Row
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return false, db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(CheckImageTagConfigMapping, tagId, configDigest)
	} else {
		row = i.db.QueryRow(CheckImageTagConfigMapping, tagId, configDigest)
	}

	var res any
	err = row.Scan(&res)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to retrieve image tag(%s) and manifest(%s) mapping", tagId, configDigest)
		return false, db.ClassifyError(err, CheckImageTagConfigMapping)
	}
	return true, nil
}

func (i *imageRegistryDaoImpl) LinkImageConfigWithTag(configDigest, tagId, txKey string) error {
	var res sql.Result
	var err error
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(LinkImageConfigDigestWithImageTag, tagId, configDigest)
	} else {
		res, err = i.db.Exec(LinkImageConfigDigestWithImageTag, tagId, configDigest)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to update IMAGE_TAG(%s) with manifest digest(%s)", tagId, configDigest)
		return db.ClassifyError(err, LinkImageConfigDigestWithImageTag)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		log.Logger().Error().Msgf("Failed to updated IMAGE_TAG(%s) with manifest digest(%s)", tagId, configDigest)
		return db.ClassifyError(err, LinkImageConfigDigestWithImageTag)
	}
	return nil
}

func (i *imageRegistryDaoImpl) CreateImageTag(regId, namespaceId, repositoryId, tag string,
	txKey string) (tagId string, err error) {
	var row *sql.Row

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return "", db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(PersistImageTag, regId, namespaceId, repositoryId, tag)
	} else {
		row = i.db.QueryRow(PersistImageTag, regId, namespaceId, repositoryId, tag)
	}

	err = row.Scan(&tagId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when persisting image tag")
		return "", db.ClassifyError(err, PersistImageTag)
	}

	return tagId, nil
}

func (i *imageRegistryDaoImpl) MarkImageBlobAsConfig(regId, namespaceId, repositoryId,
	blobDigest, txKey string) error {
	var res sql.Result
	var err error

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(UpdateImageBlobMediaType, "application/vnd.docker.container.image.v1+json", regId, namespaceId, repositoryId, blobDigest)
	} else {
		res, err = i.db.Exec(UpdateImageBlobMediaType, "application/vnd.docker.container.image.v1+json", regId, namespaceId, repositoryId, blobDigest)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to update IMAGE_BLOB_META due to db errors for (%s:%s:%s:%s)", regId, namespaceId, repositoryId, blobDigest)
		return db.ClassifyError(err, UpdateImageBlobMediaType)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		log.Logger().Error().Err(err).Msgf("Update to IMAGE_BLOB_META failed for for (%s:%s:%s:%s)", regId, namespaceId, repositoryId, blobDigest)
		return db.ClassifyError(err, UpdateImageBlobMediaType)
	}

	return nil
}

func (i *imageRegistryDaoImpl) LinkImageTagAndLayer(regId, namespaceId, repositoryId, tagId, layerDigest,
	configDigest string, index int, txKey string) error {
	var res sql.Result
	var err error
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(InsertImageTagLayerMapping, regId, namespaceId, repositoryId, tagId, layerDigest, configDigest, index)
	} else {
		res, err = i.db.Exec(InsertImageTagLayerMapping, regId, namespaceId, repositoryId, tagId, layerDigest, configDigest, index)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to insert Image Tag Layer mapping into the database")
		return db.ClassifyError(err, InsertImageTagLayerMapping)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		log.Logger().Error().Err(err).Msgf("Image Tag Layer mapping was not persisted successfully.")
		return db.ClassifyError(err, InsertImageTagLayerMapping)
	}
	return nil
}

func (i *imageRegistryDaoImpl) CheckImageTagAndLayerMappingExistence(tagId, layerDigest string,
	layerIndex int, txKey string) (bool, error) {
	var row *sql.Row
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return false, db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(CheckImageTagLayerMappingExistence, tagId, layerDigest, layerIndex)
	} else {
		row = i.db.QueryRow(CheckImageTagLayerMappingExistence, tagId, layerDigest, layerIndex)
	}
	var res any
	err := row.Scan(&res)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when checking image tag and layer mapping: %s, %s", tagId, layerDigest)
		return false, db.ClassifyError(err, CheckImageTagLayerMappingExistence)
	}
	return true, nil
}

func (i *imageRegistryDaoImpl) CheckImagePlatformConfigExistence(tagId, configDigest string,
	txKey string) (bool, error) {
	var row *sql.Row
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return false, db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(CheckImagePlatformConfigExistence, tagId, configDigest)
	} else {
		row = i.db.QueryRow(CheckImagePlatformConfigExistence, tagId, configDigest)
	}
	var res any
	err := row.Scan(&res)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when checking image platform config : %s, %s", tagId, configDigest)
		return false, db.ClassifyError(err, CheckImagePlatformConfigExistence)
	}
	return true, nil
}

func (i *imageRegistryDaoImpl) GetImageBlobStorageLocationAndSizeByIds(digest, namespaceId,
	repositoryId, regId string, txKey string) (storageLocation string, size int, err error) {
	var row *sql.Row
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return "", -1, db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(GetImageBlobStorageLocationAndSizeByIds, regId, namespaceId, repositoryId, digest)
	} else {
		row = i.db.QueryRow(GetImageBlobStorageLocationAndSizeByIds, regId, namespaceId, repositoryId, digest)
	}

	err = row.Scan(&storageLocation, &size)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to retrive data from IMAGE_BLOB_META due to errors")
		return "", -1, db.ClassifyError(err, GetImageBlobStorageLocationAndSizeByIds)
	}
	return storageLocation, size, nil
}

func (i *imageRegistryDaoImpl) PersistImagePlatformConfig(regId, namespaceId, repositoryId,
	tagId, configDigest, os, arch string, props []byte, txKey string) error {
	var res sql.Result
	var err error

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}

		res, err = tx.Exec(PersistImagePlatformConfig, regId, namespaceId, repositoryId, tagId, configDigest, os, arch, props)
	} else {
		res, err = i.db.Exec(PersistImagePlatformConfig, regId, namespaceId, repositoryId, tagId, configDigest, os, arch, props)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to persist image platform config with digest(%s)", configDigest)
		return db.ClassifyError(err, PersistImagePlatformConfig)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		log.Logger().Error().Err(err).Msgf("Unable to read sql result of insertion of image platform config")
		return db.ClassifyError(err, PersistImagePlatformConfig)
	}

	return nil
}

func (i *imageRegistryDaoImpl) PersistImageManifest(regId, namespaceId, repositoryId,
	manifestDigest, imageConfigDigest, mediaType, uniqueDigest string, size int64, content []byte, txKey string) error {

	var res sql.Result
	var err error

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(PersistImageManifest, regId, namespaceId, repositoryId,
			manifestDigest, imageConfigDigest, mediaType, size, content, uniqueDigest)

	} else {
		res, err = i.db.Exec(PersistImageManifest, regId, namespaceId, repositoryId,
			manifestDigest, imageConfigDigest, mediaType, size, content, uniqueDigest)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to persist image manifest with digest(%s)", manifestDigest)
		return db.ClassifyError(err, PersistImageManifest)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		log.Logger().Error().Err(err).Msgf("Unable to read sql result of insertion of image manifest")
		return db.ClassifyError(err, PersistImageManifest)
	}

	return nil
}

func (i *imageRegistryDaoImpl) CheckImageManifestsExistByTag(regId, namespace,
	repository, tag, txKey string) (bool, error) {
	var rows *sql.Rows
	var err error
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return false, db.ErrTxAlreadyClosed
		}
		rows, err = tx.Query(CheckImageManifestsExistByTag, regId, namespace, repository, tag)
	} else {
		rows, err = i.db.Query(CheckImageManifestsExistByTag, regId, namespace, repository, tag)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when checking manifests by tag: %s:%s:%s:%s", regId, namespace, repository, tag)
		return false, db.ClassifyError(err, CheckImageManifestsExistByTag)
	}
	return rows.Next(), nil
}