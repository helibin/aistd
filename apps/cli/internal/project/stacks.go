package project

func hasJS(c Config) bool {
	return c.APIStack == "nestjs" || c.Web || ((c.IOS || c.Android) && c.MobileStack != "flutter") || c.WeChat || c.Alipay
}

func apiLabel(c Config) string {
	if c.APIStack == "nestjs" {
		return "NestJS API"
	}
	return "Go API"
}
