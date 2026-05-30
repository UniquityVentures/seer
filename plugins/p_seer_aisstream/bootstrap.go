package p_seer_aisstream

import (
	"gorm.io/gorm"
)

func init() {
	registerPluginDBInitHook("p_seer_aisstream.models", func(db *gorm.DB) *gorm.DB {
		if err := db.AutoMigrate(&AISStreamMessage{}, &AISStreamPositionReport{}, &AISStreamStandardClassBPositionReport{}, &AISStreamUnknownMessage{}, &AISStreamAddressedSafetyMessage{}, &AISStreamAddressedBinaryMessage{}, &AISStreamAidsToNavigationReport{}, &AISStreamAssignedModeCommand{}, &AISStreamBaseStationReport{}, &AISStreamBinaryAcknowledge{}, &AISStreamBinaryBroadcastMessage{}, &AISStreamChannelManagement{}, &AISStreamCoordinatedUTCInquiry{}, &AISStreamDataLinkManagementMessage{}, &AISStreamDataLinkManagementMessageData{}, &AISStreamExtendedClassBPositionReport{}, &AISStreamGnssBroadcastBinaryMessage{}, &AISStreamGroupAssignmentCommand{}, &AISStreamInterrogation{}, &AISStreamLongRangeAisBroadcastMessage{}, &AISStreamMultiSlotBinaryMessage{}, &AISStreamSafetyBroadcastMessage{}, &AISStreamShipStaticData{}, &AISStreamSingleSlotBinaryMessage{}, &AISStreamStandardSearchAndRescueAircraftReport{}, &AISStreamStaticDataReport{}); err != nil {
			panic(err)
		}
		startAISStreamWorkerIfConfigured(db)
		return db
	})
}
