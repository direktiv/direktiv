import { FC } from "react";
import { FilterIcon } from "lucide-react";

type TreeElementProps = {
  id: string;
  label: string;
  depth?: number;
  onFilter?: () => void;
};

const TreeElement: FC<TreeElementProps> = ({ label, depth = 0, onFilter }) => (
  <div
    className="w-full flex flex-row justify-between"
    style={{ paddingLeft: `${depth * 12}px` }}
  >
    <span>{label}</span>
    {onFilter && <FilterIcon className="h-4" onClick={onFilter} />}
  </div>
);

export default TreeElement;
