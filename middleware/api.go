package middleware

import (
	"bytes"
	"fmt"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/gin-gonic/gin"
	"github.com/phinexdaz/ipapk"
	"github.com/satori/go.uuid"
	"github.com/xvwvx/ipapk-server/conf"
	"github.com/xvwvx/ipapk-server/models"
	"github.com/xvwvx/ipapk-server/serializers"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func Upload(c *gin.Context) {
	changelog := c.PostForm("changelog")
	file, err := c.FormFile("file")
	if err != nil {
		return
	}

	ext := models.BundleFileExtension(filepath.Ext(file.Filename))
	if !ext.IsValid() {
		return
	}

	_uuid := uuid.NewV4().String()
	filename := filepath.Join(".data", _uuid+string(ext.PlatformType().Extention()))

	if err := c.SaveUploadedFile(file, filename); err != nil {
		return
	}

	app, err := ipapk.NewAppParser(filename)
	if err != nil {
		return
	}

	if conf.AppConfig.IsUseAliyun {
		models.OSSPutObjectFromFile(_uuid+string(ext.PlatformType().Extention()), filename)
		os.Remove(filename)
	}

	iconBuffer := new(bytes.Buffer)
	if err := png.Encode(iconBuffer, app.Icon); err != nil {
		return
	}

	bundle := new(models.Bundle)
	bundle.UUID = _uuid
	bundle.PlatformType = ext.PlatformType()
	bundle.Name = app.Name
	bundle.BundleId = app.BundleId
	bundle.Version = app.Version
	bundle.Build = app.Build
	bundle.Size = app.Size
	bundle.Icon = iconBuffer.Bytes()
	bundle.ChangeLog = changelog

	if err := models.AddBundle(bundle); err != nil {
		return
	}

	c.JSON(http.StatusOK, &serializers.BundleJSON{
		UUID:       _uuid,
		Name:       bundle.Name,
		Platform:   bundle.PlatformType.String(),
		BundleId:   bundle.BundleId,
		Version:    bundle.Version,
		Build:      bundle.Build,
		InstallUrl: bundle.GetInstallUrl(conf.AppConfig.ProxyURL()),
		QRCodeUrl:  conf.AppConfig.ProxyURL() + "/qrcode/" + _uuid,
		IconUrl:    conf.AppConfig.ProxyURL() + "/icon/" + _uuid,
		Changelog:  bundle.ChangeLog,
		Downloads:  bundle.Downloads,
	})
}

func DelBundle(c *gin.Context) {
	_uuid := c.Param("uuid")

	bundle, err := models.GetBundleByUID(_uuid)
	if err != nil {
		c.JSON(404, map[string]string{
			"msg": "未找到bundle",
		})
		return
	}

	if conf.AppConfig.IsUseAliyun {
		err = models.OSSDelFile(_uuid+string(bundle.PlatformType.Extention()))
	} else {
		filename := filepath.Join(".data", _uuid+string(bundle.PlatformType.Extention()))
		err = os.Remove(filename)
	}
	if err != nil {
		c.JSON(404, map[string]string{
			"msg": "删除文件错误",
		})
		return
	}

	err = bundle.DeleteBundle()
	if err != nil {
		c.JSON(404, map[string]string{
			"msg": "删除数据错误",
		})
		return
	}

	c.JSON(200, map[string]string{
		"msg": "已删除",
	})
}

func GetQRCode(c *gin.Context) {
	_uuid := c.Param("uuid")

	bundle, err := models.GetBundleByUID(_uuid)
	if err != nil {
		return
	}

	data := fmt.Sprintf("%v/bundle/%v?_t=%v", conf.AppConfig.ProxyURL(), bundle.UUID, time.Now().Unix())
	code, err := qr.Encode(data, qr.L, qr.Unicode)
	if err != nil {
		return
	}
	code, err = barcode.Scale(code, 160, 160)
	if err != nil {
		return
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, code); err != nil {
		return
	}

	c.Data(http.StatusOK, "image/png", buf.Bytes())
}

func GetIcon(c *gin.Context) {
	_uuid := c.Param("uuid")

	bundle, err := models.GetBundleByUID(_uuid)
	if err != nil {
		return
	}

	c.Data(http.StatusOK, "image/png", bundle.Icon)
}

func GetChangelog(c *gin.Context) {
	_uuid := c.Param("uuid")

	bundle, err := models.GetBundleByUID(_uuid)
	if err != nil {
		return
	}

	c.HTML(http.StatusOK, "change.html", gin.H{
		"changelog": bundle.ChangeLog,
	})
}

func GetBundleId(c *gin.Context) {
	_bundleId := c.Param("bundle_id")

	bundle, err := models.GetBundleByBundleId(_bundleId)
	if err != nil {
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"bundle":     bundle,
		"installUrl": bundle.GetInstallUrl(conf.AppConfig.ProxyURL()),
		"qrCodeUrl":  conf.AppConfig.ProxyURL() + "/qrcode/" + bundle.UUID,
		"iconUrl":    conf.AppConfig.ProxyURL() + "/icon/" + bundle.UUID,
	})
}

func GetBundle(c *gin.Context) {
	_uuid := c.Param("uuid")

	bundle, err := models.GetBundleByUID(_uuid)
	if err != nil {
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"bundle":     bundle,
		"installUrl": bundle.GetInstallUrl(conf.AppConfig.ProxyURL()),
		"qrCodeUrl":  conf.AppConfig.ProxyURL() + "/qrcode/" + bundle.UUID,
		"iconUrl":    conf.AppConfig.ProxyURL() + "/icon/" + bundle.UUID,
	})
}

func GetVersions(c *gin.Context) {
	_uuid := c.Param("uuid")

	bundle, err := models.GetBundleByUID(_uuid)
	if err != nil {
		return
	}

	versions, err := bundle.GetVersions()
	if err != nil {
		return
	}

	c.HTML(http.StatusOK, "version.html", gin.H{
		"versions": versions,
		"uuid":     bundle.UUID,
	})
}

func GetBuilds(c *gin.Context) {
	_uuid := c.Param("uuid")
	version := c.Param("ver")

	bundle, err := models.GetBundleByUID(_uuid)
	if err != nil {
		return
	}

	builds, err := bundle.GetBuilds(version)
	if err != nil {
		return
	}

	var bundles []serializers.BundleJSON
	for _, v := range builds {
		bundles = append(bundles, serializers.BundleJSON{
			UUID:       v.UUID,
			Name:       v.Name,
			Platform:   v.PlatformType.String(),
			BundleId:   v.BundleId,
			Version:    v.Version,
			Build:      v.Build,
			InstallUrl: v.GetInstallUrl(conf.AppConfig.ProxyURL()),
			QRCodeUrl:  conf.AppConfig.ProxyURL() + "/qrcode/" + v.UUID,
			IconUrl:    conf.AppConfig.ProxyURL() + "/icon/" + v.UUID,
			Changelog:  bundle.ChangeLog,
			Downloads:  v.Downloads,
		})
	}

	c.HTML(http.StatusOK, "build.html", gin.H{
		"builds": bundles,
	})
}

func GetPlist(c *gin.Context) {
	_uuid := c.Param("uuid")

	bundle, err := models.GetBundleByUID(_uuid)
	if err != nil {
		return
	}

	if bundle.PlatformType != models.BundlePlatformTypeIOS {
		return
	}

	ipaUrl := conf.AppConfig.ProxyURL() + "/ipa/" + bundle.UUID

	data, err := models.NewPlist(bundle.Name, bundle.Version, bundle.BundleId, ipaUrl).Marshall()
	if err != nil {
		return
	}

	c.Data(http.StatusOK, "application/x-plist", data)
}

func DownloadAPP(c *gin.Context) {
	_uuid := c.Param("uuid")

	bundle, err := models.GetBundleByUID(_uuid)
	if err != nil {
		return
	}

	bundle.UpdateDownload()

	downloadUrl := conf.AppConfig.VisitURL() + "/ipapk/" + bundle.UUID + string(bundle.PlatformType.Extention())
	c.Redirect(http.StatusMovedPermanently, downloadUrl)
}
