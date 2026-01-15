import { BlockPathType } from "../Block";
import { BlockSuspenseBoundary } from "../Block/utils/SuspenseBoundary";
import { BlockType } from "../../schema/blocks";
import { LocalVariables } from "../primitives/Variable/VariableContext";
import { ReactElement } from "react";

export type BlockWrapperProps = {
  blockPath: BlockPathType;
  block: BlockType;
  children: (register?: (vars: LocalVariables) => void) => ReactElement;
};

export const LiveBlockWrapper = ({ children }: BlockWrapperProps) => (
  <BlockSuspenseBoundary>
    <div className="my-3">{children()}</div>
  </BlockSuspenseBoundary>
);
