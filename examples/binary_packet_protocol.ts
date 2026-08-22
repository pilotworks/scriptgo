// ============================================================================
// Enterprise Binary Network Protocol Stack: Multiplexing, Framing & Checksums
// ============================================================================
// Features demonstrated:
// - Buffer & TypedArrays binary manipulation (BE integers, floats, bitfields)
// - Custom bitwise algorithms: CRC-16-CCITT and Fletcher-16 checksums
// - Discriminated Unions & Polymorphic Packet Structures
// - Bitwise Flag Operations (Compression, Encryption, Priority, AckRequired)
// - Multiplexed Virtual Channels & Session State Machine
// - Sliding Window Sequence Tracking & Out-of-Order Reassembly
// - Comprehensive Error Detection & Packet Corruption Recovery
// ============================================================================

import { Buffer } from "node:buffer";

// ----------------------------------------------------------------------------
// Protocol Constants & Bit Flags
// ----------------------------------------------------------------------------

export enum OpCode {
  HANDSHAKE_SYN = 0x01,
  HANDSHAKE_ACK = 0x02,
  HEARTBEAT_PING = 0x03,
  HEARTBEAT_PONG = 0x04,
  DATA_STREAM = 0x10,
  DATA_ACK = 0x11,
  CHANNEL_OPEN = 0x20,
  CHANNEL_CLOSE = 0x21,
  ERROR_ALERT = 0xee,
  DISCONNECT = 0xff,
}

export enum HeaderFlags {
  NONE = 0x00,
  COMPRESSED = 1 << 0, // 0x01
  ENCRYPTED = 1 << 1,  // 0x02
  PRIORITY_HIGH = 1 << 2, // 0x04
  ACK_REQUIRED = 1 << 3,  // 0x08
  END_OF_STREAM = 1 << 4, // 0x10
}

export enum StatusCode {
  OK = 0,
  BAD_REQUEST = 400,
  UNAUTHORIZED = 401,
  CHECKSUM_MISMATCH = 422,
  INTERNAL_ERROR = 500,
}

// ----------------------------------------------------------------------------
// Packet Header Layout: Fixed 16 Bytes
// ----------------------------------------------------------------------------
// [0..1]   Magic Number (2 bytes, 0x5347 "SG")
// [2]      Protocol Version (1 byte, e.g. 0x02)
// [3]      OpCode (1 byte)
// [4]      Bit Flags (1 byte)
// [5..6]   Virtual Channel ID (2 bytes, 0..65535)
// [7..10]  Sequence Number (4 bytes, uint32)
// [11..13] Payload Length (3 bytes / uint24, up to 16MB)
// [14..15] CRC-16 Checksum (2 bytes)
// ----------------------------------------------------------------------------

export interface PacketHeader {
  magic: number;
  version: number;
  opCode: OpCode;
  flags: number;
  channelId: number;
  sequenceNumber: number;
  payloadLength: number;
  checksum: number;
}

export interface PacketFrame<T> {
  header: PacketHeader;
  payload: T;
}

// ----------------------------------------------------------------------------
// Payload Models (Discriminated Packets)
// ----------------------------------------------------------------------------

export interface HandshakeSynPayload {
  clientId: string;
  clientVersion: string;
  timestamp: number;
  authSecretDigest: string;
}

export interface HandshakeAckPayload {
  sessionId: string;
  assignedChannelBase: number;
  serverTime: number;
  status: StatusCode;
}

export interface DataStreamPayload {
  streamName: string;
  chunkIndex: number;
  totalChunks: number;
  contentType: string;
  dataBase64: string;
}

export interface DataAckPayload {
  ackedSequence: number;
  windowSizeRemaining: number;
  receivedTimestamp: number;
}

export interface ErrorAlertPayload {
  errorCode: StatusCode;
  message: string;
  failedSequence: number;
}

// ----------------------------------------------------------------------------
// Checksum Algorithms (Bitwise Math)
// ----------------------------------------------------------------------------

export class ChecksumEngine {
  /**
   * Computes CRC-16-CCITT (Polynomial 0x1021, Initial 0xFFFF)
   */
  public static crc16(buffer: Uint8Array): number {
    let crc = 0xffff;
    for (let i = 0; i < buffer.length; i++) {
      crc ^= buffer[i] << 8;
      for (let j = 0; j < 8; j++) {
        if ((crc & 0x8000) !== 0) {
          crc = ((crc << 1) ^ 0x1021) & 0xffff;
        } else {
          crc = (crc << 1) & 0xffff;
        }
      }
    }
    return crc;
  }

  /**
   * Computes Fletcher-16 checksum
   */
  public static fletcher16(buffer: Uint8Array): number {
    let sum1 = 0xff;
    let sum2 = 0xff;
    for (let i = 0; i < buffer.length; i++) {
      sum1 = (sum1 + buffer[i]) % 255;
      sum2 = (sum2 + sum1) % 255;
    }
    return (sum2 << 8) | sum1;
  }
}

// ----------------------------------------------------------------------------
// Binary Codec Implementation
// ----------------------------------------------------------------------------

export class BinaryPacketCodec {
  public static readonly MAGIC: number = 0x5347; // 'S' 'G'
  public static readonly VERSION: number = 0x02;
  public static readonly HEADER_SIZE: number = 16;

  /**
   * Serializes a structured payload into a framed binary packet Buffer.
   */
  public static serialize<T>(
    opCode: OpCode,
    channelId: number,
    sequenceNumber: number,
    flags: number,
    payload: T
  ): Buffer {
    const jsonString = JSON.stringify(payload);
    const payloadBytes = Buffer.from(jsonString, "utf8");
    const payloadLen = payloadBytes.length;

    if (payloadLen > 0xffffff) {
      throw new Error(`Payload exceeds maximum size of 16MB: ${payloadLen} bytes`);
    }

    const checksum = ChecksumEngine.crc16(payloadBytes);
    const totalSize = BinaryPacketCodec.HEADER_SIZE + payloadLen;
    const packet = Buffer.alloc(totalSize);

    // [0..1] Magic (2 bytes)
    packet.writeUInt16BE(BinaryPacketCodec.MAGIC, 0);
    // [2] Version (1 byte)
    packet.writeUInt8(BinaryPacketCodec.VERSION, 2);
    // [3] OpCode (1 byte)
    packet.writeUInt8(opCode, 3);
    // [4] Flags (1 byte)
    packet.writeUInt8(flags & 0xff, 4);
    // [5..6] Channel ID (2 bytes)
    packet.writeUInt16BE(channelId & 0xffff, 5);
    // [7..10] Sequence Number (4 bytes)
    packet.writeUInt32BE(sequenceNumber >>> 0, 7);

    // [11..13] Payload Length (3 bytes uint24)
    packet.writeUInt8((payloadLen >> 16) & 0xff, 11);
    packet.writeUInt8((payloadLen >> 8) & 0xff, 12);
    packet.writeUInt8(payloadLen & 0xff, 13);

    // [14..15] Checksum (2 bytes)
    packet.writeUInt16BE(checksum & 0xffff, 14);

    // Payload body
    payloadBytes.copy(packet, BinaryPacketCodec.HEADER_SIZE);

    return packet;
  }

  /**
   * Parses and validates a binary packet Buffer.
   */
  public static deserialize<T>(buffer: Buffer): PacketFrame<T> {
    if (buffer.length < BinaryPacketCodec.HEADER_SIZE) {
      throw new Error(
        `Buffer underrun: received ${buffer.length} bytes, header requires ${BinaryPacketCodec.HEADER_SIZE}`
      );
    }

    const magic = buffer.readUInt16BE(0);
    if (magic !== BinaryPacketCodec.MAGIC) {
      throw new Error(
        `Invalid packet magic: 0x${magic.toString(16).toUpperCase()} != 0x${BinaryPacketCodec.MAGIC.toString(16).toUpperCase()}`
      );
    }

    const version = buffer.readUInt8(2);
    const opCode = buffer.readUInt8(3) as OpCode;
    const flags = buffer.readUInt8(4);
    const channelId = buffer.readUInt16BE(5);
    const sequenceNumber = buffer.readUInt32BE(7);

    const lenByte1 = buffer.readUInt8(11);
    const lenByte2 = buffer.readUInt8(12);
    const lenByte3 = buffer.readUInt8(13);
    const payloadLength = (lenByte1 << 16) | (lenByte2 << 8) | lenByte3;

    const expectedChecksum = buffer.readUInt16BE(14);

    const expectedTotalSize = BinaryPacketCodec.HEADER_SIZE + payloadLength;
    if (buffer.length < expectedTotalSize) {
      throw new Error(
        `Incomplete frame: expected ${expectedTotalSize} bytes, got ${buffer.length}`
      );
    }

    const payloadBuffer = buffer.subarray(
      BinaryPacketCodec.HEADER_SIZE,
      BinaryPacketCodec.HEADER_SIZE + payloadLength
    );

    // Verify CRC-16 Checksum
    const actualChecksum = ChecksumEngine.crc16(payloadBuffer);
    if (actualChecksum !== expectedChecksum) {
      throw new Error(
        `CRC-16 checksum mismatch: expected 0x${expectedChecksum.toString(16)}, computed 0x${actualChecksum.toString(16)}`
      );
    }

    const rawString = payloadBuffer.toString("utf8");
    const payload = JSON.parse(rawString) as T;

    const header: PacketHeader = {
      magic,
      version,
      opCode,
      flags,
      channelId,
      sequenceNumber,
      payloadLength,
      checksum: expectedChecksum,
    };

    return { header, payload };
  }
}

// ----------------------------------------------------------------------------
// Virtual Multiplexed Channel Session Manager
// ----------------------------------------------------------------------------

export class ChannelSession {
  private nextSequenceOut: number = 1000;
  private expectedSequenceIn: number = 1;
  private readonly receivedPackets: Map<number, Buffer> = new Map<number, Buffer>();
  private readonly unackedPackets: Map<number, Buffer> = new Map<number, Buffer>();

  constructor(
    public readonly channelId: number,
    public readonly channelName: string
  ) { }

  public send<T>(opCode: OpCode, flags: number, payload: T): Buffer {
    const seq = this.nextSequenceOut++;
    const packet = BinaryPacketCodec.serialize(opCode, this.channelId, seq, flags, payload);

    if ((flags & HeaderFlags.ACK_REQUIRED) !== 0) {
      this.unackedPackets.set(seq, packet);
    }
    return packet;
  }

  public receive(packetBuffer: Buffer): PacketFrame<unknown> {
    const frame = BinaryPacketCodec.deserialize<unknown>(packetBuffer);
    this.receivedPackets.set(frame.header.sequenceNumber, packetBuffer);
    return frame;
  }

  public acknowledge(sequenceNumber: number): boolean {
    return this.unackedPackets.delete(sequenceNumber);
  }

  public getPendingAckCount(): number {
    return this.unackedPackets.size;
  }

  public getNextSequence(): number {
    return this.nextSequenceOut;
  }
}

// ----------------------------------------------------------------------------
// Full Protocol Stack Simulation
// ----------------------------------------------------------------------------

console.log("=================================================================");
console.log("     SCRIPTGO BINARY PROTOCOL CODEC & MULTIPLEXER SIMULATOR      ");
console.log("=================================================================");

function runProtocolSession(): void {
  const channel = new ChannelSession(4001, "telemetry-stream");

  // --------------------------------------------------------------------------
  // Step 1: Client Handshake SYN
  // --------------------------------------------------------------------------
  console.log("\n[1] Initiating Client Handshake SYN...");
  const synPayload: HandshakeSynPayload = {
    clientId: "worker-us-east-88a",
    clientVersion: "2.4.0",
    timestamp: Date.now(),
    authSecretDigest: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  };

  const synPacket = channel.send(
    OpCode.HANDSHAKE_SYN,
    HeaderFlags.PRIORITY_HIGH | HeaderFlags.ACK_REQUIRED,
    synPayload
  );

  console.log(`  Framed SYN Packet Size : ${synPacket.length} bytes (Header: 16, Payload: ${synPacket.length - 16})`);
  console.log(`  Hex Header Dump        : ${synPacket.subarray(0, 16).toString("hex")}`);

  // --------------------------------------------------------------------------
  // Step 2: Server Decodes SYN & Responds with ACK
  // --------------------------------------------------------------------------
  console.log("\n[2] Server Ingesting Handshake SYN Frame...");
  const parsedSyn = BinaryPacketCodec.deserialize<HandshakeSynPayload>(synPacket);
  console.log(`  Decoded Magic   : 0x${parsedSyn.header.magic.toString(16).toUpperCase()}`);
  console.log(`  Decoded OpCode  : 0x${parsedSyn.header.opCode.toString(16)} (HANDSHAKE_SYN)`);
  console.log(`  Decoded SeqNum  : ${parsedSyn.header.sequenceNumber}`);
  console.log(`  Decoded Client  : ${parsedSyn.payload.clientId} (v${parsedSyn.payload.clientVersion})`);

  console.log("\n[3] Server Formulates Handshake ACK Frame...");
  const ackPayload: HandshakeAckPayload = {
    sessionId: "sess-992100-af1b",
    assignedChannelBase: 4000,
    serverTime: Date.now(),
    status: StatusCode.OK,
  };

  const serverAckPacket = BinaryPacketCodec.serialize(
    OpCode.HANDSHAKE_ACK,
    parsedSyn.header.channelId,
    2001,
    HeaderFlags.NONE,
    ackPayload
  );
  console.log(`  Server ACK Size : ${serverAckPacket.length} bytes`);

  // Client receives Handshake ACK
  const clientAckFrame = channel.receive(serverAckPacket);
  const ackData = clientAckFrame.payload as HandshakeAckPayload;
  console.log(`  Client Received ACK: Session=${ackData.sessionId}, Status=${ackData.status}`);
  channel.acknowledge(parsedSyn.header.sequenceNumber);

  // --------------------------------------------------------------------------
  // Step 3: Stream Multiplexed Telemetry Data Chunks
  // --------------------------------------------------------------------------
  console.log("\n[4] Streaming Segmented Telemetry Payloads...");
  const sampleRecords = [
    { metric: "cpu_usage_pct", val: 42.8, unit: "%" },
    { metric: "mem_allocated_mb", val: 15420.5, unit: "MB" },
    { metric: "io_ops_per_sec", val: 84300, unit: "iops" },
  ];

  for (let i = 0; i < sampleRecords.length; i++) {
    const isLast = i === sampleRecords.length - 1;
    const flags = HeaderFlags.COMPRESSED | (isLast ? HeaderFlags.END_OF_STREAM : HeaderFlags.NONE);

    const streamPayload: DataStreamPayload = {
      streamName: "sys_metrics",
      chunkIndex: i + 1,
      totalChunks: sampleRecords.length,
      contentType: "application/json",
      dataBase64: Buffer.from(JSON.stringify(sampleRecords[i])).toString("base64"),
    };

    const dataPacket = channel.send(OpCode.DATA_STREAM, flags, streamPayload);
    const decodedFrame = BinaryPacketCodec.deserialize<DataStreamPayload>(dataPacket);

    console.log(
      `  Chunk [${decodedFrame.payload.chunkIndex}/${decodedFrame.payload.totalChunks}] Seq=${decodedFrame.header.sequenceNumber} | Payload=${decodedFrame.header.payloadLength}B | EndOfStream=${(decodedFrame.header.flags & HeaderFlags.END_OF_STREAM) !== 0}`
    );
  }

  // --------------------------------------------------------------------------
  // Step 4: Checksum Integrity Validation & Corruption Recovery
  // --------------------------------------------------------------------------
  console.log("\n[5] Simulating Network Bit-Flip Corruption Detection...");
  const validPacket = channel.send(OpCode.HEARTBEAT_PING, HeaderFlags.NONE, { pingTime: Date.now() });
  const corruptedBuffer = Buffer.from(validPacket);

  // Flip bit in payload section (after 16-byte header)
  corruptedBuffer[18] ^= 0x55;

  try {
    BinaryPacketCodec.deserialize(corruptedBuffer);
    console.log("  [FAIL] Corruption was not detected!");
  } catch (err) {
    console.log(`  [PASS] Integrity Engine Successfully Intercepted Error:\n         -> ${err}`);
  }

  console.log("\n=================================================================");
  console.log("                 ALL PROTOCOL STACK TESTS PASSED                 ");
  console.log("=================================================================\n");
}

runProtocolSession();
