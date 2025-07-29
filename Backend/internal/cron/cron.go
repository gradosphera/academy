package cron

import (
	"academy/internal/model"
	"academy/internal/service"
	"academy/internal/service/ton"
	"academy/internal/service/upload"
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	rcron "github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Cron struct {
	logger *zap.Logger
	cron   *rcron.Cron

	videoProcessingMutex      sync.Mutex
	muxClearAssetsMutex       sync.Mutex
	muxUpdateReadyStatusMutex sync.Mutex
	updateTonPaymentsMutex    sync.Mutex

	uploadService   *upload.Service
	tonService      *ton.Service
	miniAppService  *service.MiniAppService
	materialService *service.MaterialService
}

const (
	RunningDailyAt11PM   = "0 23 * * *"
	RunningHourly        = "0 * * * *"
	RunningEveryMinute   = "*/1 * * * *"
	RunningEvery2Minutes = "*/2 * * * *"
)

const daysBeforeDeletingArchivedMiniApp = 7

func NewSomeCron(
	logger *zap.Logger,
	cron *rcron.Cron,

	uploadService *upload.Service,
	tonService *ton.Service,
	miniAppService *service.MiniAppService,
	materialService *service.MaterialService,
) (c *Cron, err error) {

	c = &Cron{
		logger: logger,
		cron:   cron,

		uploadService:   uploadService,
		tonService:      tonService,
		miniAppService:  miniAppService,
		materialService: materialService,
	}

	// Uncomment to run cron-jobs before starting API.
	// c.clearChunks()
	// c.clearUploads()
	// c.clearOldMiniApps()
	// c.videoProcessing()
	// c.muxClearAssets()
	// c.muxUpdateReadyStatus()
	// c.updateTonPayments()

	_, err = c.cron.AddFunc(RunningHourly, c.clearChunks)
	if err != nil {
		return nil, err
	}
	_, err = c.cron.AddFunc(RunningDailyAt11PM, c.clearUploads)
	if err != nil {
		return nil, err
	}
	_, err = c.cron.AddFunc(RunningDailyAt11PM, c.clearOldMiniApps)
	if err != nil {
		return nil, err
	}
	_, err = c.cron.AddFunc(RunningEveryMinute, c.videoProcessing)
	if err != nil {
		return nil, err
	}
	_, err = c.cron.AddFunc(RunningEveryMinute, c.muxClearAssets)
	if err != nil {
		return nil, err
	}
	_, err = c.cron.AddFunc(RunningEveryMinute, c.muxUpdateReadyStatus)
	if err != nil {
		return nil, err
	}
	_, err = c.cron.AddFunc(RunningEvery2Minutes, c.updateTonPayments)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Cron) clearChunks() {
	ctx := context.Background()

	err := c.uploadService.ClearOldChunks(ctx)
	if err != nil {
		c.logger.Error("clearChunks: cron job failed: failed to clear old chunks", zap.Error(err))
		return
	}

	c.logger.Info("clearChunks: cron job successfully finished")
}

func (c *Cron) clearUploads() {
	ctx := context.Background()

	err := c.uploadService.ClearDanglingUploads(ctx)
	if err != nil {
		c.logger.Error("clearUploads: cron job failed: failed to clear dangling uploads", zap.Error(err))
		return
	}

	c.logger.Info("clearUploads: cron job successfully finished")
}

func (c *Cron) clearOldMiniApps() {
	ctx := context.Background()

	miniAppIDs, err := c.miniAppService.DeleteOld(ctx, time.Now().AddDate(
		0, 0, -daysBeforeDeletingArchivedMiniApp))
	if err != nil {
		c.logger.Error("clearOldMiniApps: cron job failed: failed to clear old mini apps", zap.Error(err))
		return
	}

	for _, miniAppID := range miniAppIDs {
		pathToDelete := upload.MaterialFilePath{MiniAppID: miniAppID}

		err = c.uploadService.Delete(pathToDelete.String())
		if err != nil {
			c.logger.Error("clearOldMiniApps: failed to delete mini-app related uploaded files",
				zap.String("err", err.Error()))
		}
	}

	c.logger.Info("clearOldMiniApps: cron job successfully finished")
}

func (c *Cron) updateTonPayments() {
	if ok := c.updateTonPaymentsMutex.TryLock(); !ok {
		return
	}
	defer c.updateTonPaymentsMutex.Unlock()

	ctx := context.Background()

	addresses, err := c.miniAppService.FindTonAddresses(ctx)
	if err != nil {
		c.logger.Error("updateTonPayments: cron job failed: failed to find TON addresses", zap.Error(err))
		return
	}

	for _, addr := range addresses {
		n, err := c.tonService.UpdateJettonTransfers(ctx, addr)
		if err != nil {
			c.logger.Error("updateTonPayments: failed to update jetton transfers",
				zap.Error(err),
				zap.String("addr", addr),
			)

			continue
		}

		if n != 0 {
			c.logger.Info("updateTonPayments: added new jetton transfers",
				zap.String("addr", addr),
				zap.Int("transfers count", n),
			)
		}
	}
}

func (c *Cron) videoProcessing() {
	if ok := c.videoProcessingMutex.TryLock(); !ok {
		// c.logger.Info("videoProcessing: cron job skipped")
		return
	}
	defer c.videoProcessingMutex.Unlock()

	c.muxUploadVideos()
}

func (c *Cron) muxUploadVideos() {
	ctx := context.Background()

	// false withMetadata means asset is not uploaded yet.
	withMetadata := false
	materials, err := c.materialService.FindPendingMoveToMux(ctx, withMetadata, 100)
	if err != nil {
		c.logger.Error("muxUploadVideos: cron job failed: failed to find materials",
			zap.Error(err),
		)
		return
	}

	for _, m := range materials {
		c.logger.Info("muxUploadVideos: start uploading...",
			zap.String("material_id", m.ID.String()),
			zap.String("filename", m.Filename),
		)

		if m.Category != model.MaterialCategoryLessonContent || len(m.Metadata) != 0 {
			continue
		}

		metadata, err := c.uploadService.MuxUpload(ctx, m.Filename)
		if err != nil {
			c.logger.Error("muxUploadVideos: cron job failed: uploading to mux",
				zap.String("material_id", m.ID.String()),
				zap.Error(err),
			)
			return
		}
		c.logger.Info("muxUploadVideos: finished uploading to mux",
			zap.String("material_id", m.ID.String()),
		)

		rawMetadata, err := json.Marshal(metadata)
		if err != nil {
			c.logger.Error("muxUploadVideos: cron job failed: failed to marshal metadata",
				zap.String("material_id", m.ID.String()),
				zap.Error(err),
			)
			return
		}

		// TODO: Decide if file size should also be updated.
		m.Filename = ""
		m.Metadata = rawMetadata
		m.UpdatedAt = time.Now()

		if err := c.materialService.Update(ctx, m); err != nil {
			c.logger.Error("muxUploadVideos: cron job failed: failed to update material",
				zap.String("material_id", m.ID.String()),
				zap.Error(err),
			)
			return
		}
		c.logger.Info("muxUploadVideos: finished updating the material",
			zap.String("material_id", m.ID.String()),
		)
	}

	// c.logger.Info("muxUploadVideos: cron job successfully finished")
}

func (c *Cron) muxUpdateReadyStatus() {
	if ok := c.muxUpdateReadyStatusMutex.TryLock(); !ok {
		return
	}
	defer c.muxUpdateReadyStatusMutex.Unlock()

	ctx := context.Background()

	// true withMetadata means asset is already uploaded but could be not ready for playing yet.
	withMetadata := true
	materials, err := c.materialService.FindPendingMoveToMux(ctx, withMetadata, 100)
	if err != nil {
		c.logger.Error("muxUpdateReadyStatus: cron job failed: failed to find materials",
			zap.Error(err),
		)
		return
	}

	for _, m := range materials {
		c.logger.Info("muxUpdateReadyStatus: checking Mux asset status...",
			zap.String("material_id", m.ID.String()),
		)

		if m.Category != model.MaterialCategoryLessonContent || len(m.Metadata) == 0 {
			continue
		}

		var metadata model.MuxVideoMetadata
		if err := json.Unmarshal(m.Metadata, &metadata); err != nil {
			c.logger.Error("muxUpdateReadyStatus: cron job failed: decoding metadata",
				zap.String("material_id", m.ID.String()),
				zap.Error(err),
			)
			continue
		}

		assetResp, err := c.uploadService.MuxGetAsset(ctx, metadata.AssetID)
		if err != nil {
			c.logger.Error("muxUpdateReadyStatus: cron job failed: uploading to mux",
				zap.String("material_id", m.ID.String()),
				zap.String("asset_id", metadata.AssetID),
				zap.Error(err),
			)
			continue
		}
		if assetResp.Data.Status != "ready" {
			c.logger.Error("muxUpdateReadyStatus: cron job failed: asset not ready",
				zap.String("material_id", m.ID.String()),
				zap.String("asset_id", metadata.AssetID),
				zap.String("status", assetResp.Data.Status),
			)
			continue
		}

		m.Status = model.MaterialStatusReady
		m.UpdatedAt = time.Now()

		if err := c.materialService.Update(ctx, m); err != nil {
			c.logger.Error("muxUpdateReadyStatus: cron job failed: failed to update material",
				zap.String("material_id", m.ID.String()),
				zap.Error(err),
			)
			return
		}
		c.logger.Info("muxUpdateReadyStatus: finished updating the material",
			zap.String("material_id", m.ID.String()),
			zap.String("asset_id", metadata.AssetID),
		)
	}
}

func (c *Cron) muxClearAssets() {
	if ok := c.muxClearAssetsMutex.TryLock(); !ok {
		// c.logger.Info("muxClearAssets: cron job skipped")
		return
	}
	defer c.muxClearAssetsMutex.Unlock()

	ctx := context.Background()

	assets, err := c.materialService.FindMuxAssetsToDelete(ctx, 100)
	if err != nil {
		c.logger.Error("muxClearAssets: cron job failed: failed to find materials",
			zap.Error(err),
		)
		return
	}

	for _, assetID := range assets {
		c.logger.Info("muxClearAssets: start deleting asset...",
			zap.String("asset_id", assetID),
		)

		err := c.uploadService.MuxDeleteAsset(ctx, assetID)

		if err != nil && !errors.Is(err, upload.ErrNotFound) {
			c.logger.Error("muxClearAssets: cron job failed: failed to delete asset",
				zap.String("asset_id", assetID),
				zap.Error(err),
			)
			return
		}

		if errors.Is(err, upload.ErrNotFound) {
			c.logger.Warn("muxClearAssets: can't find asset to delete in Mux",
				zap.String("asset_id", assetID),
			)
		}

		if err := c.materialService.FullyDeleleMuxAsset(ctx, assetID); err != nil {
			c.logger.Error("muxClearAssets: cron job failed: failed to delete asset from DB",
				zap.String("asset_id", assetID),
				zap.Error(err),
			)
			return
		}

		c.logger.Info("muxClearAssets: finished deleting the asset",
			zap.String("asset_id", assetID),
		)
	}

	// c.logger.Info("muxClearAssets: cron job successfully finished")
}

func (c *Cron) start(_ context.Context) error {
	c.logger.Info("cron job started")
	c.cron.Start()
	return nil
}

func (c *Cron) stop(_ context.Context) error {
	c.logger.Info("cron job stopped")
	c.cron.Stop()
	return nil
}
