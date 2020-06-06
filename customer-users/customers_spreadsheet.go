package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/jedib0t/go-pretty/table"
	"google.golang.org/api/sheets/v4"
	"gopkg.in/AlecAivazis/survey.v1"
)

const (
	// This ID is unique to the spreadsheet containing the RefArch Questionnaire responses
	// To get this, first load the spreadsheet in the browser.
	// Then, the ID is the entry between `d` and `edit` in the URL.
	// E.g: https://docs.google.com/spreadsheets/d/SHEET_ID/edit
	gruntworkCustomersSpreadsheetID = "1vvUoSZxoGhWVQhyFbceRsFTbSi3jt-0MYKDgGBSt6Jc"

	// This is the spreadsheet sheet where the Company information is recorded for each registered customer. This is
	// indicated at the bottom of the Google spreadsheet.
	companiesSheetName       = "Companies"
	companiesSheetLastColumn = "H"

	// This is the spreadsheet sheet where the user information is recorded for each registered customer.
	usersSheetName       = "Users"
	usersSheetLastColumn = "L"

	// We start with the 2nd row of the first column when scanning for data to avoid the header row.
	startOfData = "A2"
)

// The following enum iotas represent the column organization of the company spreadsheet. Whenever we change the column
// representations, this enum should be updated.
// This should be exactly in order of the column headings in the spreadsheet. That is, the first defined const is the
// first column, the second const is the next column, and so on.
const (
	companySheetName = iota
	companySheetDateSubscribed
	companySheetMaxUsers
	companySheetSubscriptionType
	companySheetHasProSupport
	companySheetIsActive
)

// The following enum iotas represent the column organization of the user spreadsheet. Whenever we change the column
// representations, this enum should be updated.
// This should be exactly in order of the column headings in the spreadsheet. That is, the first defined const is the
// first column, the second const is the next column, and so on.
const (
	userSheetFirstName = iota
	userSheetLastName
	userSheetEmail
	userSheetGithubID
	userSheetCompany
	userSheetIsActive
)

type company struct {
	name             string
	maxUsers         int
	dateSubscribed   string
	subscriptionType string
	hasProSupport    bool
}

type user struct {
	firstName string
	lastName  string
	email     string
	githubID  string
}

// lookupUsers is the main routine for looking up authorized users of Gruntwork Customers via the spreadsheet. This
// will:
// - Lookup all the active companies using the spreadsheet
// - Prompt the user to select a company they wish to see information for
// - For the selected company, find all the active and inactive users from the spreadsheet
func lookupUsers(client *sheets.Service) error {
	logger := GetProjectLogger()

	// Get the responses from the google form
	logger.Info("Retrieving Client Companies from google sheet")
	companies, err := getCompanies(client)
	if err != nil {
		return err
	}
	logger.Info("Successfully retrieved companies from google sheet")

	selected, err := selectCompany(companies)
	if err != nil {
		return err
	}
	if selected == nil {
		logger.Error("Selected company is nil! This should never happen. There's probably a bug in this code!")
		return fmt.Errorf("Impossible error")
	}

	logger.Infof("Looking up authorized users for selected company %s", selected.name)
	activeUsers, inactiveUsers, err := getUsersForCompany(client, selected.name)
	if err != nil {
		return err
	}
	logger.Info("Successfully retrieved user info from google sheet")

	printCompanyInfo(*selected, activeUsers, inactiveUsers)

	return nil
}

func printCompanyInfo(selectedCompany company, activeUsers []user, inactiveUsers []user) {
	fmt.Println("Selected company:")
	companyTabWrt := table.NewWriter()
	companyTabWrt.SetOutputMirror(os.Stdout)
	companyTabWrt.AppendRows([]table.Row{
		{"Name", selectedCompany.name},
		{"Date Subscribed", selectedCompany.dateSubscribed},
		{"Active Users", strconv.Itoa(len(activeUsers))},
		{"Max Users", strconv.Itoa(selectedCompany.maxUsers)},
		{"Subscription Type", selectedCompany.subscriptionType},
		{"Pro support", fmt.Sprintf("%v", selectedCompany.hasProSupport)},
	})
	companyTabWrt.Render()
	fmt.Println()
	fmt.Println()

	userHeader := table.Row{"First Name", "Last Name", "Email", "Github ID"}
	fmt.Println("Active Users:")
	userTabWrt := table.NewWriter()
	userTabWrt.SetOutputMirror(os.Stdout)
	userTabWrt.AppendHeader(userHeader)
	activeUserRows := []table.Row{}
	for _, authorizedUser := range activeUsers {
		activeUserRows = append(activeUserRows, table.Row{
			authorizedUser.firstName,
			authorizedUser.lastName,
			authorizedUser.email,
			authorizedUser.githubID,
		})
	}
	userTabWrt.AppendRows(activeUserRows)
	userTabWrt.Render()
	fmt.Println()
	fmt.Println()

	fmt.Println("Inactive Users:")
	inactiveTabWrt := table.NewWriter()
	inactiveTabWrt.SetOutputMirror(os.Stdout)
	inactiveTabWrt.AppendHeader(userHeader)
	inactiveUserRows := []table.Row{}
	for _, authorizedUser := range inactiveUsers {
		inactiveUserRows = append(inactiveUserRows, table.Row{
			authorizedUser.firstName,
			authorizedUser.lastName,
			authorizedUser.email,
			authorizedUser.githubID,
		})
	}
	inactiveTabWrt.AppendRows(inactiveUserRows)
	inactiveTabWrt.Render()
}

func checkSheetBool(data string) bool {
	return strings.ToLower(data) == "yes"
}

func getUsersForCompany(client *sheets.Service, companyName string) ([]user, []user, error) {
	logger := GetProjectLogger()

	readRange := fmt.Sprintf("%s!%s:%s", usersSheetName, startOfData, usersSheetLastColumn)
	resp, err := client.Spreadsheets.Values.Get(gruntworkCustomersSpreadsheetID, readRange).Do()
	if err != nil {
		return nil, nil, errors.WithStackTrace(err)
	}

	activeUsers := []user{}
	inactiveUsers := []user{}
	for i, row := range resp.Values {
		// We need all the data about the user to proceed. Some rows do not contain all the data because it was
		// partially filled, skipped, or is in progress of being edited, so we do a quick check to make sure all the
		// data we expect exists.
		if len(row) < 5 {
			// Row count is i+2 to align with spreadsheet row count. Spreadsheet row count starts at 1 and we ignored
			// the header, so +2.
			logger.Warnf("Skipping malformed row %d: %v", i+2, row)
		} else if row[userSheetCompany].(string) == companyName {
			authorizedUser := user{
				firstName: row[userSheetFirstName].(string),
				lastName:  row[userSheetLastName].(string),
				email:     row[userSheetEmail].(string),
				githubID:  row[userSheetGithubID].(string),
			}
			if checkSheetBool(row[userSheetIsActive].(string)) {
				activeUsers = append(activeUsers, authorizedUser)
			} else {
				inactiveUsers = append(inactiveUsers, authorizedUser)
			}
		}
	}
	return activeUsers, inactiveUsers, nil
}

func selectCompany(companies []company) (*company, error) {
	companyNames := []string{}
	for _, company := range companies {
		companyNames = append(companyNames, company.name)
	}

	prompt := &survey.Select{
		Message:  "Which Company would you like to see information about?",
		Options:  companyNames,
		PageSize: 20,
	}
	selectedCompany := ""
	err := survey.AskOne(prompt, &selectedCompany, nil)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}
	for _, company := range companies {
		if company.name == selectedCompany {
			return &company, nil
		}
	}
	// TODO return error; we should never reach here.
	return nil, nil
}

func getCompanies(client *sheets.Service) ([]company, error) {
	logger := GetProjectLogger()

	readRange := fmt.Sprintf("%s!%s:%s", companiesSheetName, startOfData, companiesSheetLastColumn)
	resp, err := client.Spreadsheets.Values.Get(gruntworkCustomersSpreadsheetID, readRange).Do()
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	companies := []company{}
	for _, row := range resp.Values {
		companyName := row[companySheetName].(string)
		if checkSheetBool(row[companySheetIsActive].(string)) {
			maxUsers, err := strconv.Atoi(row[companySheetMaxUsers].(string))
			if err != nil {
				return nil, errors.WithStackTrace(err)
			}
			companies = append(
				companies,
				company{
					name:             companyName,
					maxUsers:         maxUsers,
					dateSubscribed:   row[companySheetDateSubscribed].(string),
					subscriptionType: row[companySheetSubscriptionType].(string),
					hasProSupport:    checkSheetBool(row[companySheetHasProSupport].(string)),
				},
			)
		} else {
			logger.Warnf("Skipping company %s : Not active", companyName)
		}
	}

	for _, comp := range companies {
		logger.Debugf("Company: %s ; Max Users: %d", comp.name, comp.maxUsers)
	}

	return companies, nil
}
