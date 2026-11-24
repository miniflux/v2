// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package date // import "miniflux.app/v2/internal/reader/date"

import "strings"

// replacer rewrites the parts of a date that no layout accepts as written, such
// as localized day and month names, into their English form. When several pairs
// match at the same position, the first one listed wins.
var replacer = strings.NewReplacer(
	// Time zone notations.
	"GMT+0000 (Coordinated Universal Time)", "GMT",
	"GMT-", "GMT -",

	// Go only parses AM and PM.
	"a.m.", "AM",
	"p.m.", "PM",

	// English day abbreviations longer than three letters.
	"Thurs,", "Thu,",
	"Thur,", "Thu,",

	// German day names.
	"Mo,", "Mon,",
	"Di,", "Tue,",
	"Mi,", "Wed,",
	"Do,", "Thu,",
	"Fr,", "Fri,",
	"Sa,", "Sat,",
	"So,", "Sun,",

	// German month names.
	"Mär ", "Mar ",
	"Mai ", "May ",
	"Okt ", "Oct ",
	"Dez ", "Dec ",

	// French day names, abbreviated with a comma.
	"lun,", "Mon,",
	"mar,", "Tue,",
	"mer,", "Wed,",
	"jeu,", "Thu,",
	"ven,", "Fri,",
	"sam,", "Sat,",
	"dim,", "Sun,",

	// French day names, abbreviated with a period.
	"lun.", "Mon",
	"mar.", "Tue",
	"mer.", "Wed",
	"jeu.", "Thu",
	"ven.", "Fri",
	"sam.", "Sat",
	"dim.", "Sun",

	// French day names in full.
	"Lundi,", "Monday,",
	"Mardi,", "Tuesday,",
	"Mercredi,", "Wednesday,",
	"Jeudi,", "Thursday,",
	"Vendredi,", "Friday,",
	"Samedi,", "Saturday,",
	"Dimanche,", "Sunday,",

	// French month names, abbreviated.
	"jan.", "January ",
	"feb.", "February ",
	"mars.", "March ",
	"avril.", "April ",
	"mai.", "May ",
	"juin.", "June ",
	"juil.", "July",
	"août.", "August",
	"sept.", "September",
	"oct.", "October",
	"nov.", "November",
	"dec.", "December",
	"déc.", "December",
	"avr ", "Apr ",
	"jui ", "Jun ",

	// French month names in full, in lower case.
	"janvier ", "January ",
	"février ", "February ",
	"mars ", "March ",
	"avril ", "April ",
	"mai ", "May ",
	"juin ", "June ",
	"juillet ", "July ",
	"août ", "August ",
	"septembre ", "September ",
	"octobre ", "October ",
	"novembre ", "November ",
	"décembre ", "December ",

	// French month names in full, capitalized.
	"Janvier", "January",
	"Février", "February",
	"Mars", "March",
	"Avril", "April",
	"Mai", "May",
	"Juin", "June",
	"Juillet", "July",
	"Août", "August",
	"Septembre", "September",
	"Octobre", "October",
	"Novembre", "November",
	"Décembre", "December",
)
