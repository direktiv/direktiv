import { FC } from "react";

type TreeElementProps = { id: string; label: string; depth?: number };

const TreeElement: FC<TreeElementProps> = ({ label, depth = 0 }) => (
  <div className="w-full" style={{ paddingLeft: `${depth * 12}px` }}>
    {label}
  </div>
);

export default TreeElement;
