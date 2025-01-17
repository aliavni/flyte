// @generated by protoc-gen-es v1.7.2 with parameter "target=ts"
// @generated from file flyteidl/plugins/qubole.proto (package flyteidl.plugins, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * Defines a query to execute on a hive cluster.
 *
 * @generated from message flyteidl.plugins.HiveQuery
 */
export class HiveQuery extends Message<HiveQuery> {
  /**
   * @generated from field: string query = 1;
   */
  query = "";

  /**
   * @generated from field: uint32 timeout_sec = 2;
   */
  timeoutSec = 0;

  /**
   * @generated from field: uint32 retryCount = 3;
   */
  retryCount = 0;

  constructor(data?: PartialMessage<HiveQuery>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.plugins.HiveQuery";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "query", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "timeout_sec", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 3, name: "retryCount", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): HiveQuery {
    return new HiveQuery().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): HiveQuery {
    return new HiveQuery().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): HiveQuery {
    return new HiveQuery().fromJsonString(jsonString, options);
  }

  static equals(a: HiveQuery | PlainMessage<HiveQuery> | undefined, b: HiveQuery | PlainMessage<HiveQuery> | undefined): boolean {
    return proto3.util.equals(HiveQuery, a, b);
  }
}

/**
 * Defines a collection of hive queries.
 *
 * @generated from message flyteidl.plugins.HiveQueryCollection
 */
export class HiveQueryCollection extends Message<HiveQueryCollection> {
  /**
   * @generated from field: repeated flyteidl.plugins.HiveQuery queries = 2;
   */
  queries: HiveQuery[] = [];

  constructor(data?: PartialMessage<HiveQueryCollection>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.plugins.HiveQueryCollection";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 2, name: "queries", kind: "message", T: HiveQuery, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): HiveQueryCollection {
    return new HiveQueryCollection().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): HiveQueryCollection {
    return new HiveQueryCollection().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): HiveQueryCollection {
    return new HiveQueryCollection().fromJsonString(jsonString, options);
  }

  static equals(a: HiveQueryCollection | PlainMessage<HiveQueryCollection> | undefined, b: HiveQueryCollection | PlainMessage<HiveQueryCollection> | undefined): boolean {
    return proto3.util.equals(HiveQueryCollection, a, b);
  }
}

/**
 * This message works with the 'hive' task type in the SDK and is the object that will be in the 'custom' field
 * of a hive task's TaskTemplate
 *
 * @generated from message flyteidl.plugins.QuboleHiveJob
 */
export class QuboleHiveJob extends Message<QuboleHiveJob> {
  /**
   * @generated from field: string cluster_label = 1;
   */
  clusterLabel = "";

  /**
   * @generated from field: flyteidl.plugins.HiveQueryCollection query_collection = 2 [deprecated = true];
   * @deprecated
   */
  queryCollection?: HiveQueryCollection;

  /**
   * @generated from field: repeated string tags = 3;
   */
  tags: string[] = [];

  /**
   * @generated from field: flyteidl.plugins.HiveQuery query = 4;
   */
  query?: HiveQuery;

  constructor(data?: PartialMessage<QuboleHiveJob>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.plugins.QuboleHiveJob";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "cluster_label", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "query_collection", kind: "message", T: HiveQueryCollection },
    { no: 3, name: "tags", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 4, name: "query", kind: "message", T: HiveQuery },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): QuboleHiveJob {
    return new QuboleHiveJob().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): QuboleHiveJob {
    return new QuboleHiveJob().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): QuboleHiveJob {
    return new QuboleHiveJob().fromJsonString(jsonString, options);
  }

  static equals(a: QuboleHiveJob | PlainMessage<QuboleHiveJob> | undefined, b: QuboleHiveJob | PlainMessage<QuboleHiveJob> | undefined): boolean {
    return proto3.util.equals(QuboleHiveJob, a, b);
  }
}

