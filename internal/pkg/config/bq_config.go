/**
 * Copyright 2021 Napptive
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package config

import bqconfig "github.com/napptive/analytics/pkg/config"

type BQConfig struct {
	// enabled contains a flag to indicate if analytics is enabled or not
	Enabled bool
	// BigQueryConfig contains the configuration to connect to BigQuery
	Config bqconfig.BigQueryConfig
}

func (bq *BQConfig) IsValid () error {
	if bq.Enabled {
		return bq.Config.IsValid()
	}
	return nil
}

func (bq *BQConfig) Print () {
	if bq.Enabled {
		bq.Config.Print()
	}
}