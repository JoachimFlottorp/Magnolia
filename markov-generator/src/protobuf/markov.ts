/* eslint-disable */
import * as _m0 from "protobufjs/minimal";

export const protobufPackage = "proto";

export interface MarkovRequest {
  messages: string[];
  seed?: string | undefined;
}

export interface MarkovResponse {
  result: string;
  error?: string | undefined;
}

function createBaseMarkovRequest(): MarkovRequest {
  return { messages: [], seed: undefined };
}

export const MarkovRequest = {
  encode(message: MarkovRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.messages) {
      writer.uint32(10).string(v!);
    }
    if (message.seed !== undefined) {
      writer.uint32(18).string(message.seed);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): MarkovRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMarkovRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.messages.push(reader.string());
          break;
        case 2:
          message.seed = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MarkovRequest {
    return {
      messages: Array.isArray(object?.messages) ? object.messages.map((e: any) => String(e)) : [],
      seed: isSet(object.seed) ? String(object.seed) : undefined,
    };
  },

  toJSON(message: MarkovRequest): unknown {
    const obj: any = {};
    if (message.messages) {
      obj.messages = message.messages.map((e) => e);
    } else {
      obj.messages = [];
    }
    message.seed !== undefined && (obj.seed = message.seed);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<MarkovRequest>, I>>(object: I): MarkovRequest {
    const message = createBaseMarkovRequest();
    message.messages = object.messages?.map((e) => e) || [];
    message.seed = object.seed ?? undefined;
    return message;
  },
};

function createBaseMarkovResponse(): MarkovResponse {
  return { result: "", error: undefined };
}

export const MarkovResponse = {
  encode(message: MarkovResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.result !== "") {
      writer.uint32(10).string(message.result);
    }
    if (message.error !== undefined) {
      writer.uint32(18).string(message.error);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): MarkovResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMarkovResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.result = reader.string();
          break;
        case 2:
          message.error = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MarkovResponse {
    return {
      result: isSet(object.result) ? String(object.result) : "",
      error: isSet(object.error) ? String(object.error) : undefined,
    };
  },

  toJSON(message: MarkovResponse): unknown {
    const obj: any = {};
    message.result !== undefined && (obj.result = message.result);
    message.error !== undefined && (obj.error = message.error);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<MarkovResponse>, I>>(object: I): MarkovResponse {
    const message = createBaseMarkovResponse();
    message.result = object.result ?? "";
    message.error = object.error ?? undefined;
    return message;
  },
};

type Builtin = Date | Function | Uint8Array | string | number | boolean | undefined;

export type DeepPartial<T> = T extends Builtin ? T
  : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>>
  : T extends {} ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

type KeysOfUnion<T> = T extends T ? keyof T : never;
export type Exact<P, I extends P> = P extends Builtin ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & { [K in Exclude<keyof I, KeysOfUnion<P>>]: never };

function isSet(value: any): boolean {
  return value !== null && value !== undefined;
}
