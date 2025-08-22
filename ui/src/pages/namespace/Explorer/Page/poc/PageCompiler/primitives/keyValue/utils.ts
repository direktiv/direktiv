import { JsonValueType } from "../Variable/utils";
import { LocalVariablesContent } from "../Variable/LocalVariables";

export type KeyValueResolverFunction<InputType> = (
  value: InputType,
  localVariables?: LocalVariablesContent
) => ExtendedKeyValueType[];

const keyValueToObject = (kv: ExtendedKeyValueType) => ({
  [kv.key]: kv.value,
});

export const keyValueArrayToObject = (kv: ExtendedKeyValueType[]) =>
  kv.reduce((acc, curr) => ({ ...acc, ...keyValueToObject(curr) }), {});

export type ExtendedKeyValueType = {
  key: string;
  value: JsonValueType;
};
