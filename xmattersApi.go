package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GroupMembershipResult struct {
	Count int `json:"count"`
	Total int `json:"total"`
	Data  []struct {
		Group struct {
			ID            string `json:"id"`
			TargetName    string `json:"targetName"`
			RecipientType string `json:"recipientType"`
			GroupType     string `json:"groupType"`
			Links         struct {
				Self string `json:"self"`
			} `json:"links"`
		} `json:"group"`
		Member struct {
			ID            string `json:"id"`
			TargetName    string `json:"targetName"`
			FirstName     string `json:"firstName"`
			LastName      string `json:"lastName"`
			RecipientType string `json:"recipientType"`
			Links         struct {
				Self string `json:"self"`
			} `json:"links"`
		} `json:"member"`
	} `json:"data"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
}

type OnCallData struct {
	Group struct {
		ID            string `json:"id"`
		TargetName    string `json:"targetName"`
		RecipientType string `json:"recipientType"`
		GroupType     string `json:"groupType"`
		Links         struct {
			Self string `json:"self"`
		} `json:"links"`
	} `json:"group"`
	Shift struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		SiteHolidays struct {
			OnCall bool   `json:"onCall"`
			Start  string `json:"start"`
			End    string `json:"end"`
		} `json:"siteHolidays"`
		Links struct {
			Self string `json:"self"`
		} `json:"links"`
	} `json:"shift"`
	Members struct {
		Count int `json:"count"`
		Total int `json:"total"`
		Data  []struct {
			Position       int    `json:"position"`
			Delay          int    `json:"delay"`
			EscalationType string `json:"escalationType"`
			InRotation     bool   `json:"inRotation"`
			Replacements   struct {
				Count int `json:"count"`
				Total int `json:"total"`
				Data  []struct {
					Start       string `json:"start"`
					End         string `json:"end"`
					Replacement struct {
						ID            string `json:"id"`
						TargetName    string `json:"targetName"`
						RecipientType string `json:"recipientType"`
						Links         struct {
							Self string `json:"self"`
						} `json:"links"`
						FirstName string `json:"firstName"`
						LastName  string `json:"lastName"`
						Status    string `json:"status"`
					} `json:"replacement"`
				} `json:"data"`
			} `json:"replacements"`
			Member struct {
				ID              string `json:"id"`
				TargetName      string `json:"targetName"`
				RecipientType   string `json:"recipientType"`
				ExternallyOwned bool   `json:"externallyOwned"`
				ExternalKey     string `json:"externalKey"`
				Links           struct {
					Self string `json:"self"`
				} `json:"links"`
				FirstName   string `json:"firstName"`
				LastName    string `json:"lastName"`
				LicenseType string `json:"licenseType"`
				Language    string `json:"language"`
				Timezone    string `json:"timezone"`
				WebLogin    string `json:"webLogin"`
				Site        struct {
					ID    string `json:"id"`
					Name  string `json:"name"`
					Links struct {
						Self string `json:"self"`
					} `json:"links"`
				} `json:"site"`
				LastLogin   string `json:"lastLogin"`
				WhenCreated string `json:"whenCreated"`
				WhenUpdated string `json:"whenUpdated"`
				Status      string `json:"status"`
			} `json:"member"`
		} `json:"data"`
		Links struct {
			Self string `json:"self"`
			Next string `json:"next"`
		} `json:"links"`
	} `json:"members"`
	NotifyEndOfEscalation struct {
		NotifyEnabled bool `json:"notifyEnabled"`
	} `json:"notifyEndOfEscalation"`
	Start     string `json:"start"`
	End       string `json:"end"`
	Replacing string
}

type OnCallResult struct {
	Count int          `json:"count"`
	Total int          `json:"total"`
	Data  []OnCallData `json:"data"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
}

func CallAPI[V any](url string) (V, error) {

	var result V

	apiKey, err := GetApiKey()
	if err != nil {
		return result, errors.Join(errors.New("error getting API key"), err)
	}

	apiSecret, err := GetApiSecret()
	if err != nil {
		return result, errors.Join(err, errors.New("error getting API secret"))
	}

	domain, err := GetXMattersDomain()
	if err != nil {
		return result, errors.Join(err, errors.New("error getting xmatters domain"))
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s%s", domain, url), nil)
	if err != nil {
		return result, errors.Join(errors.New("error creating request"), err)
	}

	req.SetBasicAuth(apiKey, apiSecret)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, errors.Join(errors.New("error making request"), err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body")
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return result, errors.New("received non-200 response code")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return result, errors.Join(errors.New("error reading response body"), err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, errors.Join(errors.New("error parsing JSON response"), err)
	}

	return result, nil
}

func GetMySchedule(username string) (*OnCallResult, error) {
	groupData, err := CallAPI[GroupMembershipResult](fmt.Sprintf("/api/xm/1/people/%s/group-memberships", username))
	if err != nil {
		return nil, err
	}

	onCallData, err := CallAPI[OnCallResult](fmt.Sprintf(
		"/api/xm/1/on-call?groups=%s&from=%sZ&to=%sZ",
		groupData.Data[0].Group.ID,
		time.Now().Format("2006-01-02T15:04:05"),
		time.Now().AddDate(0, 3, 0).Format("2006-01-02T15:04:05"),
	))

	if err != nil {
		return nil, err
	}

	findMyShift := func(shift *OnCallData) bool {

		// If no data
		if len(shift.Members.Data) == 0 {
			return false
		}

		// If me directly
		if strings.ToUpper(shift.Members.Data[0].Member.ExternalKey) == strings.ToUpper(username) {

			// Someone else is replacing me
			if shift.Members.Data[0].Replacements.Count > 0 && strings.ToUpper(shift.Members.Data[0].Replacements.Data[0].Replacement.TargetName) != strings.ToUpper(username) {
				return false
			}

			return true
		}

		// If no replacements
		if shift.Members.Data[0].Replacements.Count == 0 {
			return false
		}

		// If I'm replacing
		if strings.ToUpper(shift.Members.Data[0].Replacements.Data[0].Replacement.TargetName) == strings.ToUpper(username) {
			shift.Replacing = fmt.Sprintf("%s %s", shift.Members.Data[0].Member.FirstName, shift.Members.Data[0].Member.LastName)
			return true
		}

		return false
	}
	onCallData.Data = filter(onCallData.Data, findMyShift)
	onCallData.Count = len(onCallData.Data)
	onCallData.Total = len(onCallData.Data)

	return &onCallData, nil
}

func filter[T any](ss []T, test func(*T) bool) (ret []T) {
	for _, s := range ss {
		if test(&s) {
			ret = append(ret, s)
		}
	}
	return
}
