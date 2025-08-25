import { JsonValueType } from "../Variable/utils";
import { LocalVariablesContent } from "../Variable/LocalVariables";

export type KeyValueResolverFunction<InputType> = (
  value: InputType,
  localVariables?: LocalVariablesContent
) => KeyValue[];

const keyValueToObject = (kv: KeyValue) => ({
  [kv.key]: kv.value,
});

export const keyValueArrayToObject = (kv: KeyValue[]) =>
  kv.reduce((acc, curr) => ({ ...acc, ...keyValueToObject(curr) }), {});

export type KeyValue = {
  key: string;
  value: JsonValueType;
};
