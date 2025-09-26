package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/ksankeerth/open-image-registry/errors/db"
	"github.com/ksankeerth/open-image-registry/log"
)

type imageRegistryDaoImpl struct {
	db *sql.DB
	*TransactionManager
}

func (i *imageRegistryDaoImpl) CheckImageManifestExistsByTag(registryId, namespaceId, repositoryId, tag string,
	txKey string) (exists bool, digest, mediaType string, err error) {
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return false, "", "", db.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(CheckImageManifestByTag, registryId, namespaceId, repositoryId, tag).Scan(&digest, &mediaType)
	} else {
		err = i.db.QueryRow(CheckImageManifestByTag, registryId, namespaceId, repositoryId, tag).Scan(&digest, &mediaType)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", "", nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when retriving image manifest for: %s/%s/%s:%s", registryId, namespaceId, repositoryId, tag)
		return false, "", "", db.ClassifyError(err, CheckImageManifestByTag)
	}
	exists = true
	return
}

func (i *imageRegistryDaoImpl) GetImageManifestByTag(registryId, namespaceId, repositoryId,
	tag string, txKey string) (exists bool, content []byte, digest, mediaType string, err error) {
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return false, nil, "", "", db.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(GetImageManifestByTag, registryId, namespaceId, repositoryId, tag).
			Scan(&digest, &mediaType, &content)
	} else {
		err = i.db.QueryRow(GetImageManifestByTag, registryId, namespaceId, repositoryId, tag).
			Scan(&digest, &mediaType, &content)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil, "", "", nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when retriving image manifest for: %s/%s/%s:%s",
			registryId, namespaceId, repositoryId, tag)
		return false, nil, "", "", db.ClassifyError(err, GetImageManifestByTag)
	}
	exists = true
	return
}

func (i *imageRegistryDaoImpl) CheckImageManifestExistsByDigest(registryId, namespaceId, repositoryId,
	digest string, txKey string) (exists bool, mediaType string, err error) {
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return false, "", db.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(CheckImageManifestByDigest, registryId, namespaceId, repositoryId, digest).Scan(&mediaType)
	} else {
		err = i.db.QueryRow(CheckImageManifestByDigest, registryId, namespaceId, repositoryId, digest).Scan(&mediaType)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when retriving image manifest for: %s/%s/%s:%s",
			registryId, namespaceId, repositoryId, digest)
		return false, "", db.ClassifyError(err, CheckImageManifestByDigest)
	}
	exists = true
	return
}

func (i *imageRegistryDaoImpl) GetImageManifestByDigest(registryId, namespaceId, repositoryId, digest string,
	txKey string) (exists bool, content []byte, mediaType string, err error) {
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return false, nil, "", db.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(GetImageManifestByDigest, registryId, namespaceId, repositoryId, digest).Scan(&mediaType, &content)
	} else {
		err = i.db.QueryRow(GetImageManifestByDigest, registryId, namespaceId, repositoryId, digest).Scan(&mediaType, &content)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil, "", nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when retriving image manifest for: %s/%s/%s:%s",
			registryId, namespaceId, repositoryId, digest)
		return false, nil, "", db.ClassifyError(err, GetImageManifestByDigest)
	}
	exists = true
	return
}

func (i *imageRegistryDaoImpl) CheckImageManifestExistsByUniqueDigest(registryId, namespaceId, repositoryId,
	uniqueDigest string, txKey string) (exists bool, id string, err error) {
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return false, "", db.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(CheckImageManifestByUniqueDigest, registryId, namespaceId, repositoryId, uniqueDigest).Scan(&id)
	} else {
		err = i.db.QueryRow(CheckImageManifestByUniqueDigest, registryId, namespaceId, repositoryId, uniqueDigest).Scan(&id)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when retriving image manifest for: %s/%s/%s:%s",
			registryId, namespaceId, repositoryId, uniqueDigest)
		return false, "", db.ClassifyError(err, CheckImageManifestByUniqueDigest)
	}
	exists = true
	return
}

func (i *imageRegistryDaoImpl) PersistImageManifest(regId, namespaceId, repositoryId,
	manifestDigest, mediaType, uniqueDigest string, size int64, content []byte, txKey string) (string, error) {

	var row *sql.Row

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return "", db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(InsertImageManifest, regId, namespaceId, repositoryId,
			manifestDigest, mediaType, size, content, uniqueDigest)

	} else {
		row = i.db.QueryRow(InsertImageManifest, regId, namespaceId, repositoryId,
			manifestDigest, mediaType, size, content, uniqueDigest)
	}
	var manifestId string
	err := row.Scan(&manifestId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to persist image manifest with digest(%s)", manifestDigest)
		return "", db.ClassifyError(err, InsertImageManifest)
	}

	return manifestId, nil
}

func (i *imageRegistryDaoImpl) CreateNamespaceAndRepositoryIfNotExist(regId, namespace,
	repository string, txKey string) (namespaceId, repositoryId string, err error) {
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return "", "", db.ErrTxAlreadyClosed
		}
		err = tx.QueryRow(GetNamespaceIdByName, regId, namespace).Scan(&namespaceId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Logger().Error().Err(err).Msgf("Unable to check namespace(%s:%s) existence ", regId, namespace)
			return "", "", db.ClassifyError(err, GetNamespaceIdByName)
		}

		if namespaceId == "" {
			err = tx.QueryRow(InsertNamespace, regId, namespace).Scan(&namespaceId)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Unable to create namespace(%s:%s)", regId, namespace)
				return "", "", db.ClassifyError(err, InsertNamespace)
			}
		}

		err = tx.QueryRow(GetRepositoryIdByName, regId, namespaceId, repository).Scan(&repositoryId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Logger().Error().Err(err).Msgf("Unable to check image repository(%s:%s) existence ", regId, repository)
			return "", "", db.ClassifyError(err, GetRepositoryIdByName)
		}

		if repositoryId == "" {
			err = tx.QueryRow(InsertImageRepository, regId, namespaceId, repository).Scan(&repositoryId)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Unable to create image repository(%s:%s)", regId, namespace)
				return "", "", db.ClassifyError(err, InsertImageRepository)
			}
		}
		return namespaceId, repositoryId, nil
	} else {
		err = i.db.QueryRow(GetNamespaceIdByName, regId, namespace).Scan(&namespaceId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Logger().Error().Err(err).Msgf("Unable to check namespace(%s:%s) existence ", regId, namespace)
			return "", "", db.ClassifyError(err, GetNamespaceIdByName)
		}

		if namespaceId == "" {
			err = i.db.QueryRow(InsertNamespace, regId, namespace).Scan(&namespaceId)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Unable to create namespace(%s:%s)", regId, namespace)
				return "", "", db.ClassifyError(err, InsertNamespace)
			}
		}

		err = i.db.QueryRow(GetRepositoryIdByName, regId, namespaceId, repository).Scan(&repositoryId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Logger().Error().Err(err).Msgf("Unable to check image repository(%s:%s) existence ", regId, repository)
			return "", "", db.ClassifyError(err, GetRepositoryIdByName)
		}

		if repositoryId == "" {
			err = i.db.QueryRow(InsertImageRepository, regId, namespaceId, repository).Scan(&repositoryId)
			if err != nil {
				log.Logger().Error().Err(err).Msgf("Unable to create image repository(%s:%s)", regId, namespace)
				return "", "", db.ClassifyError(err, InsertImageRepository)
			}
		}
		return namespaceId, repositoryId, nil
	}
}

func (i *imageRegistryDaoImpl) GetNamespaceId(regId string, namespace string,
	txKey string) (string, error) {
	var row *sql.Row
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return "", db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(GetNamespaceIdByName, regId, namespace)
	} else {
		row = i.db.QueryRow(GetNamespaceIdByName, regId, namespace)
	}
	var nsId string
	err := row.Scan(&nsId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when reading results from database")
		return "", db.ClassifyError(err, GetNamespaceIdByName)
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
		row = tx.QueryRow(GetRepositoryIdByName, regId, namespaceId, repository)
	} else {
		row = i.db.QueryRow(GetRepositoryIdByName, regId, namespaceId, repository)
	}
	var repoId string
	err := row.Scan(&repoId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when reading results from database")
		return "", db.ClassifyError(err, GetRepositoryIdByName)
	}
	return repoId, nil
}

func (i *imageRegistryDaoImpl) PersistImageBlobMetaInfo(registryId, namespaceId, repositoryId, digest, location string,
	size int64, txKey string) error {
	var res sql.Result
	var err error
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(InsertImageBlobMetaInfo, registryId, namespaceId,
			repositoryId, digest, size, location)
	} else {
		res, err = i.db.Exec(InsertImageBlobMetaInfo, registryId, namespaceId,
			repositoryId, digest, size, location)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when inserting a row into IMAGE_BLOB_META")
		return db.ClassifyError(err, InsertImageBlobMetaInfo)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		log.Logger().Error().Err(err).Msgf("Failed to insert IMAGE_BLOB_META")
		return db.ClassifyError(err, InsertImageBlobMetaInfo)
	}
	return nil
}

func (i *imageRegistryDaoImpl) GetImageBlobStorageLocationAndSize(digest, namespace, repository,
	regId string, txKey string) (storageLocation string, size int, err error) {
	var row *sql.Row
	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return "", -1, db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(GetImageBlobLocationAndSize, regId, namespace, repository, digest)
	} else {
		row = i.db.QueryRow(GetImageBlobLocationAndSize, regId, namespace, repository, digest)
	}
	err = row.Scan(&storageLocation, &size)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Logger().Error().Err(err).Msgf("Error occured when retriving storage location and size of image blob from database: %s:%s:%s:%s", regId, namespace, repository, digest)
	}
	if err != nil {
		return "", -1, db.ClassifyError(err, GetImageBlobLocationAndSize)
	}
	return
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

func (i *imageRegistryDaoImpl) CreateImageTag(regId, namespaceId, repositoryId, tag string,
	txKey string) (tagId string, err error) {
	var row *sql.Row

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return "", db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(InsertImageTag, regId, namespaceId, repositoryId, tag)
	} else {
		row = i.db.QueryRow(InsertImageTag, regId, namespaceId, repositoryId, tag)
	}

	err = row.Scan(&tagId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occured when persisting image tag")
		return "", db.ClassifyError(err, InsertImageTag)
	}

	return tagId, nil
}

func (i *imageRegistryDaoImpl) CacheImageManifestReference(regId, namespaceId, repositoryId,
	identifier string, expiresAt time.Time, txKey string) error {
	var res sql.Result
	var err error

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(InsertIntoImageRegistryCache, regId, namespaceId, repositoryId, identifier, expiresAt)
	} else {
		res, err = i.db.Exec(InsertIntoImageRegistryCache, regId, namespaceId, repositoryId, identifier, expiresAt)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to add cache entry(%s/%s/%s:%s)",
			regId, namespaceId, repositoryId, identifier)
		return db.ClassifyError(err, InsertIntoImageRegistryCache)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		log.Logger().Error().Err(err).Msgf("Unable to read sql result of insertion of image registry cache entry")
		return db.ClassifyError(err, InsertIntoImageRegistryCache)
	}

	return nil
}
func (i *imageRegistryDaoImpl) GetCacheImageManifestReference(regId, namespaceId, repositoryId,
	identifier string, txKey string) (cacheMiss bool, digest string, expiresAt time.Time, err error) {

	var row *sql.Row

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return false, "", time.Now(), db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(GetEntryFromImageRegistryCache, regId, namespaceId, repositoryId, identifier)
	} else {
		row = i.db.QueryRow(GetEntryFromImageRegistryCache, regId, namespaceId, repositoryId, identifier)
	}

	var expiryTimeTxt string
	err = row.Scan(&digest, &expiryTimeTxt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Logger().Info().Msgf("No Image Registry Cache entry found for (%s:%s:%s:%s)",
				regId, namespaceId, repositoryId, identifier)
		} else {
			log.Logger().Error().Err(err).Msgf("Unable to retrieve entry from ImageRegistryCache table")
		}
		return false, "", time.Now(), db.ClassifyError(err, GetEntryFromImageRegistryCache)
	}
	cacheMiss = false
	return
}
func (i *imageRegistryDaoImpl) DeleteCacheImageManifestReference(regId, namespaceId, repositoryId,
	identifier string, txKey string) error {
	var res sql.Result
	var err error

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(DeleteEntryFromImageRegistryCache, regId, namespaceId, repositoryId, identifier)
	} else {
		res, err = i.db.Exec(DeleteEntryFromImageRegistryCache, regId, namespaceId, repositoryId, identifier)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when removing entry from ImageRegistryCache")
		return db.ClassifyError(err, DeleteEntryFromImageRegistryCache)
	}

	r, err := res.RowsAffected()
	if err != nil {
		return db.ClassifyError(err, DeleteEntryFromImageRegistryCache)
	}
	if r != 1 {
		log.Logger().Warn().Msgf("No entry was deleted from ImageRegistryCache: %s/%s/%s:%s",
			regId, namespaceId, repositoryId, identifier)
		return nil
	}

	return nil
}
func (i *imageRegistryDaoImpl) RefreshCacheImageManifestReference(regId, namespaceId, repositoryId,
	identifier string, expiresAt time.Time, txKey string) error {
	var res sql.Result
	var err error

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(RefreshImageRegistryCacheEntry, expiresAt, regId, namespaceId, repositoryId, identifier)
	} else {
		res, err = i.db.Exec(RefreshImageRegistryCacheEntry, expiresAt, regId, namespaceId, repositoryId, identifier)
	}
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when removing entry from ImageRegistryCache")
		return db.ClassifyError(err, RefreshImageRegistryCacheEntry)
	}

	r, err := res.RowsAffected()
	if err != nil {
		return db.ClassifyError(err, RefreshImageRegistryCacheEntry)
	}
	if r != 1 {
		log.Logger().Warn().Msgf("No entry was updated in ImageRegistryCache for %s/%s/%s:%s",
			regId, namespaceId, repositoryId, identifier)
		return nil
	}

	return nil
}

func (i *imageRegistryDaoImpl) LinkImageManifestWithTag(tagId, manifestId string, txKey string) error {
	var res sql.Result
	var err error

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(InsertImageManifestTagMapping, manifestId, tagId)

	} else {
		res, err = i.db.Exec(InsertImageManifestTagMapping, manifestId, tagId)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to persist image manifest tag mapping (%s:%s)", tagId, manifestId)
		return db.ClassifyError(err, InsertImageManifestTagMapping)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		log.Logger().Error().Err(err).Msgf("Unable to persist image manifest tag mapping (%s:%s)", tagId, manifestId)
		return db.ClassifyError(err, InsertImageManifestTagMapping)
	}

	return nil
}

func (i *imageRegistryDaoImpl) UpdateManifestIdForTag(tagId string, newManifestId string, txKey string) error {
	var res sql.Result
	var err error

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return db.ErrTxAlreadyClosed
		}
		res, err = tx.Exec(UpdateImageManifestTagMappingWitNewManifestId, newManifestId, tagId)

	} else {
		res, err = i.db.Exec(UpdateImageManifestTagMappingWitNewManifestId, newManifestId, tagId)
	}

	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to update image manifest tag mapping (%s:%s)", tagId, newManifestId)
		return db.ClassifyError(err, UpdateImageManifestTagMappingWitNewManifestId)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		log.Logger().Error().Err(err).Msgf("Unable to update image manifest tag mapping (%s:%s)", tagId, newManifestId)
		return db.ClassifyError(err, UpdateImageManifestTagMappingWitNewManifestId)
	}

	return nil
}

func (i *imageRegistryDaoImpl) GetLinkedManifestByTagId(tagId string, txKey string) (manifestId string, err error) {
	var row *sql.Row

	if txKey != "" {
		tx := i.getTx(txKey)
		if tx == nil {
			return "", db.ErrTxAlreadyClosed
		}
		row = tx.QueryRow(GetManifestIdByTagId, tagId)
	} else {
		row = i.db.QueryRow(GetManifestIdByTagId, tagId)
	}

	err = row.Scan(&manifestId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Logger().Info().Msgf("No Image Manifest found for tagId(%s)", tagId)
		} else {
			log.Logger().Error().Err(err).Msgf("Unable to retrieve entry from Image table")
		}
		return "", db.ClassifyError(err, GetManifestIdByTagId)
	}
	return
}