import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { BooleanValueInput } from "./BooleanValueInput";
import { ExtendedKeyValueType } from "../../../../schema/primitives/extendedKeyValue";
import { NumberValueInput } from "./NumberValueInput";
import { StringValueInput } from "./StringValueInput";
import { VariableValueInput } from "./VariableValueInput";

type ValueInputProps = {
  value: ExtendedKeyValueType["value"];
  onChange: (value: ExtendedKeyValueType["value"]) => void;
  onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
};

export const ValueInput = ({ value, onChange, onKeyDown }: ValueInputProps) => (
  <div className="flex gap-2">
    <Select
      value={value.type}
      onValueChange={(newType: ExtendedKeyValueType["value"]["type"]) => {
        let newValue: ExtendedKeyValueType["value"];
        switch (newType) {
          case "string":
            newValue = { type: "string", value: "" };
            break;
          case "variable":
            newValue = { type: "variable", value: "" };
            break;
          case "boolean":
            newValue = { type: "boolean", value: false };
            break;
          case "number":
            newValue = { type: "number", value: 0 };
            break;
          default:
            newValue = { type: "string", value: "" };
        }
        onChange(newValue);
      }}
    >
      <SelectTrigger className="w-32">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="string">String</SelectItem>
        <SelectItem value="variable">Variable</SelectItem>
        <SelectItem value="boolean">Boolean</SelectItem>
        <SelectItem value="number">Number</SelectItem>
      </SelectContent>
    </Select>

    {/* Render value input based on type */}
    {value.type === "string" && (
      <StringValueInput
        value={value.value}
        onChange={(newValue) => onChange({ type: "string", value: newValue })}
        onKeyDown={onKeyDown}
      />
    )}

    {value.type === "variable" && (
      <VariableValueInput
        value={value.value}
        onChange={(newValue) => onChange({ type: "variable", value: newValue })}
        onKeyDown={onKeyDown}
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
        onKeyDown={onKeyDown}
      />
    )}
  </div>
);
