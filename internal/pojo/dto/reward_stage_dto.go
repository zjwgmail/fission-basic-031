package dto

type RewardStageDto struct {
	StageNum         int
	CurrentStageMax  int
	CurrentStageName string
	CurrentAwardLink string
	NextStageMax     int
	NextStageName    string
	NextAwardLink    string
}
