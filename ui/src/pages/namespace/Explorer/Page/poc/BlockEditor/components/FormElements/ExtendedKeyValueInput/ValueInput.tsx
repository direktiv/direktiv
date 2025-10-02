import {
  ExtendedKeyValueType,
  ValueType,
} from "../../../../schema/primitives/extendedKeyValue";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { BooleanValueInput } from "./BooleanValueInput";
import { NumberValueInput } from "./NumberValueInput";
import { StringValueInput } from "./StringValueInput";
import { VariableValueInput } from "./VariableValueInput";

type ValueInputProps = {
  value: ExtendedKeyValueType["value"];
  onChange: (value: ExtendedKeyValueType["value"]) => void;
  smart?: boolean;
};

const values: ValueType[] = ["string", "variable", "boolean", "number"];

export const ValueInput = ({
  value,
  onChange,
  smart = false,
}: ValueInputProps) => (
  <>
    <Select
      value={value.type}
      onValueChange={(newType: ValueType) => {
        let newValue: ExtendedKeyValueType["value"];
        switch (newType) {
          case "string":
            newValue = { type: "string", value: "" };
            break;
          case "variable":
            newValue = { type: "variable", value: "" };
            break;
          case "boolean":
            newValue = { type: "boolean", value: true };
            break;
          case "number":
            newValue = { type: "number", value: 0 };
            break;
        }
        onChange(newValue);
      }}
    >
      <SelectTrigger variant="outline">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {values.map((valueType) => (
          <SelectItem key={valueType} value={valueType}>
            {valueType}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>

    {value.type === "string" && (
      <StringValueInput
        smart={smart}
        value={value.value}
        onChange={(newValue) => onChange({ type: "string", value: newValue })}
      />
    )}

    {value.type === "variable" && (
      <VariableValueInput
        value={value.value}
        onChange={(newValue) => onChange({ type: "variable", value: newValue })}
      />
    )}

    {value.type === "boolean" && (
      <BooleanValueInput
        value={value.value}
        onChange={(newValue) => onChange({ type: "boolean", value: newValue })}
      />
    )}

    {value.type === "number" && (
      <NumberValueInput
        value={value.value}
        onChange={(newValue) => onChange({ type: "number", value: newValue })}
      />
    )}
  </>
);
