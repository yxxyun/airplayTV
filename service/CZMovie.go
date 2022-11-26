package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/lixiang4u/airplayTV/model"
	"github.com/lixiang4u/airplayTV/util"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	czHost      = "https://www.czspp.com"
	czTagUrl    = "https://www.czspp.com/%s/page/%d"
	czSearchUrl = "https://www.czspp.com/xssearch?q=%s&p=%d"
	czDetailUrl = "https://www.czspp.com/movie/%s.html"
	czPlayUrl   = "https://www.czspp.com/v_play/%s.html"
)

//========================================================================
//==============================接口实现===================================
//========================================================================

type CZMovie struct {
	movie       Movie
	httpWrapper *util.HttpWrapper
}

func (x *CZMovie) Init(movie Movie) {
	x.movie = movie
	if x.httpWrapper == nil {
		x.httpWrapper = &util.HttpWrapper{}
	}
	x.httpWrapper.SetHeader("origin", czHost)
	x.httpWrapper.SetHeader("authority", util.HandleHostname(czHost))
	x.httpWrapper.SetHeader("referer", czHost)
	x.httpWrapper.SetHeader("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36")
	x.httpWrapper.SetHeader("cookie", "")
}

func (x *CZMovie) ListByTag(tagName, page string) model.Pager {
	return x.czListByTag(tagName, page)
}

func (x *CZMovie) Search(search, page string) model.Pager {
	return x.czListBySearch(search, page)
}

func (x *CZMovie) Detail(id string) model.MovieInfo {
	return x.czVideoDetail(id)
}

func (x *CZMovie) Source(sid, vid string) model.Video {
	return x.czVideoSource(sid, vid)
}

//========================================================================
//==============================实际业务处理逻辑============================
//========================================================================

func (x *CZMovie) czListByTag(tagName, page string) model.Pager {
	_page, _ := strconv.Atoi(page)

	var pager = model.Pager{}
	pager.Limit = 25

	err := x.SetCookie()
	if err != nil {
		log.Println("[绕过人机失败]", err.Error())
		return pager
	}
	b, err := x.httpWrapper.Get(fmt.Sprintf(czTagUrl, tagName, _page))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return pager
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return pager
	}
	doc.Find(".mi_cont .mi_ne_kd ul li").Each(func(i int, selection *goquery.Selection) {
		name := selection.Find(".dytit a").Text()
		tmpUrl, _ := selection.Find(".dytit a").Attr("href")
		thumb, _ := selection.Find("img.thumb").Attr("data-original")
		tag := selection.Find(".nostag").Text()
		actors := selection.Find(".inzhuy").Text()
		resolution := selection.Find(".hdinfo span").Text()

		pager.List = append(pager.List, model.MovieInfo{
			Id:         util.CZHandleUrlToId(tmpUrl),
			Name:       name,
			Thumb:      thumb,
			Url:        tmpUrl,
			Actors:     strings.TrimSpace(actors),
			Tag:        tag,
			Resolution: resolution,
		})
	})

	doc.Find(".pagenavi_txt a").Each(func(i int, selection *goquery.Selection) {
		tmpHref, _ := selection.Attr("href")
		tmpList := strings.Split(tmpHref, "/")
		n, _ := strconv.Atoi(tmpList[len(tmpList)-1])
		if n*pager.Limit > pager.Total {
			pager.Total = n * pager.Limit
		}
	})

	pager.Current, _ = strconv.Atoi(doc.Find(".pagenavi_txt .current").Text())

	return pager
}

func (x *CZMovie) czListBySearch(query, page string) model.Pager {
	var pager = model.Pager{}
	pager.Limit = 20

	err := x.SetCookie()
	if err != nil {
		log.Println("[绕过人机失败]", err.Error())
		return pager
	}
	b, err := x.httpWrapper.Get(fmt.Sprintf(czSearchUrl, query, util.HandlePageNumber(page)))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return pager
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return pager
	}

	doc.Find(".search_list ul li").Each(func(i int, selection *goquery.Selection) {
		name := selection.Find(".dytit a").Text()
		tmpUrl, _ := selection.Find(".dytit a").Attr("href")
		thumb, _ := selection.Find("img.thumb").Attr("data-original")
		tag := selection.Find(".nostag").Text()
		actors := selection.Find(".inzhuy").Text()

		pager.List = append(pager.List, model.MovieInfo{
			Id:     util.CZHandleUrlToId(tmpUrl),
			Name:   name,
			Thumb:  thumb,
			Url:    tmpUrl,
			Actors: strings.TrimSpace(actors),
			Tag:    tag,
		})
	})

	doc.Find(".dytop .dy_tit_big span").Each(func(i int, selection *goquery.Selection) {
		if i == 0 {
			pager.Total, _ = strconv.Atoi(selection.Text())
		}
	})

	pager.Current, _ = strconv.Atoi(doc.Find(".pagenavi_txt .current").Text())

	return pager
}

func (x *CZMovie) czVideoDetail(id string) model.MovieInfo {
	var info = model.MovieInfo{}

	err := x.SetCookie()
	if err != nil {
		log.Println("[绕过人机失败]", err.Error())
		return info
	}
	b, err := x.httpWrapper.Get(fmt.Sprintf(czDetailUrl, id))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return info
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return info
	}

	doc.Find(".paly_list_btn a").Each(func(i int, selection *goquery.Selection) {
		tmpHref, _ := selection.Attr("href")
		info.Links = append(info.Links, model.Link{
			Id:    util.CZHandleUrlToId2(tmpHref),
			Name:  strings.ReplaceAll(selection.Text(), "厂长", ""),
			Url:   tmpHref,
			Group: "资源1",
		})
	})

	info.Id = id
	info.Thumb, _ = doc.Find(".dyxingq .dyimg img").Attr("src")
	info.Name = doc.Find(".dyxingq .moviedteail_tt h1").Text()
	info.Intro = strings.TrimSpace(doc.Find(".yp_context").Text())

	return info
}

func (x *CZMovie) czVideoSource(sid, vid string) model.Video {
	var video = model.Video{Id: sid}

	err := x.SetCookie()
	if err != nil {
		log.Println("[绕过人机失败]", err.Error())
		return video
	}
	b, err := x.httpWrapper.Get(fmt.Sprintf(czPlayUrl, sid))
	if err != nil {
		log.Println("[内容获取失败]", err.Error())
		return video
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
	if err != nil {
		log.Println("[文档解析失败]", err.Error())
		return video
	}

	var findLine = ""
	tmpList := strings.Split(string(b), "\n")
	for _, line := range tmpList {
		if strings.Contains(line, "md5.AES.decrypt") {
			findLine = line
			break
		}
	}
	if findLine != "" {
		video, err = x.czParseVideoSource(sid, findLine)

		bs, _ := json.MarshalIndent(video, "", "\t")
		log.Println(fmt.Sprintf("[video] %s", string(bs)))
		if err != nil {
			log.Println("[parse.video.error]", err)
		}
	}

	// 解析另一种iframe嵌套的视频
	iframeUrl, _ := doc.Find(".videoplay iframe").Attr("src")
	if strings.TrimSpace(iframeUrl) != "" {
		if _, ok := util.RefererConfig[util.HandleHost(iframeUrl)]; ok {
			//需要chromedp加载后拿播放信息（数据通过js加密了）
			video.Source = iframeUrl
			video.Url = handleIframeEncrypedSourceUrl(iframeUrl)
		} else {
			// 直接可以拿到播放信息
			video.Source, video.Type = getFrameUrlContents(iframeUrl)
			video.Url = HandleSrcM3U8FileToLocal(video.Id, video.Source, x.movie.IsCache)
			// 1、转为本地m3u8
			// 2、修改m3u8文件内容地址,支持跨域
		}
	}

	video.Name = doc.Find(".jujiinfo h3").Text()

	// 视频类型问题处理
	video = handleVideoType(video)

	return video
}

func (x *CZMovie) czParseVideoSource(id, js string) (model.Video, error) {
	var video = model.Video{}
	tmpList := strings.Split(strings.TrimSpace(js), ";")

	var data = ""
	var key = ""
	var iv = ""
	for index, str := range tmpList {
		if index == 0 {
			regex := regexp.MustCompile(`"\S+"`)
			data = strings.Trim(regex.FindString(str), `"`)
			continue
		}
		if index == 1 {
			regex := regexp.MustCompile(`"(\S+)"`)
			matchList := regex.FindStringSubmatch(str)
			if len(matchList) > 0 {
				key = matchList[len(matchList)-1]
			}
			continue
		}
		if index == 2 {
			regex := regexp.MustCompile(`\((\S+)\)`)
			matchList := regex.FindStringSubmatch(str)
			if len(matchList) > 0 {
				iv = matchList[len(matchList)-1]
			}
			continue
		}
	}

	log.Println(fmt.Sprintf("[parsing] key: %s, iv: %s", key, iv))

	if key == "" && data == "" {
		return video, errors.New("解析失败")
	}
	bs, err := util.DecryptByAes([]byte(key), []byte(iv), data)
	if err != nil {
		return video, errors.New("解密失败")
	}
	tmpList = strings.Split(string(bs), "window")
	if len(tmpList) < 1 {
		return video, errors.New("解密数据错误")
	}

	regex := regexp.MustCompile(`{url: "(\S+)",type:"(\S+)",([\S\s]*)pic:'(\S+)'}`)
	matchList := regex.FindStringSubmatch(tmpList[0])

	if len(matchList) < 1 {
		return video, errors.New("解析视频信息失败")
	}

	video.Id = id

	for index, m := range matchList {
		switch index {
		case 1:
			video.Source = m
			video.Url = m
			break
		case 2:
			video.Type = m
			break
		case 4:
			video.Thumb = m
			break
		default:
			break
		}
	}

	video.Url = HandleSrcM3U8FileToLocal(id, video.Source, x.movie.IsCache)

	return video, nil
}

func handleVideoType(v model.Video) model.Video {
	// "https://yun.m3.c-zzy.online:1011/d/%E9%98%BF%E9%87%8C1%E5%8F%B7/%E8%8B%8F%E9%87%8C%E5%8D%97/Narco-Saints.S01E01.mp4?type=m3u8"

	tmpUrl, err := url.Parse(v.Source)
	if err != nil {
		return v
	}
	if util.StringInList(tmpUrl.Host, []string{"yun.m3.c-zzy.online:1011"}) {
		v.Type = "hls"
	}
	return v
}

func getFrameUrlContents(frameUrl string) (sourceUrl, videoType string) {
	sourceUrl = frameUrl
	videoType = "auto"

	resp, err := http.Get(frameUrl)
	if err != nil {
		log.Println("[getFrameUrlContents.get.error]", err.Error())
		return
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("[getFrameUrlContents.body.error]", err.Error())
		return
	}

	// 匹配播放文件
	regEx := regexp.MustCompile(`sources: \[{(\s+)src: '(\S+)',(\s+)type: '(\S+)'`)
	r := regEx.FindStringSubmatch(string(bs))
	if len(r) < 4 {
		return
	}
	sourceUrl = r[2]

	switch r[4] {
	case "application/vnd.apple.mpegurl":
		videoType = "hls"
	}

	return
}

func handleIframeEncrypedSourceUrl(iframeUrl string) string {
	log.Println("[load.encrypted.iframe.video]")
	var err error

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var videoUrl string
	var videoUrlOk bool
	err = chromedp.Run(
		ctx,
		//chromedp.Navigate(iframeUrl),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, _, _, err := page.Navigate(iframeUrl).WithReferrer("https://www.czspp.com/").Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
		//chromedp.Evaluate(`urls;`, &res),
		// wait for footer element is visible (ie, page is loaded)
		// find and click "Example" link
		//chromedp.Click(`#example-After`, chromedp.NodeVisible),
		// retrieve the text of the textarea
		//chromedp.Value(`#div_player source`, &example),

		chromedp.WaitVisible(`#div_player`),

		chromedp.AttributeValue(`#div_player video source`, "src", &videoUrl, &videoUrlOk),
	)
	if err != nil && !strings.Contains(err.Error(), "net::ERR_ABORTED") {
		// Note: Ignoring the net::ERR_ABORTED page error is essential here
		// since downloads will cause this error to be emitted, although the
		// download will still succeed.
		log.Println("[network.error]", err)
		return ""
	}

	return videoUrl
}

func isWaf(html string) []byte {
	regEx := regexp.MustCompile(`window.location.href ="(\S+)";`)
	f := regEx.FindStringSubmatch(html)
	if len(f) < 2 {
		return nil
	}

	log.Println("[=========>]", fmt.Sprintf("%s%s", util.HandleHost(czHost), f[1]))

	resp, err := http.Get(fmt.Sprintf("%s%s", util.HandleHost(czHost), f[1]))
	if err != nil {
		log.Println("[IsWaf.error]", err.Error())
		return nil
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("[IsWaf.resp.body]", err.Error())
		return nil
	}
	log.Println("[IsWaf.error]", string(b))
	return b
}

func (x *CZMovie) GetVerifyUrl() string {
	b, err := x.httpWrapper.Get(czHost)
	if err != nil {
		log.Println("[访问主站错误]", err.Error())
		return ""
	}
	regEx := regexp.MustCompile(`<script type="text/javascript" src="(\S+)"></script>`)
	matchResult := regEx.FindStringSubmatch(string(b))

	log.Println("[人机认证]", util.ToJSON(matchResult, false))

	if len(matchResult) < 2 {
		return ""
	}
	b, err = x.httpWrapper.Get(fmt.Sprintf("%s%s", strings.TrimRight(czHost, "/"), matchResult[1]))
	if err != nil {
		log.Println("[访问认证JS错误]", err.Error())
		return ""
	}

	regEx = regexp.MustCompile(`var key="(\w+)",value="(\w+)";`)
	matchResult2 := regEx.FindStringSubmatch(string(b))
	if len(matchResult2) < 3 {
		log.Println("[匹配认证配置错误] response:", string(b))
		return ""
	}
	log.Println("[解析验证配置]", util.ToJSON(matchResult2, true))

	regEx = regexp.MustCompile(`c.get\(\"(\S+)\&key\=`)
	matchResult3 := regEx.FindStringSubmatch(string(b))
	if len(matchResult3) < 2 {
		log.Println("[匹配认证地址错误] response:", string(b))
		return ""
	}
	log.Println("[解析验证地址]", util.ToJSON(matchResult3, true))

	tmpUrl := fmt.Sprintf("%s%s&key=%s&value=%s", strings.TrimRight(czHost, "/"), matchResult3[1], matchResult2[1], matchResult2[2])

	return tmpUrl
}

func (x *CZMovie) SetCookie() error {
	tmpUrl := x.GetVerifyUrl()
	if tmpUrl == "" {
		return errors.New("解析人机认证失败")
	}
	h, body, err := x.httpWrapper.GetResponse(tmpUrl)

	if err != nil {
		return err
	}
	tmpV := strings.TrimSpace(string(body))
	if v, ok := h["Set-Cookie"]; ok && strings.Contains(strings.TrimSpace(v[0]), tmpV) {
		x.httpWrapper.SetHeader("cookie", v[0])
		return nil
	}

	return errors.New("没有发现cookie")
}
