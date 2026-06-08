package model

// 常用物流公司编码（与前端下拉选项一致）
const (
	ExpressSF    = "SF"
	ExpressZTO   = "ZTO"
	ExpressYTO   = "YTO"
	ExpressYD    = "YD"
	ExpressSTO   = "STO"
	ExpressJTSD  = "JTSD"
	ExpressEMS   = "EMS"
	ExpressJD    = "JD"
	ExpressDBL   = "DBL"
	ExpressOther = "OTHER"
)

var ExpressCompanyNames = map[string]string{
	ExpressSF:    "顺丰速运",
	ExpressZTO:   "中通快递",
	ExpressYTO:   "圆通速递",
	ExpressYD:    "韵达快递",
	ExpressSTO:   "申通快递",
	ExpressJTSD:  "极兔速递",
	ExpressEMS:   "邮政 EMS",
	ExpressJD:    "京东物流",
	ExpressDBL:   "德邦快递",
	ExpressOther: "其他快递",
}

func ExpressCompanyLabel(code string) string {
	if name, ok := ExpressCompanyNames[code]; ok {
		return name
	}
	return code
}
