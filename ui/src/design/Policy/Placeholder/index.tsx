import React, { FC } from "react";

import { Plus } from "lucide-react";
import { Separator } from "~/design/Separator";

const ConditionIcon: FC<React.HTMLAttributes<HTMLOrSVGElement>> = (props) => (
  <svg
    {...props}
    viewBox="0 0 48 48"
    preserveAspectRatio="xMinYMin"
    className="h-[48px] stroke-gray-400 stroke-2 hover:cursor-pointer hover:stroke-gray-500"
    aria-label="AND-group icon"
  >
    <line x1="0" y1="24" x2="12" y2="24" />
    <line x1="36" y1="24" x2="48" y2="24" />
    <rect x="12" y="12" width="24" height="24" rx="4" fill="none" />
  </svg>
);

const OrGroupIcon: FC<React.HTMLAttributes<HTMLOrSVGElement>> = (props) => (
  <svg
    {...props}
    viewBox="0 0 48 48"
    preserveAspectRatio="xMinYMin"
    className="h-[48px] stroke-gray-400 stroke-2 hover:cursor-pointer hover:stroke-gray-500"
    fill="none"
    aria-label="OR-group icon"
  >
    <rect width="16" height="16" x="16" y="28" rx="4" />
    <rect width="16" height="16" x="16" y="5" rx="4" fill="none" />
    <path d="M0,24H 4A 4 4 0 0 0 8 20V 16A 4 4 0 0 1 10 12H 16" />
    <path d="M0,24H 4A 4 4 0 0 1 8 28V 32A 4 4 0 0 0 10 36H 16" />
    <path d="M48,24H 44A 4 4 0 0 1 40 20V 16A 4 4 0 0 0 38 12H 32" />
    <path d="M48,24H 44A 4 4 0 0 0 40 28V 32A 4 4 0 0 1 38 36H 32" />
  </svg>
);

type PlaceholderProps = {
  addCondition?: () => void;
  addOrGroup?: () => void;
};

const Placeholder: FC<PlaceholderProps> = ({ addCondition, addOrGroup }) => (
  <div>
    <div
      className="group my-[16px] flex h-[64px] w-[160px] flex-col items-center justify-center rounded-[8px] border-2 border-dotted border-gray-400"
      aria-label="placeholder-condition"
    >
      <div className="group-hover:hidden">
        <Plus className="text-gray-500" />
      </div>
      <div className="hidden flex-row group-hover:flex">
        <ConditionIcon onClick={addCondition} />
        <Separator vertical className="mx-2" />
        <OrGroupIcon onClick={addOrGroup} />
      </div>
    </div>
  </div>
);

export { Placeholder };
