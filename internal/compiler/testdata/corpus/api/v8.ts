import {
    cachedDataVersionTag,
    getHeapStatistics,
    getHeapSpaceStatistics,
    getHeapCodeStatistics,
    getCppHeapStatistics,
    getHeapSnapshot,
    writeHeapSnapshot,
    setFlagsFromString,
    queryObjects,
    stopCoverage,
    takeCoverage,
    setHeapSnapshotNearHeapLimit,
    isStringOneByteRepresentation,
    GCProfiler,
    Serializer,
    Deserializer,
    DefaultSerializer,
    DefaultDeserializer,
    serialize,
    deserialize,
    onInit,
    onSettled,
    onBefore,
    onAfter,
    createHook,
    init,
    before,
    after,
    settled,
    addSerializeCallback,
    addDeserializeCallback,
    setDeserializeMainFunction,
    isBuildingSnapshot
} from "node:v8";

// @api: v8.cachedDataVersionTag
// @expect: v8_cachedDataVer: 1
console.log("v8_cachedDataVer: " + cachedDataVersionTag());

// @api: v8.getHeapStatistics
// @expect: v8_heapStats_total: 1048576
console.log("v8_heapStats_total: " + getHeapStatistics().total_heap_size);

// @api: v8.getHeapSpaceStatistics
// @expect: v8_heapSpace_name: new_space
console.log("v8_heapSpace_name: " + getHeapSpaceStatistics()[0].space_name);

// @api: v8.getHeapCodeStatistics
// @expect: v8_heapCode_size: 0
console.log("v8_heapCode_size: " + getHeapCodeStatistics().code_and_metadata_size);

// @api: v8.getCppHeapStatistics
// @expect: v8_cppHeapStats: true
console.log("v8_cppHeapStats: " + (typeof getCppHeapStatistics() === "object"));

// @api: v8.getHeapSnapshot
// @expect: v8_getHeapSnapshot: true
console.log("v8_getHeapSnapshot: " + (typeof getHeapSnapshot() === "object"));

// @api: v8.writeHeapSnapshot
// @expect: v8_writeHeapSnapshot: true
console.log("v8_writeHeapSnapshot: " + (writeHeapSnapshot().length > 0));

// @api: v8.setFlagsFromString
// @expect: v8_setFlags: true
setFlagsFromString("--trace-gc");
console.log("v8_setFlags: true");

// @api: v8.queryObjects
// @expect: v8_queryObjects: 0
console.log("v8_queryObjects: " + queryObjects({}));

// @api: v8.stopCoverage
// @expect: v8_stopCoverage: true
stopCoverage();
console.log("v8_stopCoverage: true");

// @api: v8.takeCoverage
// @expect: v8_takeCoverage: true
takeCoverage();
console.log("v8_takeCoverage: true");

// @api: v8.setHeapSnapshotNearHeapLimit
// @expect: v8_setHeapNearLimit: true
setHeapSnapshotNearHeapLimit(100);
console.log("v8_setHeapNearLimit: true");

// @api: v8.isStringOneByteRepresentation
// @expect: v8_isOneByte: true
console.log("v8_isOneByte: " + isStringOneByteRepresentation("abc"));

// @api: v8.v8.GCProfiler
// @expect: v8_gcprof_inst: true
const gcProf = new GCProfiler();
console.log("v8_gcprof_inst: " + (gcProf instanceof GCProfiler));

// @api: v8.GCProfiler.start
// @api: v8.GCProfiler.stop
// @expect: v8_gcprof_start_stop: 1
gcProf.start();
console.log("v8_gcprof_start_stop: " + gcProf.stop().version);

// @api: v8.v8.Serializer
// @expect: v8_ser_inst: true
const ser = new Serializer();
console.log("v8_ser_inst: " + (ser instanceof Serializer));

// @api: v8.Serializer.writeHeader
// @expect: v8_ser_writeHeader: true
ser.writeHeader();
console.log("v8_ser_writeHeader: true");

// @api: v8.Serializer.writeValue
// @expect: v8_ser_writeValue: true
console.log("v8_ser_writeValue: " + ser.writeValue(123));

// @api: v8.Serializer.releaseBuffer
// @expect: v8_ser_releaseBuf: 16
console.log("v8_ser_releaseBuf: " + ser.releaseBuffer().length);

// @api: v8.Serializer.transferArrayBuffer
// @expect: v8_ser_transferArr: true
ser.transferArrayBuffer(1, new ArrayBuffer(8));
console.log("v8_ser_transferArr: true");

// @api: v8.Serializer.writeUint32
// @expect: v8_ser_writeU32: true
ser.writeUint32(1234);
console.log("v8_ser_writeU32: true");

// @api: v8.Serializer.writeUint64
// @expect: v8_ser_writeU64: true
ser.writeUint64(0, 1234);
console.log("v8_ser_writeU64: true");

// @api: v8.Serializer.writeDouble
// @expect: v8_ser_writeDouble: true
ser.writeDouble(3.14);
console.log("v8_ser_writeDouble: true");

// @api: v8.Serializer.writeRawBytes
// @expect: v8_ser_writeRaw: true
ser.writeRawBytes(new Uint8Array(4));
console.log("v8_ser_writeRaw: true");

// @api: v8.Serializer._writeHostObject
// @expect: v8_ser_writeHost: true
ser._writeHostObject({});
console.log("v8_ser_writeHost: true");

// @api: v8.Serializer._getDataCloneError
// @expect: v8_ser_cloneErr: err
console.log("v8_ser_cloneErr: " + ser._getDataCloneError("err").message);

// @api: v8.Serializer._getSharedArrayBufferId
// @expect: v8_ser_sharedId: 0
console.log("v8_ser_sharedId: " + ser._getSharedArrayBufferId({}));

// @api: v8.Serializer._setTreatArrayBufferViewsAsHostObjects
// @expect: v8_ser_treatViews: true
ser._setTreatArrayBufferViewsAsHostObjects(true);
console.log("v8_ser_treatViews: true");

// @api: v8.v8.Deserializer
// @expect: v8_deser_inst: true
const deser = new Deserializer(new Uint8Array(0));
console.log("v8_deser_inst: " + (deser instanceof Deserializer));

// @api: v8.Deserializer.readHeader
// @expect: v8_deser_readHeader: false
console.log("v8_deser_readHeader: " + deser.readHeader());

// @api: v8.Deserializer.readValue
// @expect: v8_deser_readValue: undefined
console.log("v8_deser_readValue: " + deser.readValue());

// @api: v8.Deserializer.transferArrayBuffer
// @expect: v8_deser_transferArr: true
deser.transferArrayBuffer(1, new ArrayBuffer(8));
console.log("v8_deser_transferArr: true");

// @api: v8.Deserializer.getWireFormatVersion
// @expect: v8_deser_wireVer: 13
console.log("v8_deser_wireVer: " + deser.getWireFormatVersion());

// @api: v8.Deserializer.readUint32
// @expect: v8_deser_readU32: 0
console.log("v8_deser_readU32: " + deser.readUint32());

// @api: v8.Deserializer.readUint64
// @expect: v8_deser_readU64: 2
console.log("v8_deser_readU64: " + deser.readUint64().length);

// @api: v8.Deserializer.readDouble
// @expect: v8_deser_readDouble: 0
console.log("v8_deser_readDouble: " + deser.readDouble());

// @api: v8.Deserializer.readRawBytes
// @expect: v8_deser_readRaw: 4
console.log("v8_deser_readRaw: " + deser.readRawBytes(4).length);

// @api: v8.Deserializer._readHostObject
// @expect: v8_deser_readHost: undefined
console.log("v8_deser_readHost: " + deser._readHostObject());

// @api: v8.v8.DefaultSerializer
// @expect: v8_defSer_inst: true
const defSer = new DefaultSerializer();
console.log("v8_defSer_inst: " + (defSer instanceof DefaultSerializer));

// @api: v8.v8.DefaultDeserializer
// @expect: v8_defDeser_inst: true
const defDeser = new DefaultDeserializer();
console.log("v8_defDeser_inst: " + (defDeser instanceof DefaultDeserializer));

// @api: v8.serialize
// @expect: v8_serialize: true
const serData = serialize({ a: 1 });
console.log("v8_serialize: " + (serData.length > 0));

// @api: v8.deserialize
// @expect: v8_deserialize: object
console.log("v8_deserialize: " + typeof deserialize(serData));

// @api: v8.onInit
// @expect: v8_onInit: true
onInit(() => {});
console.log("v8_onInit: true");

// @api: v8.onSettled
// @expect: v8_onSettled: true
onSettled(() => {});
console.log("v8_onSettled: true");

// @api: v8.onBefore
// @expect: v8_onBefore: true
onBefore(() => {});
console.log("v8_onBefore: true");

// @api: v8.onAfter
// @expect: v8_onAfter: true
onAfter(() => {});
console.log("v8_onAfter: true");

// @api: v8.createHook
// @expect: v8_createHook: true
const hook = createHook({});
console.log("v8_createHook: " + (typeof hook === "object"));

// @api: v8.init
// @expect: v8_init: true
init();
console.log("v8_init: true");

// @api: v8.before
// @expect: v8_before: true
before();
console.log("v8_before: true");

// @api: v8.after
// @expect: v8_after: true
after();
console.log("v8_after: true");

// @api: v8.settled
// @expect: v8_settled: true
settled();
console.log("v8_settled: true");

// @api: v8.addSerializeCallback
// @expect: v8_addSerCb: true
addSerializeCallback(() => {});
console.log("v8_addSerCb: true");

// @api: v8.addDeserializeCallback
// @expect: v8_addDeserCb: true
addDeserializeCallback(() => {});
console.log("v8_addDeserCb: true");

// @api: v8.setDeserializeMainFunction
// @expect: v8_setDeserMain: true
setDeserializeMainFunction(() => {});
console.log("v8_setDeserMain: true");

// @api: v8.isBuildingSnapshot
// @expect: v8_isBuildingSnap: false
console.log("v8_isBuildingSnap: " + isBuildingSnapshot());
