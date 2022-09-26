package wattpilot

var propertyMap = map[string]string{
	"fbufAge":                            "fbuf_age",
	"localTime":                          "loc",
	"otaCloudProgress":                   "ocp",
	"chargingCurrent":                    "amp",
	"espFreeHeap32":                      "efh32",
	"hostname":                           "host",
	"registeredCards":                    "cards",
	"espMinFreeHeap":                     "emfh",
	"inverterDataAge":                    "inva",
	"timeServerSyncStatus":               "tsss",
	"currentlyConnectedWifi":             "ccw",
	"factoryWifiApName":                  "fwan",
	"buttonAllowCurrentChange":           "bac",
	"carType":                            "ct",
	"otaCloudLength":                     "ocl",
	"cableLock":                          "ust",
	"minimumChargingInterval":            "mci",
	"minChargePauseDuration":             "mcpd",
	"rebootCharger":                      "rst",
	"espHeapSize":                        "ehs",
	"akkuSoc":                            "fbuf_akkuSOC",
	"powerPv":                            "fbuf_pPv",
	"modelStatus":                        "modelStatus",
	"timeSinceBoot":                      "rbt",
	"usePvSurplus":                       "fup",
	"lastStaSwitchedFromConnected":       "lssfc",
	"minChargeTime":                      "fmt",
	"forceSinglePhase":                   "fsp",
	"httpStaReachable":                   "hws",
	"loadBalancingStatus":                "los",
	"avgPowerOhmpilot":                   "pvopt_averagePOhmpilot",
	"timezoneDaylightSavingMode":         "tds",
	"colorWaitCar":                       "cwc",
	"wifiPlannedMac":                     "wpb",
	"wifiStaErrorCount":                  "wsc",
	"cloudWsConnected":                   "cwsc",
	"cloudWsConnectedAge":                "cwsca",
	"lastCarStateChangedFromCharging":    "lccfc",
	"cableCurrentLimit":                  "cbl",
	"colorIdle":                          "cid",
	"lockFeedbackAge":                    "ffba",
	"roundingMode":                       "frm",
	"firmwareCarControl":                 "fwc",
	"modelStatusInternal":                "msi",
	"totalPowerAverage":                  "tpa",
	"energyCounterSinceStart":            "wh",
	"cableUnlockStatus":                  "cus",
	"espFreeHeap":                        "efh",
	"otaPartition":                       "otap",
	"httpConnectedClients":               "wcch",
	"loadGroupId":                        "log",
	"rebootCounter":                      "rbc",
	"carState":                           "car",
	"carConsumption":                     "cco",
	"logicMode":                          "lmo",
	"minChargePauseEndsAt":               "mcpea",
	"otaCloudBranches":                   "ocu",
	"pvOptSpecialCase":                   "pvopt_specialCase",
	"espResetReason":                     "rr",
	"variant":                            "var",
	"wsConnectedClients":                 "wccw",
	"adapterLimit4":                      "al4",
	"lastCarStateChangedToCharging":      "lcctc",
	"pGrid":                              "pgrid",
	"averagePAkku":                       "pvopt_averagePAkku",
	"wifiConfigs":                        "wifis",
	"friendlyName":                       "fna",
	"otaCloudUseClientAuth":              "ocuca",
	"wifiScanAge":                        "scaa",
	"transaction":                        "trx",
	"espFlashInfo":                       "efi",
	"effectiveRoundingMode":              "ferm",
	"ledBrightness":                      "lbr",
	"partitionTable":                     "part",
	"pPv":                                "ppv",
	"phaseSwitchMode":                    "psm",
	"ohmpilotTemperature":                "fbuf_ohmpilotTemperature",
	"forceState":                         "frc",
	"moduleHwPcbVersion":                 "mod",
	"schedulerWeekday":                   "sch_week",
	"adapterLimit5":                      "al5",
	"allowCharging":                      "alw",
	"awattarMaxPrice":                    "awp",
	"currentLimitPresets":                "clp",
	"firmwareDescription":                "apd",
	"loadBalancingAmpere":                "loa",
	"loadBalancingType":                  "loty",
	"residualCurrentDetection":           "rcd",
	"unlockPowerOutage":                  "upo",
	"wifiApKey":                          "wak",
	"wifiCurrentMac":                     "wcb",
	"dnsServer":                          "dns",
	"wifiScanResult":                     "scan",
	"temperatureSensors":                 "tma",
	"cloudWsEnabled":                     "cwe",
	"loadBalancingEnabled":               "loe",
	"appRecommendedVersion":              "arv",
	"zeroFeedin":                         "fzf",
	"secureBootEnabled":                  "sbe",
	"timeServerSyncMode":                 "tssm",
	"colorFinished":                      "cfi",
	"deltaCurrent":                       "pvopt_deltaA",
	"chargingEnergyLimit":                "dwo",
	"otaCloudStatus":                     "ocs",
	"lastModelStatusChange":              "lmsc",
	"otaNewestVersion":                   "onv",
	"energyTotalPersisted":               "etop",
	"loadBalancingMembers":               "lom",
	"timezoneOffset":                     "tof",
	"maxCurrentLimit":                    "ama",
	"otaCloudApp":                        "oca",
	"timeServer":                         "ts",
	"cpEnable":                           "cpe",
	"factoryWifiApKey":                   "facwak",
	"frequency":                          "fhz",
	"averagePGrid":                       "pvopt_averagePGrid",
	"lastButtonPress":                    "lbp",
	"lastForceSinglePhaseToggle":         "lfspt",
	"loadBalancingTotalAmpere":           "lot",
	"oemManufacturer":                    "oem",
	"zeroFeedinOffset":                   "zfo",
	"wifiRssi":                           "rssi",
	"awattarCurrentPrice":                "awcp",
	"colorCharging":                      "cch",
	"stopHysteresis":                     "sh",
	"timeServerEnabled":                  "tse",
	"wifiStaErrorMessage":                "wsm",
	"prioOffset":                         "po",
	"simulateUnpluggingDuration":         "sumd",
	"wifiSsid":                           "wss",
	"adapterLimit3":                      "al3",
	"minPhaseToggleWaitTime":             "mptwt",
	"norwayMode":                         "nmo",
	"wifiStateMachineState":              "wsms",
	"accessState":                        "acs",
	"espCpuFreq":                         "ecf",
	"allowedCurrent":                     "acu",
	"factoryFriendlyName":                "ffna",
	"minPhaseWishSwitchTime":             "mpwst",
	"serialNumber":                       "sse",
	"utcTime":                            "utc",
	"deltaPower":                         "pvopt_deltaP",
	"energyCounterTotal":                 "eto",
	"numberOfPhases":                     "pnp",
	"schedulerSaturday":                  "sch_satur",
	"cpEnableRequest":                    "cpr",
	"powerAkku":                          "fbuf_pAkku",
	"inverterDataOverride":               "ido",
	"minChargingCurrent":                 "mca",
	"queueSizeCloud":                     "qsc",
	"flashEncryptionMode":                "fem",
	"chargeControllerUpdateProgress":     "ccu",
	"energySetKwh":                       "esk",
	"ohmpilotTemperatureLimit":           "fot",
	"cloudClientAuth":                    "cca",
	"adapterLimit1":                      "al1",
	"adapterLimit2":                      "al2",
	"espFreeHeap8":                       "efh8",
	"espMaxHeap":                         "emhb",
	"otaCloudMessage":                    "ocm",
	"wifiEnabled":                        "wen",
	"wifiFailedMac":                      "wfb",
	"rtcResetReasons":                    "esr",
	"pvBatteryLimit":                     "fam",
	"deviceType":                         "typ",
	"loadMapping":                        "map",
	"phases":                             "pha",
	"timeServerOperatingMode":            "tsom",
	"schedulerSunday":                    "sch_sund",
	"espChipInfo":                        "eci",
	"lockFeedback":                       "ffb",
	"defaultRoute":                       "nif",
	"simulateUnplugging":                 "su",
	"adapterLimit":                       "adi",
	"temperatureCurrentLimit":            "amt",
	"powerAcTotal":                       "fbuf_pAcTotal",
	"useDynamicPricing":                  "ful",
	"lastStaSwitchedToConnected":         "lsstc",
	"energy":                             "nrg",
	"awattarCountry":                     "awc",
	"chargeControllerRecommendedVersion": "ccrv",
	"errorState":                         "err",
	"ohmpilotState":                      "fbuf_ohmpilotState",
	"ledSaveEnergy":                      "lse",
	"phaseSwitchHysteresis":              "psh",
	"phaseWishMode":                      "pwm",
	"loadPriority":                       "lop",
	"threePhaseSwitchLevel":              "spl3",
	"loadFallback":                       "lof",
	"lastPvSurplusCalculation":           "lpsc",
	"akkuMode":                           "fbuf_akkuMode",
	"averagePPv":                         "pvopt_averagePPv",
	"queueSizeWs":                        "qsw",
	"lastCarStateChangedFromIdle":        "lccfi",
	"wifiApName":                         "wan",
	"wifiStaStatus":                      "wst",
	"powerGrid":                          "fbuf_pGrid",
	"startingPower":                      "fst",
	"httpStaAuthentication":              "hsa",
	"forceSinglePhaseDuration":           "psmd",
	"partitionTableOffset":               "pto",
	"wifiScanStatus":                     "scas",
	"simulateUnpluggingAlways":           "sua",
	"timeServerSyncInterval":             "tssi",
	"effectiveLockSetting":               "lck",
	"cloudWsStarted":                     "cws",
	"ledInfo":                            "led",
	"awattarPriceList":                   "awpl",
	"chargingDurationInfo":               "cdi",
	"relayFeedback":                      "rfb",
	"forceSinglePhaseToggleWishedSince":  "fsptws",
	"firmwareVersion":                    "fwv",
	"pAkku":                              "pakku",
}
