// @generated by protoc-gen-es v1.7.2 with parameter "target=ts"
// @generated from file flyteidl/admin/project_attributes.proto (package flyteidl.admin, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { MatchableResource, MatchingAttributes } from "./matchable_resource_pb.js";

/**
 * Defines a set of custom matching attributes at the project level.
 * For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
 *
 * @generated from message flyteidl.admin.ProjectAttributes
 */
export class ProjectAttributes extends Message<ProjectAttributes> {
  /**
   * Unique project id for which this set of attributes will be applied.
   *
   * @generated from field: string project = 1;
   */
  project = "";

  /**
   * @generated from field: flyteidl.admin.MatchingAttributes matching_attributes = 2;
   */
  matchingAttributes?: MatchingAttributes;

  /**
   * Optional, org key applied to the project.
   *
   * @generated from field: string org = 3;
   */
  org = "";

  constructor(data?: PartialMessage<ProjectAttributes>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ProjectAttributes";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "project", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "matching_attributes", kind: "message", T: MatchingAttributes },
    { no: 3, name: "org", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ProjectAttributes {
    return new ProjectAttributes().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ProjectAttributes {
    return new ProjectAttributes().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ProjectAttributes {
    return new ProjectAttributes().fromJsonString(jsonString, options);
  }

  static equals(a: ProjectAttributes | PlainMessage<ProjectAttributes> | undefined, b: ProjectAttributes | PlainMessage<ProjectAttributes> | undefined): boolean {
    return proto3.util.equals(ProjectAttributes, a, b);
  }
}

/**
 * Sets custom attributes for a project
 * For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
 *
 * @generated from message flyteidl.admin.ProjectAttributesUpdateRequest
 */
export class ProjectAttributesUpdateRequest extends Message<ProjectAttributesUpdateRequest> {
  /**
   * +required
   *
   * @generated from field: flyteidl.admin.ProjectAttributes attributes = 1;
   */
  attributes?: ProjectAttributes;

  constructor(data?: PartialMessage<ProjectAttributesUpdateRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ProjectAttributesUpdateRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "attributes", kind: "message", T: ProjectAttributes },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ProjectAttributesUpdateRequest {
    return new ProjectAttributesUpdateRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ProjectAttributesUpdateRequest {
    return new ProjectAttributesUpdateRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ProjectAttributesUpdateRequest {
    return new ProjectAttributesUpdateRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ProjectAttributesUpdateRequest | PlainMessage<ProjectAttributesUpdateRequest> | undefined, b: ProjectAttributesUpdateRequest | PlainMessage<ProjectAttributesUpdateRequest> | undefined): boolean {
    return proto3.util.equals(ProjectAttributesUpdateRequest, a, b);
  }
}

/**
 * Purposefully empty, may be populated in the future.
 *
 * @generated from message flyteidl.admin.ProjectAttributesUpdateResponse
 */
export class ProjectAttributesUpdateResponse extends Message<ProjectAttributesUpdateResponse> {
  constructor(data?: PartialMessage<ProjectAttributesUpdateResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ProjectAttributesUpdateResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ProjectAttributesUpdateResponse {
    return new ProjectAttributesUpdateResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ProjectAttributesUpdateResponse {
    return new ProjectAttributesUpdateResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ProjectAttributesUpdateResponse {
    return new ProjectAttributesUpdateResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ProjectAttributesUpdateResponse | PlainMessage<ProjectAttributesUpdateResponse> | undefined, b: ProjectAttributesUpdateResponse | PlainMessage<ProjectAttributesUpdateResponse> | undefined): boolean {
    return proto3.util.equals(ProjectAttributesUpdateResponse, a, b);
  }
}

/**
 * Request to get an individual project level attribute override.
 * For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
 *
 * @generated from message flyteidl.admin.ProjectAttributesGetRequest
 */
export class ProjectAttributesGetRequest extends Message<ProjectAttributesGetRequest> {
  /**
   * Unique project id which this set of attributes references.
   * +required
   *
   * @generated from field: string project = 1;
   */
  project = "";

  /**
   * Which type of matchable attributes to return.
   * +required
   *
   * @generated from field: flyteidl.admin.MatchableResource resource_type = 2;
   */
  resourceType = MatchableResource.TASK_RESOURCE;

  /**
   * Optional, org key applied to the project.
   *
   * @generated from field: string org = 3;
   */
  org = "";

  constructor(data?: PartialMessage<ProjectAttributesGetRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ProjectAttributesGetRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "project", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "resource_type", kind: "enum", T: proto3.getEnumType(MatchableResource) },
    { no: 3, name: "org", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ProjectAttributesGetRequest {
    return new ProjectAttributesGetRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ProjectAttributesGetRequest {
    return new ProjectAttributesGetRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ProjectAttributesGetRequest {
    return new ProjectAttributesGetRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ProjectAttributesGetRequest | PlainMessage<ProjectAttributesGetRequest> | undefined, b: ProjectAttributesGetRequest | PlainMessage<ProjectAttributesGetRequest> | undefined): boolean {
    return proto3.util.equals(ProjectAttributesGetRequest, a, b);
  }
}

/**
 * Response to get an individual project level attribute override.
 * For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
 *
 * @generated from message flyteidl.admin.ProjectAttributesGetResponse
 */
export class ProjectAttributesGetResponse extends Message<ProjectAttributesGetResponse> {
  /**
   * @generated from field: flyteidl.admin.ProjectAttributes attributes = 1;
   */
  attributes?: ProjectAttributes;

  constructor(data?: PartialMessage<ProjectAttributesGetResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ProjectAttributesGetResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "attributes", kind: "message", T: ProjectAttributes },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ProjectAttributesGetResponse {
    return new ProjectAttributesGetResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ProjectAttributesGetResponse {
    return new ProjectAttributesGetResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ProjectAttributesGetResponse {
    return new ProjectAttributesGetResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ProjectAttributesGetResponse | PlainMessage<ProjectAttributesGetResponse> | undefined, b: ProjectAttributesGetResponse | PlainMessage<ProjectAttributesGetResponse> | undefined): boolean {
    return proto3.util.equals(ProjectAttributesGetResponse, a, b);
  }
}

/**
 * Request to delete a set matchable project level attribute override.
 * For more info on matchable attributes, see :ref:`ref_flyteidl.admin.MatchableAttributesConfiguration`
 *
 * @generated from message flyteidl.admin.ProjectAttributesDeleteRequest
 */
export class ProjectAttributesDeleteRequest extends Message<ProjectAttributesDeleteRequest> {
  /**
   * Unique project id which this set of attributes references.
   * +required
   *
   * @generated from field: string project = 1;
   */
  project = "";

  /**
   * Which type of matchable attributes to delete.
   * +required
   *
   * @generated from field: flyteidl.admin.MatchableResource resource_type = 2;
   */
  resourceType = MatchableResource.TASK_RESOURCE;

  /**
   * Optional, org key applied to the project.
   *
   * @generated from field: string org = 3;
   */
  org = "";

  constructor(data?: PartialMessage<ProjectAttributesDeleteRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ProjectAttributesDeleteRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "project", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "resource_type", kind: "enum", T: proto3.getEnumType(MatchableResource) },
    { no: 3, name: "org", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ProjectAttributesDeleteRequest {
    return new ProjectAttributesDeleteRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ProjectAttributesDeleteRequest {
    return new ProjectAttributesDeleteRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ProjectAttributesDeleteRequest {
    return new ProjectAttributesDeleteRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ProjectAttributesDeleteRequest | PlainMessage<ProjectAttributesDeleteRequest> | undefined, b: ProjectAttributesDeleteRequest | PlainMessage<ProjectAttributesDeleteRequest> | undefined): boolean {
    return proto3.util.equals(ProjectAttributesDeleteRequest, a, b);
  }
}

/**
 * Purposefully empty, may be populated in the future.
 *
 * @generated from message flyteidl.admin.ProjectAttributesDeleteResponse
 */
export class ProjectAttributesDeleteResponse extends Message<ProjectAttributesDeleteResponse> {
  constructor(data?: PartialMessage<ProjectAttributesDeleteResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "flyteidl.admin.ProjectAttributesDeleteResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ProjectAttributesDeleteResponse {
    return new ProjectAttributesDeleteResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ProjectAttributesDeleteResponse {
    return new ProjectAttributesDeleteResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ProjectAttributesDeleteResponse {
    return new ProjectAttributesDeleteResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ProjectAttributesDeleteResponse | PlainMessage<ProjectAttributesDeleteResponse> | undefined, b: ProjectAttributesDeleteResponse | PlainMessage<ProjectAttributesDeleteResponse> | undefined): boolean {
    return proto3.util.equals(ProjectAttributesDeleteResponse, a, b);
  }
}

