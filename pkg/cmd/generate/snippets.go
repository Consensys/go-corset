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

import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
`

// nolint
const javaTraceOf string = `
   /**
    * Construct a new trace which will be written to a given file.
    *
    * @param file File into which the trace will be written.  Observe any previous contents of this file will be lost.
    * @return Trace object to use for writing column data.
    *
    * @throws IOException If an I/O error occurs.
    */
   public static {class} of(RandomAccessFile file, List<ColumnHeader> rawHeaders, byte[] metadata) throws IOException {
      // Construct trace file header bytes
      byte[] header = constructTraceFileHeader(metadata);
      // Align headers according to register indices.
      ColumnHeader[] columnHeaders = alignHeaders(rawHeaders);
      // Determine file size
      long headerSize = determineColumnHeadersSize(columnHeaders) + header.length;
      long dataSize = determineColumnDataSize(columnHeaders);
      file.setLength(headerSize + dataSize);
      // Write headers
      writeHeaders(file,header,columnHeaders,headerSize);
      // Initialise buffers
      MappedByteBuffer[] buffers = initialiseByteBuffers(file,columnHeaders,headerSize);
      // Done
      return new {class}(buffers);
   }

  /**
   * Construct trace file header containing the given metadata bytes.
   *
   * @param metadata Metadata bytes to be embedded in the trace file.
   *
   * @return bytes making up the header.
   */
   private static byte[] constructTraceFileHeader(byte[] metadata) {
     ByteBuffer buffer = ByteBuffer.allocate(16 + metadata.length);
	 // File identifier
     buffer.put(new byte[]{'z','k','t','r','a','c','e','r'});
     // Major version
     buffer.putShort((short) 1);
     // Minor version
     buffer.putShort((short) 0);
     // Metadata length
     buffer.putInt(metadata.length);
     // Metadata
     buffer.put(metadata);
     // Done
     return buffer.array();
   }

  /**
   * Align headers ensures that the order in which columns are seen matches the order found in the trace schema.
   *
   * @param headers The headers to be aligned.
   * @return The aligned headers.
   */
   private static ColumnHeader[] alignHeaders(List<ColumnHeader> headers) {
     ColumnHeader[] alignedHeaders = new ColumnHeader[{ninputs}];
     //
     for(ColumnHeader header : headers) {
       alignedHeaders[header.register] = header;
     }
     //
     return alignedHeaders;
   }

   /**
    * Precompute the size of the trace file in order to memory map the buffers.
    *
    * @param headers Set of headers for the columns being written.
    * @return Number of bytes requires for the trace file header.
    */
   private static long determineColumnHeadersSize(ColumnHeader[] headers) {
      long nBytes = 4; // column count

      for (ColumnHeader header : headers) {
	    if(header != null) {
          nBytes += 2; // name length
          nBytes += header.name.length();
          nBytes += 1; // byte per element
          nBytes += 4; // element count
        }
      }

      return nBytes;
   }

   /**
    * Precompute the size of the trace file in order to memory map the buffers.
    *
    * @param headers Set of headers for the columns being written.
    * @return Number of bytes required for storing all column data, excluding the header.
    */
   private static long determineColumnDataSize(ColumnHeader[] headers) {
      long nBytes = 0;

      for (ColumnHeader header : headers) {
	    if(header != null) {
           nBytes += header.length * header.bytesPerElement;
        }
      }

      return nBytes;
   }

   /**
    * Write header information for the trace file.
    *
    * @param file Trace file being written.
    * @param header Trace file header
    * @param headers Column headers.
    * @param size Overall size of the header.
    */
   private static void writeHeaders(RandomAccessFile file, byte[] header, ColumnHeader[] headers, long size) throws IOException {
      final var buffer = file.getChannel().map(FileChannel.MapMode.READ_WRITE, 0, size);
      // Write trace file header
      buffer.put(header);
      // Write column count as uint32
      buffer.putInt(countHeaders(headers));
      // Write column headers one-by-one
      for(ColumnHeader h : headers) {
	    if(h != null) {
          buffer.putShort((short) h.name.length());
          buffer.put(h.name.getBytes());
          buffer.put((byte) h.bytesPerElement);
          buffer.putInt((int) h.length);
        }
      }
   }

   /**
    * Initialise one memory mapped byte buffer for each column to be written in the trace.
    * @param headers Set of headers for the columns being written.
    * @param headerSize Space required at start of trace file for header.
    * @return Buffer array with one entry per header.
    */
   private static MappedByteBuffer[] initialiseByteBuffers(RandomAccessFile file, ColumnHeader[] headers,
    long headerSize) throws IOException {
      MappedByteBuffer[] buffers = new MappedByteBuffer[{ninputs}];
      long offset = headerSize;
      for(int i=0;i<headers.length;i++) {
	    if(headers[i] != null) {
          // Determine size (in bytes) required to store all elements of this column.
          long length = headers[i].length * headers[i].bytesPerElement;
          // Preallocate space for this column.
          buffers[i] = file.getChannel().map(FileChannel.MapMode.READ_WRITE, offset, length);
          //
          offset += length;
        }
      }
      return buffers;
   }

   /**
    * Counter number of active (i.e. non-null) headers.  A header can be null if
    * it represents a column in a module which is not activated for this trace.
	*/
   private static int countHeaders(ColumnHeader[] headers) throws IOException {
     int count = 0;
	 for(ColumnHeader h : headers) {
	    if(h != null) { count++; }
	 }
     return count;
   }

   /**
    * ColumnHeader contains information about a given column in the resulting trace file.
    *
    * @param name Name of the column, as found in the trace file.
    * @param bytesPerElement Bytes required for each element in the column.
    */
   public record ColumnHeader(String name, int register, long bytesPerElement, long length) { }
`
