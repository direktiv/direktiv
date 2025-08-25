import Input from "~/design/Input";

type VariableValueInputProps = {
  value: string;
  onChange: (value: string) => void;
  onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
  placeholder?: string;
};

export const VariableValueInput = ({
  value,
  onChange,
  onKeyDown,
  placeholder = "Variable name",
}: VariableValueInputProps) => (
  <Input
    placeholder={placeholder}
    value={value}
    onKeyDown={onKeyDown}
    onChange={(e) => onChange(e.target.value)}
    className="flex-1"
  />
);
