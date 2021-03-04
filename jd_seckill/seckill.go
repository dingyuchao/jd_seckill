package jd_seckill

import (
	"errors"
	"fmt"
	"github.com/Albert-Zhan/httpc"
	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
	"jd_seckill/common"
	"jd_seckill/conf"
	"jd_seckill/service"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Seckill struct {
	client *httpc.HttpClient
	conf   *conf.Config
}

func NewSeckill(client *httpc.HttpClient, conf *conf.Config) *Seckill {
	return &Seckill{client: client, conf: conf}
}

func (this *Seckill) SkuTitle() (string, error) {
	skuId := this.conf.Read("config", "sku_id")
	req := httpc.NewRequest(this.client)
	resp, body, err := req.SetUrl(fmt.Sprintf("https://item.jd.com/%s.html", skuId)).SetMethod("get").Send().End()
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println("访问商品详情失败")
		return "", errors.New("访问商品详情失败")
	}
	html := strings.NewReader(body)
	doc, _ := goquery.NewDocumentFromReader(html)
	return strings.TrimSpace(doc.Find(".sku-name").Text()), nil
}

func (this *Seckill) MakeReserve() {
	shopTitle, err := this.SkuTitle()
	if err != nil {
		log.Println("获取商品信息失败")
	} else {
		log.Println("商品名称:" + shopTitle)
	}
	skuId := this.conf.Read("config", "sku_id")
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.conf.Read("config", "DEFAULT_USER_AGENT"))
	req.SetHeader("Referer", fmt.Sprintf("https://item.jd.com/%s.html", skuId))
	resp, body, err := req.SetUrl("https://yushou.jd.com/youshouinfo.action?callback=fetchJSON&sku=" + skuId + "&_=" + strconv.Itoa(int(time.Now().Unix()*1000))).SetMethod("get").Send().End()
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println("预约商品失败")
	} else {
		reserveUrl := gjson.Get(body, "url").String()
		req = httpc.NewRequest(this.client)
		_, _, _ = req.SetUrl("https:" + reserveUrl).SetMethod("get").Send().End()
		log.Println("预约成功，已获得抢购资格 / 您已成功预约过了，无需重复预约")
	}
}

func (this *Seckill) getSeckillUrl() (string, error) {
	skuId := this.conf.Read("config", "sku_id")
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.conf.Read("config", "DEFAULT_USER_AGENT"))
	req.SetHeader("Host", "itemko.jd.com")
	req.SetHeader("Referer", fmt.Sprintf("https://item.jd.com/%s.html", skuId))
	resp, body, err := req.SetUrl("https://itemko.jd.com/itemShowBtn?callback=jQuery{}" + strconv.Itoa(common.Rand(1000000, 9999999)) + "&skuId=" + skuId + "&from=pc&_=" + strconv.Itoa(int(time.Now().Unix()*1000))).SetMethod("get").Send().End()
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println("抢购链接获取失败，稍后自动重试")
		return "", errors.New("抢购链接获取失败，稍后自动重试")
	}
	url := gjson.Get(body, "url").String()
	if url == "" {
		log.Println("抢购链接获取失败，稍后自动重试")
		return "", errors.New("抢购链接获取失败，稍后自动重试")
	}
	//https://divide.jd.com/user_routing?skuId=8654289&sn=c3f4ececd8461f0e4d7267e96a91e0e0&from=pc
	url = strings.ReplaceAll(url, "divide", "marathon")
	//https://marathon.jd.com/captcha.html?skuId=8654289&sn=c3f4ececd8461f0e4d7267e96a91e0e0&from=pc
	url = strings.ReplaceAll(url, "user_routing", "captcha.html")
	return url, nil
}

func (this *Seckill) RequestSeckillUrl() {
	user := NewUser(this.client, this.conf)
	userInfo, err := user.GetUserInfo()
	if err != nil {
		log.Println("获取用户信息失败")
	} else {
		log.Println("用户:" + userInfo)
	}
	shopTitle, err := this.SkuTitle()
	if err != nil {
		log.Println("获取商品信息失败")
	} else {
		log.Println("商品名称:" + shopTitle)
	}
	url, _ := this.getSeckillUrl()
	skuId := this.conf.Read("config", "sku_id")
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.conf.Read("config", "DEFAULT_USER_AGENT"))
	req.SetHeader("Host", "marathon.jd.com")
	req.SetHeader("Referer", fmt.Sprintf("https://item.jd.com/%s.html", skuId))
	_, _, _ = req.SetUrl(url).SetMethod("get").Send().End()
}

func (this *Seckill) SeckillPage() {
	log.Println("访问抢购订单结算页面...")
	skuId := this.conf.Read("config", "sku_id")
	seckillNum := this.conf.Read("config", "seckill_num")
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.conf.Read("config", "DEFAULT_USER_AGENT"))
	req.SetHeader("Host", "marathon.jd.com")
	req.SetHeader("Referer", fmt.Sprintf("https://item.jd.com/%s.html", skuId))
	_, _, _ = req.SetUrl("https://marathon.jd.com/seckill/seckill.action?skuId=" + skuId + "&num=" + seckillNum + "&rid=" + strconv.Itoa(int(time.Now().Unix()))).SetMethod("get").Send().End()
}

func (this *Seckill) SeckillInitInfo() (string, error) {
	log.Println("获取秒杀初始化信息...")
	skuId := this.conf.Read("config", "sku_id")
	seckillNum := this.conf.Read("config", "seckill_num")
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.conf.Read("config", "DEFAULT_USER_AGENT"))
	req.SetHeader("Host", "marathon.jd.com")
	req.SetData("sku", skuId)
	req.SetData("num", seckillNum)
	req.SetData("isModifyAddress", "false")
	resp, body, err := req.SetUrl("https://marathon.jd.com/seckillnew/orderService/pc/init.action").SetMethod("post").Send().End()
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println("初始化秒杀信息失败")
		return "", errors.New("初始化秒杀信息失败")
	}
	return body, nil
}

func (this *Seckill) SubmitSeckillOrder() bool {
	eid := this.conf.Read("config", "eid")
	fp := this.conf.Read("config", "fp")
	skuId := this.conf.Read("config", "sku_id")
	seckillNum := this.conf.Read("config", "seckill_num")
	paymentPwd := this.conf.Read("account", "payment_pwd")
	initInfo, _ := this.SeckillInitInfo()
	address := gjson.Get(initInfo, "addressList").Array()
	defaultAddress := address[0]
	isinvoiceInfo := gjson.Get(initInfo, "invoiceInfo").Exists()
	invoiceTitle := "-1"
	invoiceContentType := "-1"
	invoicePhone := ""
	invoicePhoneKey := ""
	if isinvoiceInfo {
		invoiceTitle = gjson.Get(initInfo, "invoiceInfo.invoiceTitle").String()
		invoiceContentType = gjson.Get(initInfo, "invoiceInfo.invoiceContentType").String()
		invoicePhone = gjson.Get(initInfo, "invoiceInfo.invoicePhone").String()
		invoicePhoneKey = gjson.Get(initInfo, "invoiceInfo.invoicePhoneKey").String()
	}
	invoiceInfo := "false"
	if isinvoiceInfo {
		invoiceInfo = "true"
	}
	token := gjson.Get(initInfo, "token").String()
	log.Println("提交抢购订单...")
	req := httpc.NewRequest(this.client)
	req.SetHeader("User-Agent", this.conf.Read("config", "DEFAULT_USER_AGENT"))
	req.SetHeader("Host", "marathon.jd.com")
	req.SetHeader("Referer", fmt.Sprintf("https://marathon.jd.com/seckill/seckill.action?skuId=%s&num=%s&rid=%d", skuId, seckillNum, int(time.Now().Unix())))
	req.SetData("skuId", skuId)
	req.SetData("num", seckillNum)
	req.SetData("addressId", defaultAddress.Get("id").String())
	req.SetData("yuShou", "true")
	req.SetData("isModifyAddress", "false")
	req.SetData("name", defaultAddress.Get("name").String())
	req.SetData("provinceId", defaultAddress.Get("provinceId").String())
	req.SetData("cityId", defaultAddress.Get("cityId").String())
	req.SetData("countyId", defaultAddress.Get("countyId").String())
	req.SetData("townId", defaultAddress.Get("townId").String())
	req.SetData("addressDetail", defaultAddress.Get("addressDetail").String())
	req.SetData("mobile", defaultAddress.Get("mobile").String())
	req.SetData("mobileKey", defaultAddress.Get("mobileKey").String())
	req.SetData("email", defaultAddress.Get("email").String())
	req.SetData("postCode", "")
	req.SetData("invoiceTitle", invoiceTitle)
	req.SetData("invoiceCompanyName", "")
	req.SetData("invoiceContent", invoiceContentType)
	req.SetData("invoiceTaxpayerNO", "")
	req.SetData("invoiceEmail", "")
	req.SetData("invoicePhone", invoicePhone)
	req.SetData("invoicePhoneKey", invoicePhoneKey)
	req.SetData("invoice", invoiceInfo)
	req.SetData("password", paymentPwd)
	req.SetData("codTimeType", "3")
	req.SetData("paymentType", "4")
	req.SetData("areaCode", "")
	req.SetData("overseas", "0")
	req.SetData("phone", "")
	req.SetData("eid", eid)
	req.SetData("fp", fp)
	req.SetData("token", token)
	req.SetData("pru", "")
	resp, body, err := req.SetUrl("https://marathon.jd.com/seckillnew/orderService/pc/submitOrder.action?skuId=" + skuId).SetMethod("post").Send().End()
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println("抢购失败，网络错误")
		if this.conf.Read("messenger", "enable") == "true" && this.conf.Read("messenger", "type") == "smtp" {
			email := service.NerEmail(this.conf)
			_ = email.SendMail([]string{this.conf.Read("messenger", "email")}, "茅台抢购通知", "抢购失败，网络错误")
		}
		return false
	}
	if !gjson.Valid(body) {
		log.Println("抢购失败，返回信息:" + common.Substr(body, 0, 128))
		if this.conf.Read("messenger", "enable") == "true" && this.conf.Read("messenger", "type") == "smtp" {
			email := service.NerEmail(this.conf)
			_ = email.SendMail([]string{this.conf.Read("messenger", "email")}, "茅台抢购通知", "抢购失败，返回信息:"+common.Substr(body, 0, 128))
		}
		return false
	}
	if gjson.Get(body, "success").Bool() {
		orderId := gjson.Get(body, "orderId").String()
		totalMoney := gjson.Get(body, "totalMoney").String()
		payUrl := "https:" + gjson.Get(body, "pcUrl").String()
		log.Println(fmt.Sprintf("抢购成功，订单号:%s, 总价:%s, 电脑端付款链接:%s", orderId, totalMoney, payUrl))
		if this.conf.Read("messenger", "enable") == "true" && this.conf.Read("messenger", "type") == "smtp" {
			email := service.NerEmail(this.conf)
			msg := fmt.Sprintf("抢购成功，订单号:%s, 总价:%s, 电脑端付款链接:%s", orderId, totalMoney, payUrl)
			_ = email.SendMail([]string{this.conf.Read("messenger", "email")}, "茅台抢购通知", msg)

			service.NotifyUser(this.conf.Read("pushmsg", "token"), msg)
		}
		return true
	} else {
		log.Println("抢购失败，返回信息:" + body)
		if this.conf.Read("messenger", "enable") == "true" && this.conf.Read("messenger", "type") == "smtp" {
			email := service.NerEmail(this.conf)
			_ = email.SendMail([]string{this.conf.Read("messenger", "email")}, "茅台抢购通知", "抢购失败，返回信息:"+body)
		}
		return false
	}
}
