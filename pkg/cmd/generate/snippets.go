// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package generate

const license string = `/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
 `

const javaWarning string = `
/**
 * WARNING: This code is generated automatically.
 *
 * <p>Any modifications to this code may be overwritten and could lead to unexpected behavior.
 * Please DO NOT ATTEMPT TO MODIFY this code directly</p>.
 *
`

const javaImports string = `
import java.io.IOException;
import java.io.RandomAccessFile;
import java.math.BigInteger;
import java.nio.ByteBuffer;
import java.nio.MappedByteBuffer;
import java.nio.channels.FileChannel;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.ObjectWriter;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
`

// nolint
const javaColumnHeader string = `
   /**
    * ColumnHeader contains information about a given column in the resulting trace file.
    *
    * @param name Name of the column, as found in the trace file.
    * @param bitwidth Max number of bits required for any element in the column.
    */
   public record ColumnHeader(String name, int register, int bitwidth, int length) { }
`

// nolint
const javaColumn string = `
   /**
    * Column provides an interface for writing column data into the resulting trace file.
    *
    */
   public interface Column {
      public void write(boolean value);
      public void write(long value);
      public void write(byte[] value);
   }
`

// nolint
const javaAddMetadataSignature string = `
  /**
   * Get static metadata stored within this trace during compilation.
   */
  public Map<String,Object> getMetaData();
`

// nolint
const javaOpenSignature string = `
  /**
   * Open this trace file for the given set of columns.
   */
  public void open(Column[] columns);
`
