import { Calendar } from "lucide-react";

const NoResult = ({ message }: { message: string }) => (
  <div
    className="flex flex-col items-center justify-center gap-1 p-10"
    data-testid="listener-no-result"
  >
    <Calendar />
    <span className="text-center text-sm">{message}</span>
  </div>
);

export default NoResult;
