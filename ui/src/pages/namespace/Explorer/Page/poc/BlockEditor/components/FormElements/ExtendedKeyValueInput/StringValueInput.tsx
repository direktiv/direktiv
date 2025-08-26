import Input from "~/design/Input";

type StringValueInputProps = {
  value: string;
  onChange: (value: string) => void;
  onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
};

export const StringValueInput = ({
  value,
  onChange,
  onKeyDown,
}: StringValueInputProps) => (
  <Input
    value={value}
    onKeyDown={onKeyDown}
    onChange={(e) => onChange(e.target.value)}
  />
);
