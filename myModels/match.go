package myModels

type Goal struct {
	Time          string `json:"time"`
	HomeScorer    string `json:"home_scorer"`
	HomeScorerID  string `json:"home_scorer_id"`
	HomeAssist    string `json:"home_assist"`
	HomeAssistID  string `json:"home_assist_id"`
	Score         string `json:"score"`
	AwayScorer    string `json:"away_scorer"`
	AwayScorerID  string `json:"away_scorer_id"`
	AwayAssist    string `json:"away_assist"`
	AwayAssistID  string `json:"away_assist_id"`
	Info          string `json:"info"`
	ScoreInfoTime string `json:"score_info_time"`
}

type Card struct {
	Time          string `json:"time"`
	HomeFault     string `json:"home_fault"`
	Card          string `json:"card"`
	AwayFault     string `json:"away_fault"`
	Info          string `json:"info"`
	HomePlayerID  string `json:"home_player_id"`
	AwayPlayerID  string `json:"away_player_id"`
	ScoreInfoTime string `json:"score_info_time"`
}

type SubMatch struct {
	MatchID                    string `json:"match_id"`
	CountryID                  string `json:"country_id"`
	CountryName                string `json:"country_name"`
	LeagueID                   string `json:"league_id"`
	LeagueName                 string `json:"league_name"`
	MatchDate                  string `json:"match_date"`
	MatchStatus                string `json:"match_status"`
	MatchTime                  string `json:"match_time"`
	MatchHometeamID            string `json:"match_hometeam_id"`
	MatchHometeamName          string `json:"match_hometeam_name"`
	MatchHometeamScore         string `json:"match_hometeam_score"`
	MatchAwayteamName          string `json:"match_awayteam_name"`
	MatchAwayteamID            string `json:"match_awayteam_id"`
	MatchAwayteamScore         string `json:"match_awayteam_score"`
	MatchHometeamHalftimeScore string `json:"match_hometeam_halftime_score"`
	MatchAwayteamHalftimeScore string `json:"match_awayteam_halftime_score"`
	MatchHometeamExtraScore    string `json:"match_hometeam_extra_score"`
	MatchAwayteamExtraScore    string `json:"match_awayteam_extra_score"`
	MatchHometeamPenaltyScore  string `json:"match_hometeam_penalty_score"`
	MatchAwayteamPenaltyScore  string `json:"match_awayteam_penalty_score"`
	MatchHometeamFtScore       string `json:"match_hometeam_ft_score"`
	MatchAwayteamFtScore       string `json:"match_awayteam_ft_score"`
	MatchHometeamSystem        string `json:"match_hometeam_system"`
	MatchAwayteamSystem        string `json:"match_awayteam_system"`
	MatchLive                  string `json:"match_live"`
	MatchRound                 string `json:"match_round"`
	MatchStadium               string `json:"match_stadium"`
	MatchReferee               string `json:"match_referee"`
	TeamHomeBadge              string `json:"team_home_badge"`
	TeamAwayBadge              string `json:"team_away_badge"`
	LeagueLogo                 string `json:"league_logo"`
	CountryLogo                string `json:"country_logo"`
	LeagueYear                 string `json:"league_year"`
	FkStageKey                 string `json:"fk_stage_key"`
	StageName                  string `json:"stage_name"`
	Goalscorer                 []Goal `json:"goalscorer"`
	Cards                      []Card `json:"cards"`
}
