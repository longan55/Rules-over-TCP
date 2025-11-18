/*
* Copyright 2025-2026 longan55 or authors.
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
package rot

type Cipher interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

var _ Cipher = (*CryptNothing)(nil)

type CryptNothing struct{}

func (c *CryptNothing) Encrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (c *CryptNothing) Decrypt(data []byte) ([]byte, error) {
	return data, nil
}

type CryptConfig struct {
	cryptMap map[int]Cipher
}

func NewCryptConfig() *CryptConfig {
	return &CryptConfig{
		cryptMap: map[int]Cipher{
			0: &CryptNothing{},
		},
	}
}

func (cc *CryptConfig) AddCrypt(flag int, crypt Cipher) {
	cc.cryptMap[flag] = crypt
}
