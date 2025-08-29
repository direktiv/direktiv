import Input from "~/design/Input";

type NumberValueInputProps = {
  value: number;
  onChange: (value: number) => void;
  onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
};

export const NumberValueInput = ({
  value,
  onChange,
  onKeyDown,
}: NumberValueInputProps) => (
  <Input
    type="number"
    value={String(value)}
    onKeyDown={onKeyDown}
    onChange={(e) => {
      const numValue = parseFloat(e.target.value) || 0;
      onChange(numValue);
    }}
  />
);
