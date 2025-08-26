import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

type BooleanValueInputProps = {
  value: boolean;
  onChange: (value: boolean) => void;
};

const booleanValues = ["true", "false"] as const;

export const BooleanValueInput = ({
  value,
  onChange,
}: BooleanValueInputProps) => (
  <Select
    value={value.toString()}
    onValueChange={(val: (typeof booleanValues)[number]) => {
      onChange(val === "true");
    }}
  >
    <SelectTrigger variant="outline">
      <SelectValue />
    </SelectTrigger>
    <SelectContent>
      {booleanValues.map((value) => (
        <SelectItem key={value} value={value}>
          {value}
        </SelectItem>
      ))}
    </SelectContent>
  </Select>
);
