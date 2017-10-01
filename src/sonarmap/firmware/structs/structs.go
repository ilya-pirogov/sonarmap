package structs

type Service struct {
    Port, Version int64
    Service string
}

type Settings struct {
    AppVersion,
    Barcode,
    Brand,
    ContentID,
    IP,
    IP_Zeroconfig,
    Language,
    LanguagePack,
    Model,
    Name,
    PlatformType,
    PlatformVersion,
    SerialNumber string
    Services []Service
}
