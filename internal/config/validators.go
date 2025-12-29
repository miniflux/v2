// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config // import "miniflux.app/v2/internal/config"

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

func validateChoices(rawValue string, choices []string) error {
	if !slices.Contains(choices, rawValue) {
		return fmt.Errorf("value must be one of: %v", strings.Join(choices, ", "))
	}
	return nil
}

func validateListChoices(inputValues, choices []string) error {
	for _, value := range inputValues {
		if err := validateChoices(value, choices); err != nil {
			return err
		}
	}
	return nil
}

func validateGreaterThan(rawValue string, min int) error {
	intValue, err := strconv.Atoi(rawValue)
	if err != nil {
		return errors.New("value must be an integer")
	}
	if intValue > min {
		return nil
	}
	return fmt.Errorf("value must be at least %d", min)
}

func validateGreaterOrEqualThan(rawValue string, min int) error {
	intValue, err := strconv.Atoi(rawValue)
	if err != nil {
		return errors.New("value must be an integer")
	}
	if intValue >= min {
		return nil
	}
	return fmt.Errorf("value must be greater or equal than %d", min)
}

func validateRange(rawValue string, min, max int) error {
	intValue, err := strconv.Atoi(rawValue)
	if err != nil {
		return errors.New("value must be an integer")
	}
	if intValue < min || intValue > max {
		return fmt.Errorf("value must be between %d and %d", min, max)
	}
	return nil
}
