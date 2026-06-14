package finance_record

type FinanceRecordSummary struct {
	TotalDonationProgram        int     `json:"totalDonationProgram"`
	TotalSocialProgram          int     `json:"totalSocialProgram"`
	TotalFosterChildren         int     `json:"totalFosterChildren"`
	TotalDonationProgramExpense float64 `json:"totalDonationProgramExpense"`
	TotalSocialProgramExpense   float64 `json:"totalSocialProgramExpense"`
	TotalFosterChildrenExpense  float64 `json:"totalFosterChildrenExpense"`
	TotalDonationProgramIncome  float64 `json:"totalDonationProgramIncome"`
	TotalSocialProgramIncome    float64 `json:"totalSocialProgramIncome"`
	TotalFosterChildrenIncome   float64 `json:"totalFosterChildrenIncome"`
}
