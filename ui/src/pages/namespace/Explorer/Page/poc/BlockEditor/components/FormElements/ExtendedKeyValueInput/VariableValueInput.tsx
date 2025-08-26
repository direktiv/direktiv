import Input from "~/design/Input";

type VariableValueInputProps = {
  value: string;
  onChange: (value: string) => void;
  onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
};

export const VariableValueInput = ({
  value,
  onChange,
  onKeyDown,
}: VariableValueInputProps) => (
  <Input
    value={value}
    onKeyDown={onKeyDown}
    onChange={(e) => onChange(e.target.value)}
  />
);
