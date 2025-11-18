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

import "testing"

func TestXxx(t *testing.T) {
	//192 167
	//0xC0 0xA7
	//0x00 0x01 0x7f 0xff 0xff 0xff 0x80 0x00 0x00 0x00 0x12 0x34 0x12 0x34 0x01
	data := []byte{0x00, 0x01, 0x7F, 0xFF, 0xFF, 0xFF, 0x80, 0x00, 0x00, 0x00, 0x12, 0x34, 0x12, 0x34, 0x01, 0x41, 0xC6, 0x1a, 0x40}
	crc := ModBusCRC(data)
	t.Logf("校验码:% #x\n", crc)
	data1 := []byte{0x00, 0x02, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78}
	crc1 := ModBusCRC(data1)
	t.Logf("校验码:% #x\n", crc1)
	data2 := []byte{0x00, 0x03, 0x30, 0x31, 0x32, 0x33}
	crc2 := ModBusCRC(data2)
	t.Logf("校验码:% #x\n", crc2)
}
