package models

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

var rpl sync.Mutex

type BundlePlatformType int

const (
	BundlePlatformTypeAndroid BundlePlatformType = 1 + iota
	BundlePlatformTypeIOS
)

func (platformType BundlePlatformType) Extention() BundleFileExtension {
	var ext BundleFileExtension
	if platformType == BundlePlatformTypeAndroid {
		ext = BundleFileExtensionAndroid
	} else if platformType == BundlePlatformTypeIOS {
		ext = BundleFileExtensionIOS
	}
	return ext
}

type BundleFileExtension string

const (
	BundleFileExtensionAndroid BundleFileExtension = ".apk"
	BundleFileExtensionIOS     BundleFileExtension = ".ipa"
)

func (ext BundleFileExtension) IsValid() bool {
	if ext == BundleFileExtensionAndroid {
		return true
	} else if ext == BundleFileExtensionIOS {
		return true
	}
	return false
}

func (ext BundleFileExtension) PlatformType() BundlePlatformType {
	var platformType BundlePlatformType
	if ext == BundleFileExtensionAndroid {
		platformType = BundlePlatformTypeAndroid
	} else if ext == BundleFileExtensionIOS {
		platformType = BundlePlatformTypeIOS
	}
	return platformType
}

func (platformType BundlePlatformType) String() string {
	var out string
	if platformType == BundlePlatformTypeAndroid {
		out = "Android"
	} else if platformType == BundlePlatformTypeIOS {
		out = "iOS"
	}
	return out
}

type Bundle struct {
	ID           uint   `gorm:"primary_key"`
	UUID         string `gorm:"unique_index"`
	PlatformType BundlePlatformType
	Name         string
	BundleId     string
	Version      string
	Build        string
	Size         int64
	Icon         []byte
	ChangeLog    string `gorm:"type:text"`
	Downloads    uint64 `gorm:"default:0"`
	CreatedAt    time.Time
}

func GetList() ([]*Bundle, error) {
	var bundles []*Bundle

	rows, err := orm.Raw("SELECT * FROM (SELECT * FROM bundles ORDER BY created_at DESC) GROUP BY bundle_id, platform_type").
		Rows()
	defer rows.Close()
	for rows.Next() {
		var bundle Bundle
		orm.ScanRows(rows, &bundle)
		bundles = append(bundles, &bundle)
	}

	return bundles, err
}

func AddBundle(bundle *Bundle) error {
	return orm.Create(bundle).Error
}

func (bundle *Bundle) DeleteBundle() error {
	err := orm.Delete(bundle).Error
	return err
}

func GetBundleByBundleId(bundleId, platform string) (*Bundle, error) {
	var bundle Bundle

	if platform == "" {
		err := orm.Where("bundle_id = ?", bundleId).
			Order("created_at DESC").Limit(1).Find(&bundle).Error
		return &bundle, err
	}

	platformType := BundlePlatformTypeIOS
	if platform == BundlePlatformTypeAndroid.String() {
		platformType = BundlePlatformTypeAndroid
	}

	err := orm.Where("bundle_id = ? AND platform_type= ?", bundleId, int(platformType)).
		Order("created_at DESC").Limit(1).Find(&bundle).Error
	return &bundle, err
}

func GetBundleByUID(uuid string) (*Bundle, error) {
	var bundle Bundle

	err := orm.Where("uuid = ?", uuid).Find(&bundle).Limit(1).Error
	return &bundle, err
}

func (bundle *Bundle) UpdateBundle(field string, value interface{}) error {
	err := orm.Model(&bundle).Update(field, value).Error
	return err
}

func (bundle *Bundle) GetInstallUrl(baseUrl string) string {
	var out string
	if bundle.PlatformType == BundlePlatformTypeAndroid {
		out = baseUrl + "/apk/" + bundle.UUID
	} else if bundle.PlatformType == BundlePlatformTypeIOS {
		out = "itms-services://?action=download-manifest&url=" + baseUrl + "/plist/" + bundle.UUID
	}
	return out
}

func (bundle *Bundle) UpdateDownload() {
	val := bundle.Downloads
	rpl.Lock()
	atomic.AddUint64(&val, 1)
	bundle.UpdateBundle("downloads", val)
	rpl.Unlock()
}

func (bundle *Bundle) GetVersions() (VersionInfo, error) {
	results := make(VersionInfo, 0)

	rows, err := orm.Table("bundles").Select("version, count(DISTINCT build) AS builds").
		Where("bundle_id = ? AND platform_type= ?", bundle.BundleId, int(bundle.PlatformType)).Group("version").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		var builds int
		if err := rows.Scan(&version, &builds); err != nil {
			return nil, err
		}
		results = append(results, vInfo{version, builds})
	}
	sort.Sort(results)

	return results, nil
}

func (bundle *Bundle) GetBuilds(version string) ([]*Bundle, error) {
	var bundles []*Bundle
	err := orm.Where("bundle_id = ? AND version = ? AND platform_type = ?", bundle.BundleId, version, int(bundle.PlatformType)).
		Order("created_at DESC").Find(&bundles).Error

	return bundles, err
}
