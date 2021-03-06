package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/xvwvx/ipapk-server/conf"
	"github.com/xvwvx/ipapk-server/middleware"
	"github.com/xvwvx/ipapk-server/models"
	"github.com/xvwvx/ipapk-server/templates"
	"github.com/xvwvx/ipapk-server/utils"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func Init() {
	_, err := os.Stat(".data")
	if os.IsNotExist(err) {
		os.MkdirAll(".data", 0755)
	}

	if err := utils.InitCA(); err != nil {
		log.Fatal(err)
	}

	if err := conf.InitConfig("config.json"); err != nil {
		log.Fatal(err)
	}

	if err := models.InitDB(); err != nil {
		log.Fatal(err)
	}

	if conf.AppConfig.IsUseAliyun {
		if err := models.InitOSS(); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	Init()

	router := gin.Default()
	router.SetFuncMap(templates.TplFuncMap)
	router.LoadHTMLGlob("public/views/*")

	router.Static("ipapk", ".data")
	router.Static("static", "public/static")
	router.StaticFile("myCA.cer", ".ca/myCA.cer")

	router.GET("/", middleware.GetList)
	router.POST("/upload", middleware.Upload)
	router.GET("/del/:uuid", middleware.DelBundle)
	router.GET("/bundleId/:bundle_id", middleware.GetBundleId)
	router.GET("/bundle/:uuid", middleware.GetBundle)
	router.GET("/log/:uuid", middleware.GetChangelog)
	router.GET("/qrcode/:uuid", middleware.GetQRCode)
	router.GET("/icon/:uuid", middleware.GetIcon)
	router.GET("/plist/:uuid", middleware.GetPlist)
	router.GET("/ipa/:uuid", middleware.DownloadAPP)
	router.GET("/apk/:uuid", middleware.DownloadAPP)
	router.GET("/version/:uuid", middleware.GetVersions)
	router.GET("/version/:uuid/:ver", middleware.GetBuilds)

	srv := &http.Server{
		Addr:    conf.AppConfig.Addr(),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %v\n", err)
		}
		//if err := srv.ListenAndServeTLS(".ca/mycert1.cer", ".ca/mycert1.key"); err != nil {
		//	log.Printf("listen: %v\n", err)
		//}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}
