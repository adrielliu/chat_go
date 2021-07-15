/**
 * Created by lock
 * Date: 2019-08-12
 * Time: 11:36
 */
package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"chat_go/config"
	"net/http"
)

func main() {
	siteConfig := config.Conf.Site
	port := siteConfig.SiteBase.ListenPort
	addr := fmt.Sprintf(":%d", port)
	logrus.Fatal(http.ListenAndServe(addr, http.FileServer(http.Dir("./site/"))))
}
