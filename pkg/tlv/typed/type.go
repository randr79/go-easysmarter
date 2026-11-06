package typed

import "fmt"

// go":generate stringer -type=Type -trimprefix=Type
type Type int

const (
	TypeModel                        Type = 0x0001 // 1
	TypeDescription                  Type = 0x0002 // 2
	TypeMacAddress                   Type = 0x0003 // 3
	TypeIpAddress                    Type = 0x0004 // 4
	TypeSubnetMask                   Type = 0x0005 // 5
	TypeGatewayAddress               Type = 0x0006 // 6
	TypeFirmwareVersion              Type = 0x0007 // 7
	TypeHardwareRevision             Type = 0x0008 // 8
	TypeDhcpEnabled                  Type = 0x0009 // 9
	TypePortCount                    Type = 0x000A // 10
	TypeManagementVlan               Type = 0x000B // 11
	TypeLedEnabled                   Type = 0x000C // 12
	TypeAutosaving                   Type = 0x000D // 13
	TypeFactoryDefault               Type = 0x000E // 14
	TypeFlashTypeId                  Type = 0x000F // 15
	TypePortOrderReversed            Type = 0x0010 // 16
	TypeUserName                     Type = 0x0200 // 512
	TypeNewUserName                  Type = 0x0201 // 513
	TypePassword                     Type = 0x0202 // 514
	TypeNewPassword                  Type = 0x0203 // 515
	TypeRsaChallenge                 Type = 0x0210 // 528
	TypeReadConfig                   Type = 0x0300 // 768
	TypeWriteConfig                  Type = 0x0301 // 769
	TypeReboot                       Type = 0x0305 // 773
	TypeFactoryReset                 Type = 0x0500 // 1280
	TypeRebootToFirmwareUpgradeMode  Type = 0x0600 // 1536
	TypePersistConfig                Type = 0x0900 // 2304
	TypeGetToken                     Type = 0x0901 // 2305
	TypePortSettings                 Type = 0x1000 // 4096
	TypeIgmpSnoopingEnabled          Type = 0x1100 // 4352
	TypeIgmpMulticastIps             Type = 0x1101 // 4353
	TypeIgmpReportSuppressionEnabled Type = 0x1102 // 4354
	TypeLagConfig                    Type = 0x1200 // 4608
	TypeMultiTenantUplink            Type = 0x2000 // 8192
	TypePortVlanEnabled              Type = 0x2100 // 8448
	TypePortVlanMembers              Type = 0x2101 // 8449
	TypePortVlanCount                Type = 0x2102 // 8450
	TypeIeee8021QVlanEnabled         Type = 0x2200 // 8704
	TypeIeee8021QVlanMembers         Type = 0x2201 // 8705
	TypeIeee8021QPortNativeVlan      Type = 0x2202 // 8706
	TypeIeee8021QMaxVlans            Type = 0x2203 // 8707
	TypeQosType                      Type = 0x3000 // 12288
	TypeQosPortPriority              Type = 0x3001 // 12289
	TypeBandwidthLimitIngress        Type = 0x3100 // 12544
	TypeBandwidthLimitEgress         Type = 0x3101 // 12545
	TypeStormControl                 Type = 0x3200 // 12800
	TypePortStatistics               Type = 0x4000 // 16384
	TypePortMirroring                Type = 0x4100 // 16640
	TypeCableTest                    Type = 0x4200 // 16896
	TypeLoopPreventionEnabled        Type = 0x4300 // 17152
)

func TLVByString(name string) (TLV, error) {
	var instance TLV
	switch name {
	case "Model":
		instance = new(Model)
	case "Description":
		instance = new(Description)
	case "MacAddress":
		instance = new(MacAddress)
	case "IpAddress":
		instance = new(IpAddress)
	case "SubnetMask":
		instance = new(SubnetMask)
	case "GatewayAddress":
		instance = new(GatewayAddress)
	case "FirmwareVersion":
		instance = new(FirmwareVersion)
	case "HardwareRevision":
		instance = new(HardwareRevision)
	case "DhcpEnabled":
		instance = new(DhcpEnabled)
	case "PortCount":
		instance = new(PortCount)
	case "LedEnabled":
		instance = new(LedEnabled)
	case "Autosaving":
		instance = new(Autosaving)
	case "FactoryDefault":
		instance = new(FactoryDefault)
	case "FlashTypeId":
		instance = new(FlashTypeId)
	case "PortOrderReversed":
		instance = new(PortOrderReversed)
	case "UserName":
		instance = new(UserName)
	case "NewUserName":
		instance = new(NewUserName)
	case "Password":
		instance = new(Password)
	case "NewPassword":
		instance = new(NewPassword)
	case "RsaChallenge":
		instance = new(RsaChallenge)
	case "ReadConfig":
		instance = new(ReadConfig)
	case "WriteConfig":
		instance = new(WriteConfig)
	case "Reboot":
		instance = new(Reboot)
	case "FactoryReset":
		instance = new(FactoryReset)
	case "RebootToFirmwareUpgradeMode":
		instance = new(RebootToFirmwareUpgradeMode)
	case "PersistConfig":
		instance = new(PersistConfig)
	case "GetToken":
		instance = new(GetToken)
	case "PortSettings":
		instance = new(PortSettings)
	case "IgmpSnoopingEnabled":
		instance = new(IgmpSnoopingEnabled)
	case "IgmpMulticastIps":
		instance = new(IgmpMulticastIps)
	case "IgmpReportSuppressionEnabled":
		instance = new(IgmpReportSuppressionEnabled)
	case "LagConfig":
		instance = new(LagConfig)
	case "MultiTenantUplink":
		instance = new(MultiTenantUplink)
	case "PortVlanEnabled":
		instance = new(PortVlanEnabled)
	case "PortVlanMembers":
		instance = new(PortVlanMembers)
	case "PortVlanCount":
		instance = new(PortVlanCount)
	case "Ieee8021QVlanEnabled":
		instance = new(Ieee8021QVlanEnabled)
	case "Ieee8021QVlanMembers":
		instance = new(Ieee8021QVlanMembers)
	case "Ieee8021QPortNativeVlan":
		instance = new(Ieee8021QPortNativeVlan)
	case "Ieee8021QMaxVlans":
		instance = new(Ieee8021QMaxVlans)
	case "QosType":
		instance = new(QosType)
	case "QosPortPriority":
		instance = new(QosPortPriority)
	case "BandwidthLimitIngress":
		instance = new(BandwidthLimitIngress)
	case "BandwidthLimitEgress":
		instance = new(BandwidthLimitEgress)
	case "StormControl":
		instance = new(StormControl)
	case "PortStatistics":
		instance = new(PortStatistics)
	case "PortMirroring":
		instance = new(PortMirroring)
	case "CableTest":
		instance = new(CableTest)
	case "LoopPreventionEnabled":
		instance = new(LoopPreventionEnabled)
	default:
		return nil, fmt.Errorf("unknown Type %s", name)
	}
	return instance, nil
}
