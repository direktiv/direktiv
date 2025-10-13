import Input from "~/design/Input";

type NumberValueInputProps = {
  value: number;
  onChange: (value: number) => void;
};

export const NumberValueInput = ({
  value,
  onChange,
}: NumberValueInputProps) => (
  <Input
    type="number"
    value={String(value)}
    onChange={(e) => {
      const numValue = parseFloat(e.target.value) || 0;
      onChange(numValue);
    }}
  />
);
