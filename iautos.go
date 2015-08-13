//
package main


import (
    "github.com/PuerkitoBio/goquery"
    "fmt"
    //"github.com/robertkrimen/otto"
    "log"
    "regexp"
    "strings"
    "strconv"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    //"sync"
    //"runtime"
    //"flag"
    "github.com/willf/bloom"
)


func IautosSpider(start_url string) {
      fmt.Printf("Fetching: %s\n",start_url)
      var is_seller,addr string
      fmt.Printf("spider url is : %s\n",start_url)
      if strings.Contains(start_url,"as2ds9vepcatcpbnscac") {
        is_seller = "个人"
      } else if strings.Contains(start_url,"as1ds9vepcatcpbnscac") {
        is_seller = "商家"
      }
      addr = strings.Split(start_url,"/")[3]
      dict := map[string]string{"jiangsu": "江苏", "anhui": "安徽","shandong": "山东", "shanghai": "上海","zhejiang": "浙江", "jiangxi": "江西"}
      province, _ := dict[addr]

      doc, err := goquery.NewDocument(start_url) 
      check(err)

      var urls []string
      doc.Find(".carShow ul li h4 a").Each(func(i int, s *goquery.Selection) {
        href, _ := s.Attr("href")
        urls = append(urls, href)
        //url := s.Find("h4 a").Attr("href")
        // title := s.Find("h4 a").Text()
        // register_date := s.Find(".txt .set .year").Text()
        // fmt.Printf("Review %d: %s - %s\n", i, title, register_date)
      })



      db, err := sql.Open("mysql", "panpan:panpan@tcp(192.168.2.231:3306)/spider?charset=utf8")
      check(err)
      defer db.Close()

      for idx,url := range urls {
        //fmt.Printf("Url %d : %s\n", idx, url)
    		filter := bloom.NewWithEstimates(100000,0.00001)
    		if filter.TestString(url) {
    			fmt.Printf("%s is already in the database,continue!\n",url)
    			continue
    		} else {
    			filter.AddString(url)
    		}
    		
		
        car, err := goquery.NewDocument(url) 
        check(err)

        content := car.Find("div[class='cd-content clearfix'] .clearfix .main")
        carinfo := content.Find(".cd-summary")
        title := carinfo.Find("h2 b").Text()
        rel_time := carinfo.Find("h2 span").Text()

        re, _ := regexp.Compile(`(\d{4}|\d{2})-(\d{2})-(\d{2})`)
        release_time := re.Find([]byte(rel_time))
        //fmt.Println(release_time)
        prices := carinfo.Find(".summary-txt .h136 .price").Text()
        re, _ = regexp.Compile(`(\d+).(\d+)`)
        price := re.Find([]byte(prices))
        owner_readme := content.Find(".cd-details .postscript p").Text()
        // execute javascript to show the telephone number
        // tele_div := carinfo.Find(".summary-txt script")
        // vm := otto.New()
        // _, err = vm.Run(tele_div.Text())
        // check(err)

        telephone := carinfo.Find(".summary-txt .call-num").Text()
        name := carinfo.Find(".summary-txt .seller-name span").Text()
        var configs []string
        carinfo.Find(".summary-txt .h136 dl dd").Each(func(i int, s *goquery.Selection) {
            if i != 0{
                configs = append(configs, s.Text()) 
            }
        })
        fmt.Printf("Confis Length is : %d\n", len(configs))
        var reg_date,register_date,config,address string
        var mileage,displacement,transmission string
        if len(configs) == 3 {
            reg_date, config, address = configs[0], configs[1], configs[2]
        } else if len(configs) == 4 {
            reg_date, config, address = configs[0], configs[1], configs[3]
        }
        reg_date2 := strings.Replace(reg_date, "年", "-", -1)
        register_date = strings.Replace(reg_date2, "月", "-", -1)
        register_date += "01"
        address = province + address
        for _, val := range strings.Split(config,"，") {
            if strings.Contains(val,"万公里") {
                re, _ = regexp.Compile(`(\d+).(\d+)`)
                mid_var := re.Find([]byte(val))
                if len(mileage) == 0 {
                    mileage = string(mid_var) 
                }
                
            } else if strings.Contains(val,"L") {
                re, _ = regexp.Compile(`(\d+).(\d+)`)
                mid_var := re.Find([]byte(val))
                if len(displacement) == 0 {
                    displacement = string(mid_var) 
                }
            } else {
                if strings.Contains(val,"手动") {
                    transmission = "手动"
                } else if strings.Contains(val,"自动") {
                    transmission = "自动"
                } else if strings.Contains(val,"手自一体") {
                    transmission = "手自一体"
                }
                
            }
        }

        fmt.Printf("里程：%s\n", mileage)
        fmt.Printf("排量：%s\n", displacement)
        fmt.Printf("变速箱：%s\n", transmission)
        fmt.Printf("车主姓名：%s\n", name)
        fmt.Printf("商家或个人：%s\n", is_seller)
        fmt.Printf("详细地址：%s\n", address)
        fmt.Printf("Info %d : %s - %s - %s - %s - %s - %s - %s - %s \n %s", idx, url, title, release_time, price, register_date, config, address, telephone, owner_readme)

        // Prepare statement for inserting data
        stmtIns, err := db.Prepare("insert ignore into sell_car_info(title,car_config,name,telephone_num,addrs,release_time,prices,is_seller,info_src,url,owner_readme,mileage,register_date,transmission,displacement) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
        check(err)
        defer stmtIns.Close() 
        
        _, err = stmtIns.Exec(title,config,name,telephone,address,release_time,price,is_seller,"iautos",url,owner_readme,mileage,register_date,transmission,displacement)
        check(err)
        
      
  }
}

// fatal if there is an error
func check(err error) {
    if err != nil {
        log.Fatal(err)
    }
}


func main() {

  var start_urls []string

  for i :=1; i < 5; i++ {
     start_urls = append(start_urls,"http://so.iautos.cn/jiangsu/p" + strconv.Itoa(i) + "as2ds9vepcatcpbnscac/")
     start_urls = append(start_urls,"http://so.iautos.cn/anhui/p" + strconv.Itoa(i) + "as2ds9vepcatcpbnscac/")
     start_urls = append(start_urls,"http://so.iautos.cn/shandong/p" + strconv.Itoa(i) + "as2ds9vepcatcpbnscac/")
     start_urls = append(start_urls,"http://so.iautos.cn/shanghai/p" + strconv.Itoa(i) + "as2ds9vepcatcpbnscac/")
     start_urls = append(start_urls,"http://so.iautos.cn/zhejiang/p" + strconv.Itoa(i) + "as2ds9vepcatcpbnscac/")
     start_urls = append(start_urls,"http://so.iautos.cn/jiangxi/p" + strconv.Itoa(i) + "as2ds9vepcatcpbnscac/")
  }
  for i :=1; i < 8; i++ {
     start_urls = append(start_urls,"http://so.iautos.cn/jiangsu/p" + strconv.Itoa(i) + "as1ds9vepcatcpbnscac/")
     start_urls = append(start_urls,"http://so.iautos.cn/anhui/p" + strconv.Itoa(i) + "as1ds9vepcatcpbnscac/")
     start_urls = append(start_urls,"http://so.iautos.cn/shandong/p" + strconv.Itoa(i) + "as1ds9vepcatcpbnscac/")
     start_urls = append(start_urls,"http://so.iautos.cn/shanghai/p" + strconv.Itoa(i) + "as1ds9vepcatcpbnscac/")
     start_urls = append(start_urls,"http://so.iautos.cn/zhejiang/p" + strconv.Itoa(i) + "as1ds9vepcatcpbnscac/")
     start_urls = append(start_urls,"http://so.iautos.cn/jiangxi/p" + strconv.Itoa(i) + "as1ds9vepcatcpbnscac/")
  }
  queue := make(chan string,12)     

  go enqueue(start_urls,queue)
  for uri := range queue {
    IautosSpider(uri)
  }

}


func enqueue(start_urls []string,c chan string) {
    for _, start_url := range start_urls {
      c <- start_url
    }                                           
    close(c)
}