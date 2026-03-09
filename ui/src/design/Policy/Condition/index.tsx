import { Card } from "~/design/Card";

type ConditionComponentProps = {
  label: string;
  value: string;
  operator: string;
  invert?: boolean;
};

const Condition = ({
  label,
  operator,
  value,
  invert,
}: ConditionComponentProps) => (
  <Card
    background="weight-2"
    className="flex min-w-40 flex-col items-center justify-end py-2"
    aria-label="condition"
  >
    <div className="self-center text-xs">{label}</div>
    <div className="text-xs">
      {invert === true && <>!</>}
      {operator}
    </div>
    {Array.isArray(value) ? (
      value.map((v: string, idx: number) => (
        <div key={idx} className="max-w-[40px] truncate text-center text-xs">
          {v}
        </div>
      ))
    ) : (
      <div className="max-w-80 truncate text-center text-xs">{value}</div>
    )}
  </Card>
);

export { Condition };
