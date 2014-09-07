// testEcs project main.go
package main

import (
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/jmcvetta/napping"
	"net/http"
	"os"
)

const channelUrl string = "http://localhost:8080/ecs-api/rest/merchants/merchv06/channels"
const productUrl string = "http://localhost:8080/ecs-api/rest/merchants/merchv06/products"
const magentoMockurl string = "http://localhost:8080/ecs-integration-magento-mock/magento/products/"

func sHeader() *http.Header {
	header := &http.Header{}
	header.Set("Accept", "application/ecs+json;v0.7")
	header.Set("Authorization", "KEY2")
	header.Set("ECS-ApiVersion", "0.1")
	return header
}

type Channel struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	Endpoint        string `json:"endpoint"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	Token           string `json:"token"`
	CreateTimestamp string `json:"createTimestamp"`
	Active          string `json:"active"`
}

type Product struct {
	Code        string  `json:"code"`
	Sku         string  `json:"sku"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Id          string  `json:"id"`
	Qty         float32 `json:"qty"`
}

type MagProductAdjustmentRequest struct {
	Sku     string `json:"sku"`
	Add     string `json:"add"`
	Host    string `json:"host"`
	ApiUser string `json:"apiUser"`
	ApiKey  string `json:"apiKey"`
}

type CreateRequest struct {
	Entity Channel `json:"entity"`
}

type Constraints struct {
	Limit  string `json:"limit"`
	Offset string `json:"offset"`
}

type ListRequest struct {
	constraints Constraints `json:"constraints"`
}

type ChannelListResponse struct {
	ReferenceId string
	Count       int
	Data        []Channel
}

type ProductsListResponse struct {
	ReferenceId string
	Count       int
	Data        []Product
}

type SimpleRepose struct {
	Data  string
	RefId string
}

func main() {
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
	}))

	session := napping.Session{}
	session.Header = sHeader()

	m.Get("/", func(ren render.Render) {
		ren.HTML(200, "home", nil)
	})

	// get a list of all channels
	m.Get("/channels", func(ren render.Render) {
		constraints := Constraints{"1000", "0"}
		payload := ListRequest{constraints}
		result := ChannelListResponse{}
		resp, err := session.Post(channelUrl+"/list", &payload, &result, nil)
		_ = err
		_ = resp
		ren.HTML(200, "channels", result.Data)
	})

	// get data for specific channel
	m.Get("/channels/:code", func(ren render.Render, params martini.Params) {
		code := params["code"]
		result := ChannelListResponse{}
		resp, _ := session.Get(channelUrl+"/"+code, nil, &result, nil)
		_ = resp
		ren.HTML(200, "channel", result.Data[0])

	})

	// create a new channel
	m.Get("/new", func(ren render.Render) {
		channel := Channel{}
		ren.HTML(200, "channel", channel)
	})

	m.Post("/channels", func(ren render.Render, r *http.Request, w http.ResponseWriter) {

		c := Channel{}
		c.Code = r.FormValue("code")
		if len(r.FormValue("active")) > 0 {
			c.Active = r.FormValue("active")
		} else {
			c.Active = "yes"
		}
		c.Endpoint = r.FormValue("endpoint")
		c.Name = r.FormValue("name")
		c.Password = r.FormValue("password")
		c.Token = r.FormValue("token")
		c.Type = r.FormValue("type")
		c.Username = r.FormValue("username")
		c.CreateTimestamp = r.FormValue("createTimestamp")
		payload := CreateRequest{c}
		//ren.JSON(200, payload)
		result := SimpleRepose{}
		fmt.Println(c)
		if len(c.Code) < 1 {
			fmt.Println("Sending POST request")
			resp, err := session.Post(channelUrl, payload, &result, nil)
			_ = resp
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Println("Sending PUT request")
			resp, err := session.Put(channelUrl+"/"+c.Code, payload, &result, nil)
			_ = resp
			if err != nil {
				panic(err)
			}
		}

		c.Code = result.Data
		r.Method = "GET"
		http.Redirect(w, r, "/channels", 302)
	})

	m.Get("/products", func(ren render.Render) {
		constraints := Constraints{"1000", "0"}
		payload := ListRequest{constraints}
		result := ProductsListResponse{}
		resp, err := session.Post(productUrl+"/list", &payload, &result, nil)
		_ = err
		_ = resp
		ren.HTML(200, "products", result.Data)
	})

	m.Get("/products/fetch", func(ren render.Render, r *http.Request) {
		sku := r.URL.Query().Get("sku")
		host := "new.a.touchlab.co.za"
		apiUser := "adent"
		apiKey := "password"
		params := napping.Params{}
		params["host"] = host
		params["apiUser"] = apiUser
		params["apiKey"] = apiKey
		params["sku"] = sku
		result := Product{}
		resp, err := session.Get(magentoMockurl, &params, &result, nil)
		_ = err
		_ = resp
		ren.HTML(200, "magproduct", result)

	})

	m.Get("products/adjust", func(ren render.Render, r *http.Request) {
		sku := r.URL.Query().Get("sku")
		adjust := r.URL.Query().Get("adjustment")
		host := "new.a.touchlab.co.za"
		apiUser := "adent"
		apiKey := "password"
		payload := MagProductAdjustmentRequest{}
		payload.Add = adjust
		payload.ApiKey = apiKey
		payload.ApiUser = apiUser
		payload.Host = host
		payload.Sku = sku
		result := Product{}
		resp, err := session.Put(magentoMockurl, &payload, &result, nil)
		_ = resp
		_ = err
		ren.JSON(200, result)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	m.RunOnAddr(":" + port)
}
