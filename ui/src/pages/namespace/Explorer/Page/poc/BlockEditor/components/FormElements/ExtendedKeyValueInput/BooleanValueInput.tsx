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

export const BooleanValueInput = ({
  value,
  onChange,
}: BooleanValueInputProps) => (
  <Select
    value={value.toString()}
    onValueChange={(val: "true" | "false") => {
      onChange(val === "true");
    }}
  >
    <SelectTrigger className="flex-1">
      <SelectValue />
    </SelectTrigger>
    <SelectContent>
      <SelectItem value="true">true</SelectItem>
      <SelectItem value="false">false</SelectItem>
    </SelectContent>
  </Select>
);
